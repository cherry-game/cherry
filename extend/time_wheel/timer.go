package cherryTimeWheel

import (
	"time"

	clog "github.com/cherry-game/cherry/logger"
)

// Timer is a user-facing handle with Start/Stop/Clear.
// It is NOT goroutine-safe: a Timer must be driven from a single goroutine (the
// owner that created it). The underlying *timerNode is owned by the wheel's driver
// goroutine and only mutated through the command channel.
type Timer struct {
	id       uint64               // unique ID
	d        time.Duration        // bound interval
	f        func()               // bound callback
	once     bool                 // one-shot vs recurring
	schedule Scheduler            // Scheduler-driven; nil for interval/one-shot timers
	nextFunc func() time.Duration // dynamic next-delay (SetNext); nil = disabled
	node     *timerNode           // non-nil when running; single-goroutine owner only
	tw       *TimeWheel           // owner
}

// ID returns the timer's unique ID.
func (t *Timer) ID() uint64 {
	return t.id
}

// Start submits the timer to the wheel and starts it, keeping its original type:
// a one-shot timer stays one-shot, a recurring timer stays recurring. A
// schedule-driven timer recomputes its first expiry from the scheduler; if the
// scheduler reports no next expiry, nothing is scheduled.
func (t *Timer) Start() {
	if t.f == nil {
		return
	}
	// Restart safety: the timer ID is shared, so a running node must be removed
	// before enqueuing a new one, otherwise the wheel holds two nodes with the
	// same id (the old one would be leaked until its slot is swept).
	if t.node != nil {
		t.Stop()
	}

	var node *timerNode
	if t.schedule != nil {
		node = t.tw.scheduleNode(t.id, t.schedule, t.f)
		if node == nil {
			return
		}
	} else if t.once && t.nextFunc != nil {
		// One-shot timer with a SetNext delay: compute the single fire time from
		// fn; without one it fires at the original delay d. It still fires once.
		delay := t.nextFunc()
		if delay <= 0 {
			return
		}
		node = t.tw.newNode(t.id, t.f, delay)
	} else {
		node = t.tw.startNode(t.id, t.f, t.d, t.once)
		node.nextFunc = t.nextFunc
	}
	t.node = node
	t.tw.submitAdd(node)
}

// Stop cancels the timer immediately: the running flag is cleared atomically, so
// no new callback starts after Stop returns. A remove command is queued to detach
// and reset the node promptly. A callback already executing at the moment Stop
// is called cannot be interrupted. Stop clears the node immediately, so a
// subsequent Start builds a fresh node.
func (t *Timer) Stop() {
	if t.node == nil {
		return
	}
	t.node.running.Store(false)
	t.tw.submitRemove(t.id)
	t.node = nil
}

// IsRunning reports whether the timer is active (added/started and not yet
// stopped or completed).
func (t *Timer) IsRunning() bool {
	if t.node == nil {
		return false
	}
	return t.node.running.Load()
}

// IsOnce reports whether the timer is one-shot (fires once, then finishes).
func (t *Timer) IsOnce() bool {
	return t.once
}

// SetNext sets or replaces the dynamic next-delay callback.
//
// For a recurring timer, fn is invoked by the driver after every fire to
// compute the delay until the next fire; a non-positive return stops the timer.
// The first fire always uses the timer's original delay d. Setting fn takes
// effect immediately: the already-queued next fire is re-armed.
//
// For a one-shot timer, fn computes the single-fire delay on the next Start;
// without one the timer fires at its original delay d. The timer still fires
// exactly once, then stops.
//
// nil clears dynamic mode, reverting to the timer's original behavior (fixed
// interval or one-shot). A schedule-driven timer (created via AddScheduleTimer)
// is a separate system and rejects any SetNext call: it is ignored with a
// warning. SetNext must be called from the timer's owner goroutine (the same
// goroutine that drives Start/Stop).
func (t *Timer) SetNext(fn func() time.Duration) {
	if t.schedule != nil {
		clog.Warnf("[Timer] SetNext ignored: schedule-driven timer. id=%d", t.id)
		return
	}

	t.nextFunc = fn
	if t.node != nil {
		t.tw.submitNext(t.id, fn)
	}
}

// Clear stops the timer and releases references for safe reuse.
func (t *Timer) Clear() {
	if t.node != nil {
		t.Stop()
	}
	t.node = nil
	t.f = nil
	t.nextFunc = nil
}
