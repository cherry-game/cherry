package cherryTimeWheel

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Interleaving Stop and Timer.Stop must not deadlock
func TestConcurrentStopAndTimerStop(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)

	timers := make([]*Timer, 0, 1000)
	for range 1000 {
		timers = append(timers, tw.AddTimer(time.Hour, func() {}))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tw.Stop()
	}()
	for _, timer := range timers {
		wg.Add(1)
		go func(tm *Timer) {
			defer wg.Done()
			tm.Stop()
		}(timer)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("deadlock between Stop and Timer.Stop")
	}
}

// A panicking dispatch callback must not kill the driver goroutine
func TestDispatchPanicDoesNotKillWheel(t *testing.T) {
	tw := newTestWheel(10 * time.Millisecond)
	defer tw.Stop()

	var fired atomic.Int32
	tw.AddOnceTimer(20*time.Millisecond, func() { panic("boom") })
	tw.AddOnceTimer(40*time.Millisecond, func() { fired.Store(1) })

	time.Sleep(200 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("wheel died after dispatch panic")
	}
}

// Add before Start must not panic and must fire correctly after Start
func TestAddBeforeStart(t *testing.T) {
	tw := NewTimeWheel(10 * time.Millisecond)

	var fired atomic.Int32
	timer := tw.AddOnceTimer(50*time.Millisecond, func() { fired.Store(1) })
	if timer == nil {
		t.Fatal("AddOnce returned nil")
	}

	tw.Start()
	defer tw.Stop()

	time.Sleep(200 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatal("timer added before Start did not fire")
	}
}

// Cancel immediately removes the node; ActiveCount drops to zero
func TestCancelFreesNode(t *testing.T) {
	tw := NewTimeWheelWithHint(10*time.Millisecond, 1024)
	tw.Start()
	defer tw.Stop()

	const n = 1000
	timers := make([]*Timer, n)
	for i := range timers {
		timers[i] = tw.AddOnceTimer(time.Hour, func() {})
	}

	// wait until all addCmds are consumed (AddOnce submits asynchronously)
	deadline := time.Now().Add(2 * time.Second)
	for tw.ActiveCount() != int64(n) {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d, want %d", tw.ActiveCount(), n)
		}
		time.Sleep(time.Millisecond)
	}
	for _, tm := range timers {
		tm.Stop()
	}
	// The removeCmd reclaim is asynchronous; wait until the driver drains all removeCmds.
	deadline = time.Now().Add(2 * time.Second)
	for tw.ActiveCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d after cancel, want 0", tw.ActiveCount())
		}
		time.Sleep(time.Millisecond)
	}
}

// Million-scale benchmark: add throughput with pre-allocated nodeMap
func BenchmarkAddOnceHint(b *testing.B) {
	tw := NewTimeWheelWithHint(10*time.Millisecond, 1<<20)
	tw.Start()
	defer tw.Stop()

	b.ResetTimer()
	for b.Loop() {
		tw.AddOnceTimer(time.Hour, func() {})
	}
}

// Million-scale benchmark: add then cancel (immediate detach cost)
func BenchmarkCancelHint(b *testing.B) {
	tw := NewTimeWheelWithHint(10*time.Millisecond, 1<<20)
	tw.Start()
	defer tw.Stop()

	b.ResetTimer()
	for b.Loop() {
		timer := tw.AddOnceTimer(time.Hour, func() {})
		timer.Stop()
	}
}
