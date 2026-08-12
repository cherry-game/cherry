package cherryActor

import (
	"sync/atomic"
	"testing"
	"time"

	ctimeWheel "github.com/cherry-game/cherry/extend/time_wheel"
)

// TestActorTimer_MillionTimersPerTick creates one million one-shot timers whose
// delays are spread evenly so that exactly 100 timers expire on each tick. It
// verifies the actorTimer layer fires every timer under this million-scale,
// uniform load. Skipped in -short mode.
func TestActorTimer_MillionTimersPerTick(t *testing.T) {
	if testing.Short() {
		t.Skip("skip million-timer test in short mode")
	}

	// 1ms tick → 10000 ticks ≈ 10s to drain the whole set.
	tw := ctimeWheel.NewTimeWheelWithHint(time.Millisecond, 1<<20)
	tw.Start()
	defer tw.Stop()

	system := NewSystem()
	system.timeWheel = tw
	timer := newTimer(&Actor{system: system})

	const (
		n       = 1000000
		perTick = 100
	)
	var fired atomic.Int64

	start := time.Now()
	for i := range n {
		// delay starts at DefaultTick (AddOnce's floor) and increases by 1 tick
		// every perTick timers, spreading all n timers across n/perTick ticks.
		d := ctimeWheel.DefaultTick + time.Duration(i/perTick)*time.Millisecond
		timer.AddOnce(d, func() { fired.Add(1) })
	}
	t.Logf("inject %d timers in %v", n, time.Since(start))

	// Consume the queue, simulating the actor's processTimer (Pop + invokeFunc).
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			if id := timer.Pop(); id != 0 {
				timer.invokeFunc(id)
			} else {
				time.Sleep(time.Millisecond)
			}
		}
	}()

	deadline := time.Now().Add(30 * time.Second)
	for fired.Load() != n {
		if time.Now().After(deadline) {
			t.Fatalf("fired %d/%d", fired.Load(), n)
		}
		time.Sleep(time.Millisecond)
	}
}
