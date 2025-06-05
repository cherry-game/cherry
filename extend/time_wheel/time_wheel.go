// Package cherryTimeWheel file from https://github.com/RussellLuo/timingwheel
package cherryTimeWheel

import (
	"sync/atomic"
	"time"
	"unsafe"

	cutils "github.com/cherry-game/cherry/extend/utils"
	clog "github.com/cherry-game/cherry/logger"
)

// TimeWheel is an implementation of Hierarchical Timing Wheels.
type TimeWheel struct {
	tick          int64            // in milliseconds
	wheelSize     int64            // wheel size
	interval      int64            // in milliseconds
	currentTime   int64            // in milliseconds
	buckets       []*bucket        // bucket list
	queue         *DelayQueue      // delay queue
	overflowWheel unsafe.Pointer   // type: *TimeWheel The higher-level overflow wheel.
	exitC         chan struct{}    // exit chan
	waitGroup     waitGroupWrapper // wait group
}

// NewTimeWheel creates an instance of TimeWheel with the given tick and wheelSize.
func NewTimeWheel(tick time.Duration, wheelSize int64) *TimeWheel {
	tickMs := int64(tick / time.Millisecond)
	if tickMs <= 0 {
		clog.Error("tick must be greater than or equal to 1ms")
		return nil
	}

	startMs := TimeToMS(time.Now().UTC())

	return newTimingWheel(
		tickMs,
		wheelSize,
		startMs,
		NewDelayQueue(int(wheelSize)),
	)
}

// newTimingWheel is an internal helper function that really creates an instance of TimeWheel.
func newTimingWheel(tickMs int64, wheelSize int64, startMs int64, queue *DelayQueue) *TimeWheel {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}

	return &TimeWheel{
		tick:        tickMs,
		wheelSize:   wheelSize,
		currentTime: truncate(startMs, tickMs),
		interval:    tickMs * wheelSize,
		buckets:     buckets,
		queue:       queue,
		exitC:       make(chan struct{}),
	}
}

// add inserts the timer t into the current timing wheel.
func (tw *TimeWheel) add(t *Timer) bool {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if t.expiration < currentTime+tw.tick {
		// Already expired
		return false
	}

	if t.expiration < currentTime+tw.interval {
		// Put it into its own bucket
		virtualID := t.expiration / tw.tick
		b := tw.buckets[virtualID%tw.wheelSize]
		b.Add(t)

		// Set the bucket expiration time
		if b.SetExpiration(virtualID * tw.tick) {
			// The bucket needs to be enqueued since it was an expired bucket.
			// We only need to enqueue the bucket when its expiration time has changed,
			// i.e. the wheel has advanced and this bucket get reused with a new expiration.
			// Any further calls to set the expiration within the same wheel cycle will
			// pass in the same value and hence return false, thus the bucket with the
			// same expiration will not be enqueued multiple times.
			tw.queue.Offer(b, b.Expiration())
		}
		return true
	} else {
		// Out of the interval. Put it into the overflow wheel
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				unsafe.Pointer(newTimingWheel(
					tw.interval,
					tw.wheelSize,
					currentTime,
					tw.queue,
				)),
			)
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}

		return (*TimeWheel)(overflowWheel).add(t)
	}
}

// addOrRun inserts the timer t into the current timing wheel, or run the
// timer's task if it has already expired.
func (tw *TimeWheel) addOrRun(t *Timer) {
	if !tw.add(t) {
		// Already expired
		// Like the standard time.AfterFunc (https://golang.org/pkg/time/#AfterFunc),
		// always execute the timer's task in its own goroutine.
		if t.isAsync {
			go t.task()
		} else {
			t.task()
		}
	}
}

func (tw *TimeWheel) advanceClock(expiration int64) {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if expiration >= currentTime+tw.tick {
		currentTime = truncate(expiration, tw.tick)
		atomic.StoreInt64(&tw.currentTime, currentTime)

		// Try to advance the clock of the overflow wheel if present
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel != nil {
			(*TimeWheel)(overflowWheel).advanceClock(currentTime)
		}
	}
}

// Start starts the current timing wheel.
func (tw *TimeWheel) Start() {
	tw.waitGroup.Wrap(func() {
		tw.queue.Poll(tw.exitC, func() int64 {
			return TimeToMS(time.Now().UTC())
		})
	})

	tw.waitGroup.Wrap(func() {
		for {
			select {
			case elem := <-tw.queue.C:
				b := elem.(*bucket)
				tw.advanceClock(b.Expiration())
				b.Flush(tw.addOrRun)
			case <-tw.exitC:
				return
			}
		}
	})
}

// Stop stops the current timing wheel.
//
// If there is any timer's task being running in its own goroutine, Stop does
// not wait for the task to complete before returning. If the caller needs to
// know whether the task is completed, it must coordinate with the task explicitly.
func (tw *TimeWheel) Stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
// It returns a Timer that can be used to cancel the call using its Stop method.
func (tw *TimeWheel) AfterFunc(id uint64, d time.Duration, f func(), async ...bool) *Timer {
	t := &Timer{
		id:         id,
		expiration: TimeToMS(time.Now().UTC().Add(d)),
		task:       f,
		isAsync:    getAsyncValue(async...),
	}
	tw.addOrRun(t)

	return t
}

func (tw *TimeWheel) AddEveryFunc(id uint64, d time.Duration, f func(), async ...bool) *Timer {
	return tw.ScheduleFunc(id, &EverySchedule{Interval: d}, f, async...)
}

func (tw *TimeWheel) BuildAfterFunc(d time.Duration, f func()) *Timer {
	id := NextID()
	return tw.AfterFunc(id, d, f)
}

func (tw *TimeWheel) BuildEveryFunc(d time.Duration, f func(), async ...bool) *Timer {
	id := NextID()
	return tw.AddEveryFunc(id, d, f, async...)
}

// ScheduleFunc calls f (in its own goroutine) according to the execution
// plan scheduled by s. It returns a Timer that can be used to cancel the
// call using its Stop method.
//
// If the caller want to terminate the execution plan halfway, it must
// stop the timer and ensure that the timer is stopped actually, since in
// the current implementation, there is a gap between the expiring and the
// restarting of the timer. The wait time for ensuring is short since the
// gap is very small.
//
// Internally, ScheduleFunc will ask the first execution time (by calling
// s.Next()) initially, and create a timer if the execution time is non-zero.
// Afterwards, it will ask the next execution time each time f is about to
// be executed, and f will be called at the next execution time if the time
// is non-zero.
func (tw *TimeWheel) ScheduleFunc(id uint64, s Scheduler, f func(), async ...bool) *Timer {
	expiration := s.Next(time.Now())
	if expiration.IsZero() {
		// No time is scheduled, return nil.
		return nil
	}

	t := &Timer{
		id:         id,
		expiration: TimeToMS(expiration),
		isAsync:    getAsyncValue(async...),
	}

	t.task = func() {
		// Schedule the task to execute at the next time if possible.
		nextExpiration := s.Next(MSToTime(t.expiration))
		if !expiration.IsZero() {
			t.expiration = TimeToMS(nextExpiration)
			tw.addOrRun(t)
		}

		// Actually execute the task.
		cutils.Try(f, func(errString string) {
			clog.Warn(errString)
		})
	}

	tw.addOrRun(t)
	return t
}

func (tw *TimeWheel) NextID() uint64 {
	return NextID()
}

func getAsyncValue(asyncTask ...bool) bool {
	if len(asyncTask) > 0 {
		return asyncTask[0]
	}
	return false
}
