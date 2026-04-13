package cherryNats

import (
	"sync"
	"time"
)

var _timerPool = sync.Pool{
	New: func() any {
		return time.NewTimer(time.Hour)
	},
}

func acquireTimer(d time.Duration) *time.Timer {
	t := _timerPool.Get().(*time.Timer)
	resetTimer(t)
	t.Reset(d)
	return t
}

func releaseTimer(t *time.Timer) {
	resetTimer(t)
	_timerPool.Put(t)
}

func resetTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
