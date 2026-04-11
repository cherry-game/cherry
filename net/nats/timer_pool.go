package cherryNats

import (
	"sync"
	"time"
)

var _timerPool = sync.Pool{
	New: func() any {
		return time.NewTimer(0)
	},
}

func acquireTimer(d time.Duration) *time.Timer {
	t := _timerPool.Get().(*time.Timer)
	t.Reset(d)
	return t
}

func releaseTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	_timerPool.Put(t)
}
