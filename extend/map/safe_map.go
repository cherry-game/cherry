// file from https://github.com/beego/beego/blob/develop/core/utils/safemap.go

package cherrySafeMap

import (
	"sync"
)

// SafeMap is a map with lock
type SafeMap struct {
	lock *sync.RWMutex
	m    map[interface{}]interface{}
}

// NewSafeMap return new safemap
func NewSafeMap() *SafeMap {
	return &SafeMap{
		lock: new(sync.RWMutex),
		m:    make(map[interface{}]interface{}),
	}
}

// Get from maps return the k's value
func (m *SafeMap) Get(k interface{}) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if val, ok := m.m[k]; ok {
		return val
	}
	return nil
}

// Set Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *SafeMap) Set(k interface{}, v interface{}) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if val, ok := m.m[k]; !ok {
		m.m[k] = v
	} else if val != v {
		m.m[k] = v
	} else {
		return false
	}
	return true
}

// Check Returns true if k is exist in the map.
func (m *SafeMap) Check(k interface{}) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.m[k]
	return ok
}

// Delete the given key and value.
func (m *SafeMap) Delete(k interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.m, k)
}

// Items returns all items in safemap.
func (m *SafeMap) Items() map[interface{}]interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	r := make(map[interface{}]interface{})
	for k, v := range m.m {
		r[k] = v
	}
	return r
}

// Count returns the number of items within the map.
func (m *SafeMap) Count() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.m)
}
