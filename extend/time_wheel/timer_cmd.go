package cherryTimeWheel

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
		n.running.Store(false)
		tw.detachNode(n)
		tw.reset(n)
		tw.activeNum.Add(-1)
	}
}
