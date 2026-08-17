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
	tw.AddOnceTimer(50*time.Millisecond, func() {
		fired.Store(1)
	})

	time.Sleep(200 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("AddOnce did not fire")
	}
}

func TestAddOnce_FiresOnce(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	tw.AddOnceTimer(50*time.Millisecond, func() {
		count.Add(1)
	})

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
	})

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
	})

	time.Sleep(100 * time.Millisecond)
	if c := count.Load(); c < 3 {
		t.Fatalf("sub-tick interval fired %d times, expected >=3", c)
	}
}

func TestTimerStop_PreventsFire(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	timer := tw.AddOnceTimer(200*time.Millisecond, func() {
		fired.Store(1)
	})

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
	})

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

	timer := tw.AddTimer(20*time.Millisecond, func() {})

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
	timer := tw.AddOnceTimer(20*time.Millisecond, func() {
		fired.Store(1)
	})

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
	tw.AddOnceTimer(300*time.Millisecond, func() {
		fired.Store(1)
	})

	time.Sleep(600 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("long timer did not fire (cascade failed)")
	}
}

func TestCascade_MultiLevel(t *testing.T) {
	tw := newTestWheel(time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddOnceTimer(20*time.Second, func() {
		fired.Store(1)
	})

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
			tw.AddOnceTimer(100*time.Millisecond, func() {})
		}()
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)
}

func TestAddOnce_Dispatch(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddOnceTimer(50*time.Millisecond, func() {
		fired.Store(1)
	})

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
	})

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
	tw.AddOnceTimer(50*time.Millisecond, func() {
		fired.Store(1)
	})

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
	tw.AddOnceTimer(50*time.Millisecond, func() {
		fired.Store(1)
	})

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
	timer := tw.AddOnceTimer(50*time.Millisecond, func() {
		fired.Store(1)
	})

	timer.Clear()

	time.Sleep(200 * time.Millisecond)
	if fired.Load() == 1 {
		t.Fatal("timer fired after Clear")
	}
}

func TestTimerClear_Idempotent(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.AddOnceTimer(50*time.Millisecond, func() {})
	timer.Clear()
	timer.Clear() // must not panic
}

// ---- Timer.Stop repeat ----

func TestTimerStop_AlreadyStopped(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	timer := tw.AddOnceTimer(200*time.Millisecond, func() {})

	timer.Stop()
	timer.Stop() // must not panic (no-op after the first Stop)
}

// ---- AddOnce Stop then Start becomes recurring ----

func TestAddOnce_StopStart(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var count atomic.Int32
	timer := tw.AddOnceTimer(100*time.Millisecond, func() {
		count.Add(1)
	})

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

	timer := tw.AddOnceTimer(time.Hour, func() {})
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
			timer := tw.AddOnceTimer(50*time.Millisecond, func() {
				count.Add(1)
			})
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

	timer := tw.AddTimer(20*time.Millisecond, func() {})

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
		tw.AddOnceTimer(time.Second, func() {})
	}
}

func BenchmarkAdd(b *testing.B) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	b.ResetTimer()
	for b.Loop() {
		tw.AddTimer(time.Second, func() {})
	}
}

func BenchmarkTimerStop(b *testing.B) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	b.ResetTimer()
	for b.Loop() {
		timer := tw.AddOnceTimer(time.Hour, func() {})
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
		tw.AddOnceTimer(d, func() { fired.Add(1) })
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
	})
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
	recurring := tw.AddTimer(10*time.Millisecond, func() {})
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
	tw.AddOnceTimer(20*time.Millisecond, func() {
		count.Add(1)
	})

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
				timer := tw.AddTimer(time.Millisecond, func() {})
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
