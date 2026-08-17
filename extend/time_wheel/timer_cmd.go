package cherryTimeWheel

import "time"

// timerCmd is a command submitted to the driver goroutine through the MPSC channel.
type timerCmd interface {
	exec(tw *TimeWheel)
}

// addCmd inserts a node into the wheel.
type addCmd struct {
	node *timerNode
}

func (c *addCmd) exec(tw *TimeWheel) {
	tw.handleAddNode(c.node)
}

// removeCmd detaches and resets the node with the given id.
type removeCmd struct {
	id uint64
}

func (c *removeCmd) exec(tw *TimeWheel) {
	if n, ok := tw.nodeMap[c.id]; ok {
		tw.detachNode(n)
		tw.stopNode(n)
		tw.activeNum.Add(-1)
	}
}

// nextCmd updates the dynamic next-delay callback of an active node.
// No-op if the node is no longer in the wheel (stopped / already fired).
type nextCmd struct {
	id uint64
	fn func() time.Duration
}

func (c *nextCmd) exec(tw *TimeWheel) {
	// schedule-driven nodes reject nextFunc (SetNext is an interval supplement).
	// SetNext only records the callback; fn is invoked by dispatchList after the
	// next fire, so the already-queued first fire keeps its original delay d.
	n, ok := tw.nodeMap[c.id]
	if !ok || n.schedule != nil {
		return
	}
	n.nextFunc = c.fn
}
