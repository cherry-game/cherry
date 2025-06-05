package cherryTimeWheel

import (
	"sync"
	"sync/atomic"
	"time"
)

// truncate returns the result of rounding x toward zero to a multiple of m.
// If m <= 0, Truncate returns x unchanged.
func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}

// TimeToMS returns an integer number, which represents t in milliseconds.
func TimeToMS(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// MSToTime returns the UTC time corresponding to the given Unix time,
// t milliseconds since January 1, 1970 UTC.
func MSToTime(t int64) time.Time {
	return time.Unix(0, t*int64(time.Millisecond))
}

type waitGroupWrapper struct {
	sync.WaitGroup
}

func (w *waitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

var _nextID uint64

func NextID() uint64 {
	return atomic.AddUint64(&_nextID, 1)
}
