package cherryTimeWheel

import (
	"sync/atomic"
	"time"
)

// timerNode is an internal linked-list node. Only accessed by the driver goroutine,
// except running which is atomic and also written by the owning Timer (Stop).
type timerNode struct {
	id       uint64               // unique timer ID, key of nodeMap
	expire   int64                // expiry tick (absolute tick count)
	next     *timerNode           // next node in the same slot list
	prev     *timerNode           // previous node in the same slot list
	level    int8                 // level: -1 for near, 0~3 for t[0..3]
	index    int32                // slot index: 0~255 for near, 0~63 for levels
	cb       func()               // business callback, invoked on expiry
	interval time.Duration        // recurring interval (0 for one-shot)
	schedule Scheduler            // AddScheduleTimer scheduler; nil = fixed interval or one-shot
	nextFunc func() time.Duration // dynamic next-delay (SetNext); nil = disabled
	running  atomic.Bool          // 1:1 with its Timer; driver checks before dispatch, owner reads via IsRunning
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
