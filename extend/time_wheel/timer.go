package cherryTimeWheel

import "time"

// Timer is a user-facing handle with Start/Stop/Clear.
// It is NOT goroutine-safe: a Timer must be driven from a single goroutine (the
// owner that created it). The underlying *timerNode is owned by the wheel's driver
// goroutine and only mutated through the command channel.
type Timer struct {
	id   uint64        // unique ID
	tw   *TimeWheel    // owner
	d    time.Duration // bound interval
	f    func()        // bound callback
	node *timerNode    // non-nil when running; single-goroutine owner only
}

// ID returns the timer's unique ID.
func (t *Timer) ID() uint64 {
	return t.id
}

// Start starts (or restarts) the timer using the bound interval d and callback f.
// After Start the timer keeps firing every d until Stop/Clear is called.
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
	t.tw.startNode(t, true)
}

// Stop cancels the timer. Asynchronous: the removal is queued and processed by the
// driver on a later cycle, so the callback may fire once more after Stop returns.
// Stop clears the node immediately, so a subsequent Start builds a fresh node.
func (t *Timer) Stop() {
	if t.node == nil {
		return
	}
	t.tw.submitRemove(t.id)
	t.node = nil
}

// Clear stops the timer and releases references for safe reuse.
func (t *Timer) Clear() {
	if t.node != nil {
		t.Stop()
	}
	t.node = nil
	t.f = nil
}
