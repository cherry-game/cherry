package cherryProto

import "sync"

// SessionData is the session key/values of one connection.
//
// It is the one part of a connection reached from several goroutines — a socket
// read goroutine routes on it while the connection's own actor writes it — so the
// lock that makes those accesses safe lives here, with the data. Every read and
// write goes through a method below; the map itself is never handed out.
//
// It sits next to Session because it is what a Session's Data is built from: a
// session handed to a handler carries a Clone, so the handler reads a consistent
// copy while the connection keeps writing to the live values.
//
// The zero value is ready to use.
type SessionData struct {
	lock sync.RWMutex
	data map[string]string
}

// Set writes one key. A blank key or value is ignored.
func (p *SessionData) Set(key, value string) {
	if key == "" || value == "" {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if p.data == nil {
		p.data = make(map[string]string)
	}

	p.data[key] = value
}

// Get reads one key, returning "" when it is not set.
func (p *SessionData) Get(key string) string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.data[key]
}

// Has reports whether the key is set.
func (p *SessionData) Has(key string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	_, found := p.data[key]
	return found
}

// Remove drops one key.
func (p *SessionData) Remove(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.data, key)
}

// Clone returns a copy of the key/values, taken in one go so the result is a
// consistent view rather than a mix of before and after some concurrent write.
//
// The copy is the caller's alone, so it can be read on any goroutine without
// locking. It is nil when there is nothing to copy.
func (p *SessionData) Clone() map[string]string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if len(p.data) < 1 {
		return nil
	}

	data := make(map[string]string, len(p.data))
	for k, v := range p.data {
		data[k] = v
	}

	return data
}
