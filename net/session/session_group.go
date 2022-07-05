package cherrySession

import (
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"go.uber.org/zap/zapcore"
	"sync"
	"sync/atomic"
)

const (
	groupStatusWorking = 0
	groupStatusClosed  = 1
)

// SessionFilter represents a filter which was used to filter session when Multicast,
// the session will receive the message while filter returns true.
type SessionFilter func(*Session) bool

// Group represents a session group which used to manage a number of
// sessions, data send to the group will send to all session in it.
type Group struct {
	mu       sync.RWMutex
	status   int32                    // channel current status
	name     string                   // channel name
	sessions map[cfacade.SID]*Session // session id map to session instance
}

// NewGroup returns a new group instance
func NewGroup(n string) *Group {
	return &Group{
		status:   groupStatusWorking,
		name:     n,
		sessions: make(map[cfacade.SID]*Session),
	}
}

// Member returns specified UID's session
func (c *Group) Member(uid int64) (*Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, s := range c.sessions {
		if s.UID() == uid {
			return s, nil
		}
	}

	return nil, cerr.SessionMemberNotFound
}

// Members returns all member's UID in current group
func (c *Group) Members() []int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var members []int64
	for _, s := range c.sessions {
		members = append(members, s.UID())
	}

	return members
}

// Multicast  push  the message to the filtered clients
func (c *Group) Multicast(route string, v interface{}, filter SessionFilter) error {
	if c.isClosed() {
		return cerr.SessionClosedGroup
	}

	if clog.LogLevel(zapcore.DebugLevel) {
		clog.Debugf("multicast. [route = %s, data = %+v]", route, v)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, s := range c.sessions {
		if !filter(s) {
			continue
		}
		s.Push(route, v)
	}

	return nil
}

// Broadcast push  the message(s) to  all members
func (c *Group) Broadcast(route string, v interface{}) error {
	if c.isClosed() {
		return cerr.SessionClosedGroup
	}

	if clog.LogLevel(zapcore.DebugLevel) {
		clog.Debugf("broadcast. [route = %s, data = %+v]", route, v)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, s := range c.sessions {
		s.Push(route, v)
	}

	return nil
}

// Contains check whether a UID is contained in current group or not
func (c *Group) Contains(uid int64) bool {
	_, err := c.Member(uid)
	return err == nil
}

// Add add session to group
func (c *Group) Add(session *Session) error {
	if c.isClosed() {
		return cerr.SessionClosedGroup
	}

	if clog.LogLevel(zapcore.DebugLevel) {
		session.Debugf("add session to [group = %s, SID = %v, UID = %d]", c.name, session.SID(), session.UID())
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	id := session.sid
	_, ok := c.sessions[session.sid]
	if ok {
		return cerr.SessionDuplication
	}

	c.sessions[id] = session
	return nil
}

// Leave remove specified UID related session from group
func (c *Group) Leave(s *Session) error {
	if c.isClosed() {
		return cerr.SessionClosedGroup
	}

	if clog.LogLevel(zapcore.DebugLevel) {
		s.Debugf("remove session from [group = %s, UID = %d]", c.name, s.UID())
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.sessions, s.sid)
	return nil
}

// LeaveAll clear all sessions in the group
func (c *Group) LeaveAll() error {
	if c.isClosed() {
		return cerr.SessionClosedGroup
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.sessions = make(map[cfacade.SID]*Session)
	return nil
}

// Count get current member amount in the group
func (c *Group) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.sessions)
}

func (c *Group) isClosed() bool {
	if atomic.LoadInt32(&c.status) == groupStatusClosed {
		return true
	}
	return false
}

// Close destroy group, which will release all resource in the group
func (c *Group) Close() error {
	if c.isClosed() {
		return cerr.SessionClosedGroup
	}

	atomic.StoreInt32(&c.status, groupStatusClosed)

	// release all reference
	c.sessions = make(map[cfacade.SID]*Session)
	return nil
}
