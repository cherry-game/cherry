package cherryTimeWheel

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestWheel(tick time.Duration) *TimeWheel {
	tw := NewTimeWheel(tick)
	tw.Start()
	return tw
}

func TestAddOnce_Fires(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddTimer(50*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	time.Sleep(200 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("AddOnce did not fire")
	}
}

func TestAddOnce_FiresOnce(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	tw.AddTimer(50*time.Millisecond, func() {
		count.Add(1)
	}, true)

	time.Sleep(300 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Fatalf("AddOnce fired %d times, expected 1", c)
	}
}

func TestAdd_Repeats(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	tw.AddTimer(30*time.Millisecond, func() {
		count.Add(1)
	}, false)

	time.Sleep(250 * time.Millisecond)
	c := count.Load()
	if c < 5 || c > 10 {
		t.Fatalf("Add fired %d times, expected 6~9", c)
	}
}

func TestAdd_SubTickInterval(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	tw.AddTimer(3*time.Millisecond, func() {
		count.Add(1)
	}, false)

	time.Sleep(100 * time.Millisecond)
	if c := count.Load(); c < 3 {
		t.Fatalf("sub-tick interval fired %d times, expected >=3", c)
	}
}

func TestTimerStop_PreventsFire(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	timer := tw.AddTimer(200*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	timer.Stop()

	time.Sleep(400 * time.Millisecond)
	if fired.Load() == 1 {
		t.Fatal("timer fired after Stop")
	}
}

func TestTimerStopAndStart(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(30*time.Millisecond, func() {
		count.Add(1)
	}, false)

	time.Sleep(120 * time.Millisecond)
	timer.Stop()

	// The removeCmd reclaim is asynchronous: wait until the driver drains it
	// before asserting, so the node is fully detached.
	deadline := time.Now().Add(time.Second)
	for tw.ActiveCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d after Stop, want 0", tw.ActiveCount())
		}
		time.Sleep(time.Millisecond)
	}

	before := count.Load()
	time.Sleep(200 * time.Millisecond)
	if count.Load() != before {
		t.Fatal("timer fired after Stop")
	}

	timer.Start()
	time.Sleep(120 * time.Millisecond)
	if count.Load() <= before {
		t.Fatal("timer did not fire after restart")
	}
}

func TestTimerStart_WhileRunning(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.AddTimer(20*time.Millisecond, func() {}, false)

	time.Sleep(60 * time.Millisecond)
	// Start while already running: the old node must be removed before the new one
	// is enqueued, leaving exactly one active node.
	timer.Start()

	deadline := time.Now().Add(time.Second)
	for tw.ActiveCount() != 1 {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d after Start-while-running, want 1", tw.ActiveCount())
		}
		time.Sleep(time.Millisecond)
	}
}

func TestTimerStop_AfterFired(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	timer := tw.AddTimer(20*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	time.Sleep(150 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("timer did not fire in time")
	}

	timer.Stop() // must not panic; the node is already gone from the wheel
}

func TestAddSchedule_EverySchedule(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	tw.AddScheduleTimer(&EverySchedule{Interval: 30 * time.Millisecond}, func() {
		count.Add(1)
	})

	time.Sleep(250 * time.Millisecond)
	if c := count.Load(); c < 5 {
		t.Fatalf("AddSchedule (Every) fired %d times, expected >=5", c)
	}
}

func TestAddSchedule_FixedDate(t *testing.T) {
	tw := newTestWheel(100 * time.Millisecond)
	defer tw.Stop()

	now := time.Now()
	min, sec := now.Minute(), now.Second()+1
	if sec >= 60 {
		min++
		sec = 0
	}

	var fired atomic.Int32
	tw.AddScheduleTimer(&FixedDateSchedule{Hour: -1, Minute: min, Second: sec}, func() {
		fired.Store(1)
	})

	time.Sleep(3 * time.Second)
	if fired.Load() != 1 {
		t.Fatal("FixedDateSchedule did not fire")
	}
}

func TestCascade_NearWrap(t *testing.T) {
	tw := newTestWheel(time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddTimer(300*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	time.Sleep(600 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("long timer did not fire (cascade failed)")
	}
}

func TestCascade_MultiLevel(t *testing.T) {
	tw := newTestWheel(time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddTimer(20*time.Second, func() {
		fired.Store(1)
	}, true)

	time.Sleep(25 * time.Second)
	if fired.Load() != 1 {
		t.Fatal("very long timer did not fire (multi-level cascade failed)")
	}
}

// TestCascade_ResetCancelledNode verifies a cancelled node (running=false) is
// reset by cascade instead of re-enqueued. Regression: cascade used to key on
// cb != nil, but Stop clears running, not cb, so cancelled nodes lingered until
// their slot was swept.
func TestCascade_ResetCancelledNode(t *testing.T) {
	tw := NewTimeWheel(10 * time.Millisecond)

	node := &timerNode{id: 1}
	node.expire = 1 << 14
	node.cb = func() {}
	node.running.Store(false) // Stop semantics: running=false, cb still set

	tw.t[1][1] = node
	tw.cascade(1, 1<<14) // idx = (1<<14 >> 14) & 0x3F = 1 → cascades t[1][1]

	if node.cb != nil {
		t.Fatal("cancelled node was re-enqueued instead of reset")
	}
	if node.interval != 0 || node.schedule != nil {
		t.Fatal("cancelled node not fully reset")
	}
}

// TestReset_ClearsLinks verifies reset clears the linked-list pointers, so a
// finished node held by a Timer handle cannot keep sibling nodes reachable via
// its residual next/prev (which would pin them in memory until the handle is
// released).
func TestReset_ClearsLinks(t *testing.T) {
	tw := NewTimeWheel(10 * time.Millisecond)

	a := &timerNode{id: 1}
	b := &timerNode{id: 2}
	a.next = b
	b.prev = a

	tw.reset(a)

	if a.next != nil || a.prev != nil {
		t.Fatal("reset should clear linked-list pointers")
	}
}

func TestConcurrentAddStop(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var wg sync.WaitGroup
	n := 100

	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tw.AddTimer(100*time.Millisecond, func() {}, true)
		}()
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)
}

func TestAddOnce_Dispatch(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddTimer(50*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	time.Sleep(200 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("AddOnce did not fire via dispatch")
	}
}

func TestAfterStop_NoMoreFires(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)

	var count atomic.Int32
	tw.AddTimer(30*time.Millisecond, func() {
		count.Add(1)
	}, false)

	time.Sleep(100 * time.Millisecond)
	tw.Stop()

	before := count.Load()
	time.Sleep(200 * time.Millisecond)
	after := count.Load()

	if after != before {
		t.Fatalf("timer fired after Stop: before=%d, after=%d", before, after)
	}
}

// ---- Start/Stop idempotency ----

func TestStart_Idempotent(t *testing.T) {
	tw := NewTimeWheel(10 * time.Millisecond)
	tw.Start()
	tw.Start() // must not panic
	tw.Stop()
}

func TestStop_Idempotent(t *testing.T) {
	tw := NewTimeWheel(10 * time.Millisecond)
	tw.Start()
	tw.Stop()
	tw.Stop() // must not panic
}

func TestStop_Closed_NoRestart(t *testing.T) {
	tw := NewTimeWheel(10 * time.Millisecond)

	tw.Start()
	tw.Stop()

	// Stop is terminal: a second Start must not revive a closed wheel.
	tw.Start()

	var fired atomic.Int32
	tw.AddTimer(50*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	time.Sleep(200 * time.Millisecond)
	if fired.Load() != 0 {
		t.Fatal("timer fired after a closed wheel was restarted")
	}
	if got := tw.ActiveCount(); got != 0 {
		t.Fatalf("ActiveCount = %d after closed wheel, want 0", got)
	}
}

func TestNewTimeWheel_TickFallback(t *testing.T) {
	// tick < 1ms falls back to DefaultTick (10ms) instead of failing.
	tw := NewTimeWheel(500 * time.Microsecond)
	tw.Start()
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddTimer(50*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	time.Sleep(200 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("timer did not fire after tick fallback")
	}
}

// ---- Timer.Clear ----

func TestTimerClear_PreventsFire(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	timer := tw.AddTimer(50*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	timer.Clear()

	time.Sleep(200 * time.Millisecond)
	if fired.Load() == 1 {
		t.Fatal("timer fired after Clear")
	}
}

func TestTimerClear_Idempotent(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.AddTimer(50*time.Millisecond, func() {}, true)
	timer.Clear()
	timer.Clear() // must not panic
}

// ---- Timer.Stop repeat ----

func TestTimerStop_AlreadyStopped(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.AddTimer(200*time.Millisecond, func() {}, true)

	timer.Stop()
	timer.Stop() // must not panic (no-op after the first Stop)
}

// ---- AddOnce Stop then Start becomes recurring ----

func TestAddOnce_StopStart(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(100*time.Millisecond, func() {
		count.Add(1)
	}, true)

	// cancel before it fires
	time.Sleep(50 * time.Millisecond)
	timer.Stop()

	// Start keeps the original one-shot type
	timer.Start()

	time.Sleep(300 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Fatalf("AddOnce + Stop + Start should fire once (one-shot kept), got %d", c)
	}
}

// ---- submit after Stop ----

func TestStop_RemoveCmdAfterStop(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)

	timer := tw.AddTimer(time.Hour, func() {}, true)
	tw.Stop()

	timer.Stop() // must not panic after TimeWheel is stopped
}

// ---- AddSchedule boundary ----

type nilSchedule struct{}

func (s *nilSchedule) Next(prev time.Time) time.Time {
	return time.Time{}
}

func TestAddSchedule_NilNext(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.AddScheduleTimer(&nilSchedule{}, func() {})
	if timer != nil {
		t.Fatal("AddSchedule should return nil when Next returns zero time")
	}
}

func TestAddSchedule_StopNoFire(t *testing.T) {
	tw := NewTimeWheel(10 * time.Millisecond)
	tw.Start()

	var count atomic.Int32
	tw.AddScheduleTimer(&EverySchedule{Interval: 30 * time.Millisecond}, func() {
		count.Add(1)
	})

	time.Sleep(100 * time.Millisecond)
	tw.Stop()

	before := count.Load()
	time.Sleep(200 * time.Millisecond)
	if count.Load() != before {
		t.Fatalf("AddSchedule fired after Stop: before=%d, after=%d", before, count.Load())
	}
}

// ---- concurrent Stop + Start ----

func TestConcurrentStopStart(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var wg sync.WaitGroup
	var count atomic.Int32
	n := 50

	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timer := tw.AddTimer(50*time.Millisecond, func() {
				count.Add(1)
			}, true)
			timer.Stop()
			timer.Start()
			time.Sleep(100 * time.Millisecond)
			timer.Stop()
		}()
	}

	wg.Wait()
}

// TestTimerStopStart_NodeIdentity verifies that after Stop→Start, a same-id
// different-node is not wrongly deleted by dispatchList and the restart works.
func TestTimerStopStart_NodeIdentity(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.AddTimer(20*time.Millisecond, func() {}, false)

	time.Sleep(30 * time.Millisecond)
	timer.Stop()

	timer.Start()

	time.Sleep(60 * time.Millisecond)
	if got := tw.ActiveCount(); got != 1 {
		t.Fatalf("ActiveCount = %d after restart, want 1", got)
	}

	timer.Stop()
	deadline := time.Now().Add(time.Second)
	for tw.ActiveCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d after second Stop, want 0", tw.ActiveCount())
		}
		time.Sleep(time.Millisecond)
	}
}

func BenchmarkAddOnce(b *testing.B) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	b.ResetTimer()
	for b.Loop() {
		tw.AddTimer(time.Second, func() {}, true)
	}
}

func BenchmarkAdd(b *testing.B) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	b.ResetTimer()
	for b.Loop() {
		tw.AddTimer(time.Second, func() {}, false)
	}
}

func BenchmarkTimerStop(b *testing.B) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	b.ResetTimer()
	for b.Loop() {
		timer := tw.AddTimer(time.Hour, func() {}, true)
		timer.Stop()
	}
}

// TestMillionTimers creates one million one-shot timers with scattered delays
// and verifies they all fire correctly. Skipped in -short mode.
func TestMillionTimers(t *testing.T) {
	if testing.Short() {
		t.Skip("skip million-timer test in short mode")
	}

	tw := NewTimeWheelWithHint(10*time.Millisecond, 1<<20)
	tw.Start()
	defer tw.Stop()

	const n = 1000000
	var fired atomic.Int64

	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	for i := range n {
		// scatter delays over 2s ~ 5s so all timers stay active while injecting
		d := 2*time.Second + time.Duration(i%3000)*time.Millisecond
		tw.AddTimer(d, func() { fired.Add(1) }, true)
	}
	t.Logf("inject %d timers in %v", n, time.Since(start))

	// wait until all addCmds are consumed by the driver
	deadline := time.Now().Add(15 * time.Second)
	for tw.ActiveCount() != int64(n) {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d, want %d", tw.ActiveCount(), n)
		}
		time.Sleep(10 * time.Millisecond)
	}

	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	t.Logf("HeapAlloc delta ~%dMB", (memAfter.HeapAlloc-memBefore.HeapAlloc)>>20)

	// wait until all timers fire (max delay 5s + slack)
	deadline = time.Now().Add(15 * time.Second)
	for fired.Load() != int64(n) {
		if time.Now().After(deadline) {
			t.Fatalf("fired %d/%d", fired.Load(), n)
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got := tw.ActiveCount(); got != 0 {
		t.Fatalf("ActiveCount = %d after all fired, want 0", got)
	}
}

// AddSchedule with a sub-tick interval must fire every tick, not degrade to
// once per near-wheel rotation (regression for the missing expire > current
// guard on the schedule reschedule path).
func TestAddSchedule_SubTickInterval(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	tw.AddScheduleTimer(&EverySchedule{Interval: 3 * time.Millisecond}, func() {
		count.Add(1)
	})

	time.Sleep(100 * time.Millisecond)
	if c := count.Load(); c < 3 {
		t.Fatalf("sub-tick schedule fired %d times, expected >=3", c)
	}
}

// Clear releases the callback; a subsequent Start must not launch a timer
// with a nil callback (regression: nil callback fired, panic swallowed).
func TestTimerClear_ThenStart(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(20*time.Millisecond, func() {
		count.Add(1)
	}, false)
	timer.Clear()
	timer.Start() // callback was released; must not start a timer

	time.Sleep(80 * time.Millisecond)
	if count.Load() != 0 {
		t.Fatalf("timer fired after Clear+Start, count=%d", count.Load())
	}
	if got := tw.ActiveCount(); got != 0 {
		t.Fatalf("ActiveCount = %d after Clear+Start, want 0", got)
	}
}

// A one-shot timer that reuses a pool node previously owned by a recurring
// timer must fire exactly once (regression: stale interval leaked into reuse
// and turned the one-shot into a periodic timer).
func TestAddOnce_AfterRecurringPoolReuse(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	// Recurring timer: its node returns to the pool on Stop.
	recurring := tw.AddTimer(10*time.Millisecond, func() {}, false)
	time.Sleep(30 * time.Millisecond)
	recurring.Stop()

	deadline := time.Now().Add(time.Second)
	for tw.ActiveCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d, want 0", tw.ActiveCount())
		}
		time.Sleep(time.Millisecond)
	}

	// One-shot reusing a pooled node: must fire exactly once.
	var count atomic.Int32
	tw.AddTimer(20*time.Millisecond, func() {
		count.Add(1)
	}, true)

	time.Sleep(200 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Fatalf("one-shot fired %d times after reusing a recurring node, want 1", c)
	}
}

// Concurrent Add/Timer.Stop from many goroutines must be race-free and drain
// cleanly while the wheel keeps running.
func TestConcurrentStopAndAdd(t *testing.T) {
	tw := NewTimeWheelWithHint(10*time.Millisecond, 1<<10)
	tw.Start()
	defer tw.Stop()

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				timer := tw.AddTimer(time.Millisecond, func() {}, false)
				timer.Stop()
			}
		}()
	}
	wg.Wait()

	// The removeCmd reclaim is asynchronous: wait until the driver drains all removeCmds.
	deadline := time.Now().Add(time.Second)
	for tw.ActiveCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d after concurrent Add/Stop, want 0", tw.ActiveCount())
		}
		time.Sleep(time.Millisecond)
	}
}

// ---- NewTimer / NewScheduleTimer (construct-then-start) ----

func TestNewTimer_NotStarted(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	timer := tw.NewTimer(30*time.Millisecond, func() {
		fired.Store(1)
	}, true)

	// Constructed but not started: must not fire nor occupy the wheel.
	time.Sleep(100 * time.Millisecond)
	if fired.Load() != 0 {
		t.Fatal("NewTimer fired before Start")
	}
	if got := tw.ActiveCount(); got != 0 {
		t.Fatalf("ActiveCount = %d before Start, want 0", got)
	}

	timer.Start()
	time.Sleep(100 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("NewTimer did not fire after Start")
	}
}

func TestNewTimer_Recurring(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.NewTimer(30*time.Millisecond, func() {
		count.Add(1)
	}, false)
	timer.Start()

	time.Sleep(200 * time.Millisecond)
	if c := count.Load(); c < 4 {
		t.Fatalf("recurring NewTimer fired %d times, expected >=4", c)
	}
}

func TestNewScheduleTimer_NotStarted(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.NewScheduleTimer(&EverySchedule{Interval: 30 * time.Millisecond}, func() {
		count.Add(1)
	})

	// Constructed but not started: must not fire nor occupy the wheel.
	time.Sleep(100 * time.Millisecond)
	if count.Load() != 0 {
		t.Fatal("NewScheduleTimer fired before Start")
	}
	if got := tw.ActiveCount(); got != 0 {
		t.Fatalf("ActiveCount = %d before Start, want 0", got)
	}

	timer.Start()
	time.Sleep(200 * time.Millisecond)
	if c := count.Load(); c < 4 {
		t.Fatalf("NewScheduleTimer fired %d times after Start, expected >=4", c)
	}
}

func TestNewScheduleTimer_NoNext(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.NewScheduleTimer(&nilSchedule{}, func() {})
	timer.Start() // scheduler has no next expiry: nothing scheduled, no panic

	if timer.IsRunning() {
		t.Fatal("timer reported running when nothing was scheduled")
	}
	if got := tw.ActiveCount(); got != 0 {
		t.Fatalf("ActiveCount = %d, want 0", got)
	}
}

// ---- SetNext (dynamic next-delay callback) ----

// ---- once + SetNext lifecycle ----
//
// A one-shot timer uses its *current* delay for each Start: the original delay
// d before a SetNext, the fn() delay after one. Each Start fires exactly once
// and then stops. SetNext never changes an already-queued fire (nextCmd only
// records the callback).
//
//   - Scenario 1 (SetNext after the first Start): first fire at d, restart at fn().
//   - Scenario 2 (SetNext before the first Start): the first fire already uses fn().
//   - Scenario 3 (SetNext inside the callback) is an actor-layer pattern (the
//     callback runs on the actor goroutine, the handle owner). It is not tested
//     here because at the wheel layer callbacks run on the driver goroutine,
//     which would race with the test goroutine driving the handle.

func TestSetNext_Once_UsesNextFuncDelay(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.NewTimer(100*time.Millisecond, func() {
		count.Add(1)
	}, true) // one-shot, not started; d=100ms
	timer.SetNext(func() time.Duration { return 20 * time.Millisecond }) // use 20ms for the single fire
	timer.Start()

	time.Sleep(50 * time.Millisecond) // fired at ~20ms, well before d=100ms
	if c := count.Load(); c != 1 {
		t.Fatalf("once + SetNext fired %d times by 50ms, expected 1 (used nextFunc delay, not d)", c)
	}
	if timer.IsRunning() {
		t.Fatal("timer still running after once fire")
	}
}

func TestSetNext_Once_RunningNoRecompute(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(150*time.Millisecond, func() {
		count.Add(1)
	}, true) // once, d=150ms
	time.Sleep(50 * time.Millisecond) // still waiting for the first fire
	timer.SetNext(func() time.Duration { return 10 * time.Millisecond })

	// A nextCmd must NOT recompute a one-shot: the first fire stays at d=150ms.
	// If it recomputed, the fire would move to ~60ms and count would be 1 here.
	time.Sleep(50 * time.Millisecond) // now at ~100ms < d
	if c := count.Load(); c != 0 {
		t.Fatalf("once timer fired early (%d times), expected 0 (nextCmd must not recompute a one-shot)", c)
	}

	time.Sleep(80 * time.Millisecond) // now at ~180ms > d=150ms
	if c := count.Load(); c != 1 {
		t.Fatalf("once timer fired %d times, expected exactly 1", c)
	}
	if timer.IsRunning() {
		t.Fatal("timer still running after once fire")
	}
}

// Scenario 1: SetNext after the first Start — the restart uses the fn() delay.
func TestSetNext_Once_StartAfterSetNext(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.NewTimer(50*time.Millisecond, func() {
		count.Add(1)
	}, true) // once, d=50ms
	timer.Start() // first Start: no nextFunc yet → fire at d=50ms

	time.Sleep(80 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Fatalf("first Start fired %d times, expected 1 (d=50ms)", c)
	}

	timer.SetNext(func() time.Duration { return 20 * time.Millisecond })
	timer.Start() // restart: nextFunc set → fire at fn()=20ms

	time.Sleep(40 * time.Millisecond) // 20ms after restart
	if c := count.Load(); c != 2 {
		t.Fatalf("after restart fired %d times, expected 2 (SetNext delay used)", c)
	}
}

func TestSetNext_StopsOnNonPositive(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(20*time.Millisecond, func() {
		count.Add(1)
	}, false) // recurring
	timer.SetNext(func() time.Duration { return 0 }) // first fire at d, then stop

	time.Sleep(200 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Fatalf("timer fired %d times, expected 1 (first fire at d, then SetNext(0) stops)", c)
	}
	if timer.IsRunning() {
		t.Fatal("timer reported running after SetNext returned 0")
	}
}

func TestSetNext_DynamicSequence(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	delays := []time.Duration{20 * time.Millisecond, 40 * time.Millisecond, 0}
	var idx atomic.Int32
	var count atomic.Int32
	timer := tw.AddTimer(20*time.Millisecond, func() {
		count.Add(1)
	}, false)
	timer.SetNext(func() time.Duration {
		i := idx.Add(1)
		if int(i-1) >= len(delays) {
			return 0
		}
		return delays[i-1]
	})
	// fn is invoked after each fire: first fire at d=20ms → delays[0] (20ms) →
	// fire → delays[1] (40ms) → fire → delays[2] (0) stop. Total 3 fires.

	time.Sleep(400 * time.Millisecond)
	if c := count.Load(); c != 3 {
		t.Fatalf("dynamic sequence fired %d times, expected 3 (20ms,40ms then stop)", c)
	}
}

func TestNew_ThenSetNext_ThenStart(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.NewTimer(30*time.Millisecond, func() {
		count.Add(1)
	}, false) // recurring, not started
	timer.SetNext(func() time.Duration { return 30 * time.Millisecond })
	timer.Start()

	time.Sleep(250 * time.Millisecond)
	if c := count.Load(); c < 4 {
		t.Fatalf("New+SetNext+Start fired %d times, expected >=4", c)
	}
}

func TestSetNext_NilClears(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(20*time.Millisecond, func() {
		count.Add(1)
	}, true) // one-shot
	timer.SetNext(func() time.Duration { return 20 * time.Millisecond })
	timer.SetNext(nil) // revert to one-shot

	time.Sleep(200 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Fatalf("SetNext(nil) fired %d times, expected exactly 1", c)
	}
}

func TestSetNext_StopStartKeepsDynamic(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(30*time.Millisecond, func() {
		count.Add(1)
	}, false)
	timer.SetNext(func() time.Duration { return 30 * time.Millisecond })
	timer.Stop()
	timer.Start() // nextFunc must persist across restart

	time.Sleep(150 * time.Millisecond)
	if c := count.Load(); c < 3 {
		t.Fatalf("timer fired %d times after restart, expected >=3 (nextFunc persisted)", c)
	}
}

func TestSetNext_RejectedOnSchedule(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddScheduleTimer(&EverySchedule{Interval: 30 * time.Millisecond}, func() {
		count.Add(1)
	})
	timer.SetNext(func() time.Duration { return time.Hour }) // rejected: schedule is its own system
	timer.SetNext(nil)                                       // also rejected, no panic

	if timer.nextFunc != nil {
		t.Fatal("SetNext was accepted on a schedule-driven timer")
	}

	time.Sleep(250 * time.Millisecond)
	if c := count.Load(); c < 5 {
		t.Fatalf("schedule timer fired %d times, expected >=5", c)
	}
}

func TestSetNext_RaceFree(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddTimer(20*time.Millisecond, func() {
		count.Add(1)
	}, false)
	timer.SetNext(func() time.Duration { return 20 * time.Millisecond })

	time.Sleep(80 * time.Millisecond)
	for range 10 {
		timer.SetNext(func() time.Duration { return 20 * time.Millisecond })
		time.Sleep(time.Millisecond)
	}
	if count.Load() == 0 {
		t.Fatal("timer did not fire with SetNext")
	}
}
