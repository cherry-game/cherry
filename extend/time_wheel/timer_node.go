package cherryTimeWheel

import "time"

// timerNode is an internal linked-list node. Only accessed by the driver goroutine.
type timerNode struct {
	id       uint64        // unique timer ID, key of nodeMap
	expire   int64         // expiry tick (absolute tick count)
	next     *timerNode    // next node in the same slot list
	prev     *timerNode    // previous node in the same slot list
	level    int8          // level: -1 for near, 0~3 for t[0..3]
	index    int32         // slot index: 0~255 for near, 0~63 for levels
	cb       func()        // business callback, invoked on expiry
	interval time.Duration // recurring interval (0 for one-shot)
	schedule Scheduler     // AddSchedule scheduler; nil = fixed interval or one-shot
}

// listInsert inserts node at the head of the slot list pointed to by head.
func listInsert(head **timerNode, node *timerNode) {
	node.next = *head
	node.prev = nil
	if *head != nil {
		(*head).prev = node
	}
	*head = node
}

// detachList detaches and returns the whole slot list.
func detachList(head **timerNode) *timerNode {
	h := *head
	*head = nil
	return h
}
