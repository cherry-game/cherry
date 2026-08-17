package cherryTimeWheel

import "time"

// Timer is a user-facing handle with Start/Stop/Clear.
// It is NOT goroutine-safe: a Timer must be driven from a single goroutine (the
// owner that created it). The underlying *timerNode is owned by the wheel's driver
// goroutine and only mutated through the command channel.
type Timer struct {
	id   uint64        // unique ID
	d    time.Duration // bound interval
	f    func()        // bound callback
	once bool          // one-shot vs recurring
	node *timerNode    // non-nil when running; single-goroutine owner only
	tw   *TimeWheel    // owner
}

// ID returns the timer's unique ID.
func (t *Timer) ID() uint64 {
	return t.id
}

// Start restarts the timer, keeping its original type: a one-shot timer stays
// one-shot, a recurring timer stays recurring.
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
	node := t.tw.startNode(t.id, t.f, t.d, t.once)
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

// Clear stops the timer and releases references for safe reuse.
func (t *Timer) Clear() {
	if t.node != nil {
		t.Stop()
	}
	t.node = nil
	t.f = nil
}
