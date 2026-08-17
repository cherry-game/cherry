package cherryTimeWheel

import "sync/atomic"

var _nextID atomic.Uint64

// nextID returns a globally unique, monotonically increasing ID.
func nextID() uint64 {
	return _nextID.Add(1)
}
