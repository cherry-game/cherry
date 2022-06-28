// Package cherrySync cherrySync file from https://github.com/beego/beego/blob/develop/core/utils/safemap.go
package cherrySync

import cerr "github.com/cherry-game/cherry/error"

var (
	// errReturn indicates that the more than borrowed elements were returned.
	errReturn   = cerr.Error("discarding limited token, resource pool is full, someone returned multiple times")
	placeholder placeholderType
)

type (
	placeholderType = struct{}

	// Limit controls the concurrent requests.
	Limit struct {
		pool chan placeholderType
	}
)

// NewLimit creates a Limit that can borrow n elements from it concurrently.
func NewLimit(n int) Limit {
	return Limit{
		pool: make(chan placeholderType, n),
	}
}

// Borrow borrows an element from Limit in blocking mode.
func (l Limit) Borrow() {
	l.pool <- placeholder
}

// Return returns the borrowed resource, returns error only if returned more than borrowed.
func (l Limit) Return() error {
	select {
	case <-l.pool:
		return nil
	default:
		return errReturn
	}
}

// TryBorrow tries to borrow an element from Limit, in non-blocking mode.
// If success, true returned, false for otherwise.
func (l Limit) TryBorrow() bool {
	select {
	case l.pool <- placeholder:
		return true
	default:
		return false
	}
}
