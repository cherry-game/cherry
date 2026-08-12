// Package cherryTimeWheel implements a 5-level hierarchical timer wheel.
//
// Architecture:
//   - 512 slots (256 near + 4×64 levels)
//   - Single driver goroutine, lock-free
//   - MPSC submission via timerCmd interface
//   - sync.Pool for timerNode reuse
//   - Dispatch runs synchronously in the driver goroutine and only fires timers;
//     callbacks must be lightweight (business logic is pushed to actor queues by
//     the upper layer, not executed inside the wheel)
//
// Concurrency:
//   - Start/Stop are lifecycle methods and MUST be called serially from a single
//     goroutine (the caller serializes them); they are not synchronized internally.
//   - submitAdd/submitRemove are safe to call from any number of goroutines concurrently.
//   - A *Timer handle is NOT goroutine-safe; drive it from the single goroutine
//     that created it.
//   - Timer.Stop is asynchronous: the removal is queued and the callback may fire
//     once more after Stop returns.
//   - timerNode / nodeMap / wheel slots are owned by the driver goroutine and are
//     only touched there.
//
// Cascade (bitwise, branch-free):
//
//	near (256t):  current & 0xFF == 0       → cascade(0)
//	t[0] (16Kt):  current & 0x3FFF == 0     → cascade(1)
//	t[1] (1Mt):   current & 0xFFFFF == 0    → cascade(2)
//	t[2] (67Mt):  current & 0x3FFFFFF == 0  → cascade(3)
//
// Limitations:
//   - Fixed 5 levels cap the maximum expressible delay at 2^32 ticks
//     (~497 days at the 10ms default tick, ~49 days at 1ms). Longer delays wrap
//     to a wrong slot and fire early, so keep delays within this ceiling.
//   - The wheel advances on a fixed ticker instead of sleeping until the next
//     expiry, so it burns a little CPU even when idle (negligible at 10ms).
//   - submitAdd/submitRemove use a bounded MPSC buffer; Add/Stop block under backpressure,
//     which only happens if the driver goroutine stalls.
//   - Stop is terminal: once stopped, the wheel is closed and cannot be restarted
//     (Start becomes a no-op); create a new wheel instead.
package cherryTimeWheel

import (
	"sync"
	"sync/atomic"
	"time"

	cutils "github.com/cherry-game/cherry/extend/utils"
	clog "github.com/cherry-game/cherry/logger"
)

// 5-level timer wheel geometry
const (
	NEAR_SHIFT  = 8                     // 2^8 = 256 slots
	NEAR_SIZE   = 1 << NEAR_SHIFT       // 256
	LEVEL_SHIFT = 6                     // 2^6 = 64 slots per level
	LEVEL_SIZE  = 1 << LEVEL_SHIFT      // 64
	NEAR_MASK   = NEAR_SIZE - 1         // 0xFF
	LEVEL_MASK  = LEVEL_SIZE - 1        // 0x3F
	SUBMIT_CAP  = 4096                  // MPSC submit channel buffer
	DefaultTick = 10 * time.Millisecond // minimum tick
)

// TimeWheel is a 5-level hierarchical timer wheel.
// Wheel advance is serialized in a single driver goroutine; command submission
// (submitAdd/submitRemove) is concurrent from many goroutines.
type TimeWheel struct {
	near [NEAR_SIZE]*timerNode     // level 0: 256 slots, 1-tick precision
	t    [4][LEVEL_SIZE]*timerNode // levels 1~4: 64 slots each

	current   int64                     // monotonic tick counter
	startTime atomic.Pointer[time.Time] // monotonic clock base, written by Start only

	submitCh chan timerCmd // MPSC submit channel (single channel keeps FIFO-ordered)
	exitCh   chan struct{}

	nodePool  sync.Pool             // *timerNode pool
	nodeMap   map[uint64]*timerNode // id → *timerNode (driver goroutine only)
	activeNum atomic.Int64          // active timer count

	tickDur time.Duration // tick interval as Duration
	wg      sync.WaitGroup

	stopped atomic.Int32 // 1 = stopped (terminal); submit drops commands when set
}

// NewTimeWheel creates a timer wheel instance.
func NewTimeWheel(tick time.Duration) *TimeWheel {
	return NewTimeWheelWithHint(tick, 0)
}

// NewTimeWheelWithHint creates a timer wheel instance with a pre-allocated
// nodeMap capacity hint (recommended for million-scale workloads).
func NewTimeWheelWithHint(tick time.Duration, hint int) *TimeWheel {
	if tick < time.Millisecond {
		clog.Warnf("[NewTimeWheel] tick=%v < 1ms, fallback to %v", tick, DefaultTick)
		tick = DefaultTick
	}
	tw := &TimeWheel{
		submitCh: make(chan timerCmd, SUBMIT_CAP),
		nodeMap:  make(map[uint64]*timerNode, hint),
		nodePool: sync.Pool{
			New: func() any { return &timerNode{} },
		},
		tickDur: tick,
	}
	return tw
}

// Start begins the driver goroutine. Idempotent. Must not be called concurrently
// with Stop (lifecycle is serialized by the caller).
func (tw *TimeWheel) Start() {
	if tw.stopped.Load() == 1 || tw.exitCh != nil {
		return
	}
	tw.exitCh = make(chan struct{})
	st := time.Now()
	tw.startTime.Store(&st)

	tw.wg.Add(1)
	go func() {
		defer tw.wg.Done()
		ticker := time.NewTicker(tw.tickDur)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tw.tick()
			case cmd := <-tw.submitCh:
				cmd.exec(tw)
				tw.drainCmds()
			case <-tw.exitCh:
				return
			}
		}
	}()
}

// Stop shuts down the driver goroutine. Idempotent and terminal: once stopped the
// wheel is closed and cannot be restarted. Must not be called concurrently with
// Start (lifecycle is serialized by the caller).
func (tw *TimeWheel) Stop() {
	if tw.stopped.Load() == 1 || tw.exitCh == nil {
		return
	}
	tw.stopped.Store(1)
	close(tw.exitCh)
	tw.wg.Wait()
	tw.exitCh = nil
}

// ActiveCount returns the number of active timers.
func (tw *TimeWheel) ActiveCount() int64 {
	return tw.activeNum.Load()
}

// ---- MPSC ----

// submitAdd enqueues a node to be inserted by the driver goroutine. Thread-safe.
func (tw *TimeWheel) submitAdd(node *timerNode) {
	if tw.stopped.Load() == 1 {
		return
	}
	tw.submitCh <- &addCmd{node: node}
}

// submitRemove enqueues a timer id to be removed by the driver goroutine. Thread-safe.
func (tw *TimeWheel) submitRemove(id uint64) {
	if tw.stopped.Load() == 1 {
		return
	}
	tw.submitCh <- &removeCmd{id: id}
}

// drainCmds processes any backlogged commands (driver goroutine only).
func (tw *TimeWheel) drainCmds() {
	for {
		select {
		case cmd := <-tw.submitCh:
			cmd.exec(tw)
		default:
			return
		}
	}
}

// ---- Internal (driver goroutine only) ----

// recycle removes node from nodeMap (if it is still the current entry), clears
// its callback and schedule, and returns it to the pool. Caller owns the
// activeNum adjustment.
func (tw *TimeWheel) recycle(node *timerNode) {
	if n, ok := tw.nodeMap[node.id]; ok && n == node {
		delete(tw.nodeMap, node.id)
	}
	node.cb = nil
	node.interval = 0
	node.schedule = nil
	tw.nodePool.Put(node)
}

func (tw *TimeWheel) handleAddNode(node *timerNode) {
	// Defensive: replace a stale node with the same id if one is still present.
	// Only happens if the single-goroutine Timer contract is violated (e.g. a
	// concurrent Start); detach and recycle it so it isn't leaked.
	if old, ok := tw.nodeMap[node.id]; ok {
		tw.detachNode(old)
		tw.recycle(old)
		tw.activeNum.Add(-1)
	}
	node.next = nil
	node.prev = nil
	tw.enqueue(node)
	tw.nodeMap[node.id] = node
	tw.activeNum.Add(1)
}

// detachNode removes a node from its linked list in O(1).
func (tw *TimeWheel) detachNode(node *timerNode) {
	var head **timerNode
	if node.level < 0 {
		head = &tw.near[node.index]
	} else {
		head = &tw.t[node.level][node.index]
	}
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		*head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	node.next = nil
	node.prev = nil
}

// ---- Wheel advance & cascade ----

func (tw *TimeWheel) tick() {
	target := tw.elapsedTicks()
	for {
		cur := tw.current
		if cur > target {
			break
		}
		idx := cur & NEAR_MASK
		if idx == 0 {
			tw.cascade(0, cur)
		}
		tw.dispatchList(detachList(&tw.near[int(idx)]))
		tw.current++
	}
}

// cascade moves timers from an upper level to lower levels.
func (tw *TimeWheel) cascade(level int, current int64) {
	if level >= 4 {
		return
	}
	shift := NEAR_SHIFT + level*LEVEL_SHIFT
	idx := int((current >> shift) & LEVEL_MASK)
	head := detachList(&tw.t[level][idx])
	for node := head; node != nil; {
		next := node.next
		if node.cb != nil {
			tw.enqueue(node)
		} else {
			// Defensive: a cancelled node (cb == nil) left in a slot is recycled
			// consistently with dispatchList.
			tw.recycle(node)
			tw.activeNum.Add(-1)
		}
		node = next
	}
	if idx == 0 {
		tw.cascade(level+1, current)
	}
}

// enqueue inserts a node into the correct slot. Pure bitwise, no modulo.
func (tw *TimeWheel) enqueue(node *timerNode) {
	if node.expire <= tw.current {
		node.expire = tw.current + 1
	}
	delta := node.expire - tw.current
	switch {
	case delta < NEAR_SIZE:
		node.level = -1
		node.index = int32(node.expire & NEAR_MASK)
		listInsert(&tw.near[node.index], node)
	case delta < (1 << (NEAR_SHIFT + LEVEL_SHIFT)):
		node.level = 0
		node.index = int32((node.expire >> NEAR_SHIFT) & LEVEL_MASK)
		listInsert(&tw.t[0][node.index], node)
	case delta < (1 << (NEAR_SHIFT + 2*LEVEL_SHIFT)):
		node.level = 1
		node.index = int32((node.expire >> (NEAR_SHIFT + LEVEL_SHIFT)) & LEVEL_MASK)
		listInsert(&tw.t[1][node.index], node)
	case delta < (1 << (NEAR_SHIFT + 3*LEVEL_SHIFT)):
		node.level = 2
		node.index = int32((node.expire >> (NEAR_SHIFT + 2*LEVEL_SHIFT)) & LEVEL_MASK)
		listInsert(&tw.t[2][node.index], node)
	default:
		node.level = 3
		node.index = int32((node.expire >> (NEAR_SHIFT + 3*LEVEL_SHIFT)) & LEVEL_MASK)
		listInsert(&tw.t[3][node.index], node)
	}
}

// ---- Expiry & dispatch ----

// dispatchList processes expired timers: runs the callback, then either
// reschedules (schedule / recurring) or recycles (one-shot).
func (tw *TimeWheel) dispatchList(head *timerNode) {
	for node := head; node != nil; {
		next := node.next
		if n, ok := tw.nodeMap[node.id]; ok && n == node {
			delete(tw.nodeMap, node.id)
		}

		cutils.Try(func() {
			node.cb()
		}, func(errString string) {
			clog.Warnf("[time_wheel] dispatch panic. id=%d, err=%s", node.id, errString)
		})

		switch {
		case node.schedule != nil:
			// AddSchedule: reschedule by the scheduler until it returns zero time.
			if nextExp := node.schedule.Next(tw.ticksToTime(tw.nowTicks())); !nextExp.IsZero() {
				node.expire = tw.timeToTicks(nextExp)
				tw.handleAddNode(node)
			} else {
				tw.recycle(node)
			}
		case node.interval > 0:
			// Recurring: reschedule at a fixed interval.
			ticks := max(tw.durationToTicks(node.interval), 1)
			node.expire = tw.nowTicks() + ticks
			tw.handleAddNode(node)
		default:
			// One-shot: recycle.
			tw.recycle(node)
		}

		tw.activeNum.Add(-1)
		node = next
	}
}

// ---- Public API ----

// Add creates a recurring timer and starts it.
func (tw *TimeWheel) Add(d time.Duration, f func()) *Timer {
	return tw.addTimer(d, f, true)
}

// AddOnce creates a one-shot timer.
func (tw *TimeWheel) AddOnce(d time.Duration, f func()) *Timer {
	return tw.addTimer(d, f, false)
}

func (tw *TimeWheel) addTimer(d time.Duration, f func(), recurring bool) *Timer {
	t := &Timer{
		id: nextID(),
		tw: tw,
		d:  d,
		f:  f,
	}
	tw.startNode(t, recurring)
	return t
}

// startNode binds a node to the timer and submits it to the wheel. The
// recurring flag marks the node as interval-driven (one-shot keeps interval
// zero). The business callback is stored on the node so no per-timer closure
// is allocated.
func (tw *TimeWheel) startNode(t *Timer, recurring bool) {
	ticks := tw.durationToTicks(t.d)
	ticks = max(ticks, 1)
	node := tw.nodePool.Get().(*timerNode)
	node.id = t.id
	node.expire = tw.nowTicks() + ticks
	node.cb = t.f
	node.interval = 0
	if recurring {
		node.interval = t.d
	}
	t.node = node
	tw.submitAdd(node)
}

// AddSchedule creates a timer driven by a Scheduler.
func (tw *TimeWheel) AddSchedule(s Scheduler, f func()) *Timer {
	firstExp := s.Next(time.Now())
	if firstExp.IsZero() {
		return nil
	}

	t := &Timer{
		id: nextID(),
		tw: tw,
		f:  f,
	}

	node := tw.nodePool.Get().(*timerNode)
	node.id = t.id
	node.expire = tw.timeToTicks(firstExp)
	node.cb = f
	node.schedule = s
	t.node = node
	tw.submitAdd(node)
	return t
}

// ---- Time conversion ----

// elapsedTicks returns the monotonic tick count since start, immune to wall-clock jumps.
func (tw *TimeWheel) elapsedTicks() int64 {
	st := tw.startTime.Load()
	if st == nil {
		return 0
	}
	return int64(time.Since(*st) / tw.tickDur)
}

// nowTicks returns the current tick count (monotonic clock).
func (tw *TimeWheel) nowTicks() int64 {
	return tw.elapsedTicks()
}

// timeToTicks converts a time to ticks (Scheduler rescheduling, keeps wall-clock semantics).
func (tw *TimeWheel) timeToTicks(t time.Time) int64 {
	st := tw.startTime.Load()
	if st == nil {
		return 0
	}
	return int64(t.Sub(*st) / tw.tickDur)
}

// durationToTicks converts a duration to a tick count.
func (tw *TimeWheel) durationToTicks(d time.Duration) int64 {
	return int64(d / tw.tickDur)
}

// ticksToTime converts a tick count to a time (Scheduler rescheduling).
func (tw *TimeWheel) ticksToTime(ticks int64) time.Time {
	st := tw.startTime.Load()
	if st == nil {
		return time.Time{}
	}
	return st.Add(time.Duration(ticks) * tw.tickDur)
}
