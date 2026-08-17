// Package cherryTimeWheel implements a 5-level hierarchical timer wheel.
//
// Architecture:
//   - 512 slots (256 near + 4×64 levels)
//   - Single driver goroutine, lock-free
//   - MPSC submission via timerCmd interface
//   - timerNode is owned 1:1 by its Timer handle (no object pool; a node lives as
//     long as its handle and is reclaimed by GC)
//   - Dispatch runs synchronously in the driver goroutine and only fires timers;
//     callbacks must be lightweight (business logic is pushed to actor queues by
//     the upper layer, not executed inside the wheel)
//   - A timer armed via SetNext recomputes its next delay after every fire; a
//     non-positive delay stops it (dynamic interval mode)
//
// Concurrency:
//   - Start/Stop are lifecycle methods and MUST be called serially from a single
//     goroutine (the caller serializes them); they are not synchronized internally.
//   - submitAdd/submitRemove are safe to call from any number of goroutines concurrently.
//   - A *Timer handle is NOT goroutine-safe; drive it from the single goroutine
//     that created it.
//   - Timer.Stop is immediate: the node's running flag is cleared atomically, so no
//     new callback starts after Stop returns (a callback already executing is not
//     interrupted). A remove command is still queued to reclaim the node.
//   - timerNode / nodeMap / wheel slots are owned by the driver goroutine and are
//     only touched there (the atomic running flag is also written by Stop).
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
//   - submitAdd/submitRemove use a bounded MPSC buffer; AddTimer/RemoveTimer block
//     under backpressure, which only happens if the driver goroutine stalls.
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
	near      [NEAR_SIZE]*timerNode     // level 0: 256 slots, 1-tick precision
	t         [4][LEVEL_SIZE]*timerNode // levels 1~4: 64 slots each
	current   int64                     // monotonic tick counter
	startTime atomic.Pointer[time.Time] // monotonic clock base, written by Start only
	submitCh  chan timerCmd             // MPSC submit channel (single channel keeps FIFO-ordered)
	exitCh    chan struct{}             // exit channel
	nodeMap   map[uint64]*timerNode     // id → *timerNode (driver goroutine only)
	activeNum atomic.Int64              // active timer count
	tickDur   time.Duration             // tick interval as Duration
	wg        sync.WaitGroup            // wait group
	stopped   atomic.Bool               // true = stopped (terminal); submit drops commands when set
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
		tickDur:  tick,
	}
	return tw
}

// Start begins the driver goroutine. Idempotent. Must not be called concurrently
// with Stop (lifecycle is serialized by the caller).
func (tw *TimeWheel) Start() {
	if tw.stopped.Load() || tw.exitCh != nil {
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
	if tw.stopped.Load() || tw.exitCh == nil {
		return
	}
	tw.stopped.Store(true)
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
	if tw.stopped.Load() {
		return
	}
	tw.submitCh <- &addCmd{node: node}
}

// submitRemove enqueues a timer id to be removed by the driver goroutine. Thread-safe.
func (tw *TimeWheel) submitRemove(id uint64) {
	if tw.stopped.Load() {
		return
	}
	tw.submitCh <- &removeCmd{id: id}
}

// submitNext enqueues a dynamic next-delay callback update for the driver. Thread-safe.
func (tw *TimeWheel) submitNext(id uint64, fn func() time.Duration) {
	if tw.stopped.Load() {
		return
	}
	tw.submitCh <- &nextCmd{id: id, fn: fn}
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

// reset removes node from nodeMap (if it is still the current entry) and clears
// its callback and schedule. The node stays alive while its Timer handle still
// references it (running=false marks it finished). Caller owns the activeNum
// adjustment.
func (tw *TimeWheel) reset(node *timerNode) {
	if n, ok := tw.nodeMap[node.id]; ok && n == node {
		delete(tw.nodeMap, node.id)
	}
	node.cb = nil
	node.interval = 0
	node.schedule = nil
	node.nextFunc = nil
	node.next = nil
	node.prev = nil
}

func (tw *TimeWheel) handleAddNode(node *timerNode) {
	// Defensive: replace a stale node with the same id if one is still present.
	// Only happens if the single-goroutine Timer contract is violated (e.g. a
	// concurrent Start); detach and reset it so it isn't leaked.
	if old, ok := tw.nodeMap[node.id]; ok {
		tw.detachNode(old)
		tw.reset(old)
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
		if node.running.Load() {
			tw.enqueue(node)
		} else {
			// A cancelled node left in a slot is stopped consistently with dispatchList.
			tw.stopNode(node)
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
// reschedules (schedule / recurring) or resets (one-shot).
func (tw *TimeWheel) dispatchList(head *timerNode) {
	for node := head; node != nil; {
		next := node.next
		if n, ok := tw.nodeMap[node.id]; ok && n == node {
			delete(tw.nodeMap, node.id)
		}

		// Cancelled by Stop: stop without firing.
		if !node.running.Load() {
			tw.stopNode(node)
			tw.activeNum.Add(-1)
			node = next
			continue
		}

		cutils.Try(func() {
			node.cb()
		}, func(errString string) {
			clog.Warnf("[time_wheel] dispatch panic. id=%d, err=%s", node.id, errString)
		})

		switch {
		case node.schedule != nil:
			// AddScheduleTimer: reschedule by the scheduler until it returns zero time.
			if nextExp := node.schedule.Next(tw.ticksToTime(tw.nowTicks())); !nextExp.IsZero() {
				node.expire = tw.timeToTicks(nextExp)
				tw.handleAddNode(node)
			} else {
				tw.stopNode(node)
			}
		case node.nextFunc != nil && node.interval > 0:
			// SetNext dynamic mode (recurring only): recompute the next delay after
			// every fire. A one-shot keeps its once semantics and stops here.
			if tw.rearmNext(node) {
				tw.handleAddNode(node)
			}
		case node.interval > 0:
			// Recurring: reschedule at a fixed interval.
			ticks := max(tw.durationToTicks(node.interval), 1)
			node.expire = tw.nowTicks() + ticks
			tw.handleAddNode(node)
		default:
			// One-shot: stop.
			tw.stopNode(node)
		}

		tw.activeNum.Add(-1)
		node = next
	}
}

// ---- Public API ----

// NewTimer creates a timer bound to the wheel without starting it. Call Start
// to submit it to the driver. A one-shot timer (once=true) fires once; a
// recurring timer fires at a fixed interval.
func (tw *TimeWheel) NewTimer(d time.Duration, f func(), once bool) *Timer {
	return &Timer{
		id:   nextID(),
		tw:   tw,
		d:    d,
		f:    f,
		once: once,
	}
}

// NewScheduleTimer creates a Scheduler-driven timer without starting it. Call
// Start to compute the first expiry and submit it to the driver.
func (tw *TimeWheel) NewScheduleTimer(s Scheduler, f func()) *Timer {
	return &Timer{
		id:       nextID(),
		tw:       tw,
		f:        f,
		schedule: s,
	}
}

// AddTimer creates a timer and starts it immediately. A one-shot timer
// (once=true) fires a single time and then finishes; a recurring timer fires
// repeatedly at the given interval.
func (tw *TimeWheel) AddTimer(d time.Duration, f func(), once bool) *Timer {
	t := tw.NewTimer(d, f, once)
	t.Start()
	return t
}

// AddScheduleTimer creates a Scheduler-driven timer and starts it immediately.
// It returns nil when the scheduler reports no next expiry.
func (tw *TimeWheel) AddScheduleTimer(s Scheduler, f func()) *Timer {
	t := tw.NewScheduleTimer(s, f)
	t.Start()
	if t.node == nil {
		return nil // scheduler reported no next expiry
	}
	return t
}

// RemoveTimer stops the timer with the given id. Asynchronous: the removal is
// queued to the driver, so a callback already dispatching may still fire once.
// Use Timer.Stop for an immediate cancel via the handle.
func (tw *TimeWheel) RemoveTimer(id uint64) {
	tw.submitRemove(id)
}

// newNode builds a node that fires once after delay. The caller sets interval
// for recurring timers or assigns the node to the handle.
func (tw *TimeWheel) newNode(id uint64, f func(), delay time.Duration) *timerNode {
	node := &timerNode{id: id}
	node.expire = tw.nowTicks() + max(tw.durationToTicks(delay), 1)
	node.cb = f
	node.running.Store(true)
	return node
}

// startNode builds a node bound to the given timer parameters. A recurring
// timer (once=false) carries its interval for the driver to reschedule; a
// one-shot keeps interval zero. The caller assigns the returned node to the
// timer handle and submits it.
func (tw *TimeWheel) startNode(id uint64, f func(), d time.Duration, once bool) *timerNode {
	node := tw.newNode(id, f, d)
	if !once {
		node.interval = d
	}
	return node
}

// nextDelay runs node.nextFunc with panic protection and returns the delay;
// a non-positive result means the timer should stop. Driver goroutine only.
func (tw *TimeWheel) nextDelay(node *timerNode) time.Duration {
	var next time.Duration
	cutils.Try(func() {
		next = node.nextFunc()
	}, func(errString string) {
		clog.Warnf("[time_wheel] nextFunc panic. id=%d, err=%s", node.id, errString)
	})
	return next
}

// stopNode marks a node finished and releases it from the wheel. Driver
// goroutine only. Callers handle detach (if the node is still in a slot) and
// activeNum accounting.
func (tw *TimeWheel) stopNode(node *timerNode) {
	node.running.Store(false)
	tw.reset(node)
}

// rearmNext computes the next fire from node.nextFunc and returns whether the
// node should be re-enqueued; a non-positive delay stops it. Driver goroutine
// only; the caller re-enqueues (and accounts activeNum) per its context.
func (tw *TimeWheel) rearmNext(node *timerNode) bool {
	if next := tw.nextDelay(node); next > 0 {
		node.expire = tw.nowTicks() + max(tw.durationToTicks(next), 1)
		return true
	}
	tw.stopNode(node)
	return false
}

// scheduleNode builds a node from the scheduler's first expiry, or returns nil
// when the scheduler reports no next expiry. The caller assigns the returned
// node to the timer handle and submits it.
func (tw *TimeWheel) scheduleNode(id uint64, s Scheduler, f func()) *timerNode {
	firstExp := s.Next(time.Now())
	if firstExp.IsZero() {
		return nil
	}
	node := &timerNode{id: id}
	node.expire = tw.timeToTicks(firstExp)
	node.cb = f
	node.schedule = s
	node.running.Store(true)
	return node
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
