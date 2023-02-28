package cherryMap

import (
	"fmt"
	"sync"
)

// Map 泛型map，可加锁
type Map[K comparable, V any] struct {
	mutex sync.RWMutex
	m     map[K]V
	safe  bool
}

func NewMap[K comparable, V any](safe ...bool) *Map[K, V] {
	mp := &Map[K, V]{
		m:    make(map[K]V),
		safe: false,
	}

	if len(safe) > 0 {
		mp.safe = safe[0]
	}

	return mp
}

func (p *Map[K, V]) Put(key K, value V) {
	if p.safe {
		p.mutex.Lock()
		defer p.mutex.Unlock()
	}

	p.m[key] = value
}

func (p *Map[K, V]) Get(key K) (V, bool) {
	if p.safe {
		p.mutex.RLock()
		defer p.mutex.RUnlock()
	}

	value, found := p.m[key]
	return value, found
}

func (p *Map[K, V]) Remove(key K) (V, bool) {
	if p.safe {
		p.mutex.Lock()
		defer p.mutex.Unlock()
	}

	v, found := p.m[key]
	if found {
		delete(p.m, key)
	}
	return v, found
}

func (p *Map[K, V]) Size() int {
	if p.safe {
		p.mutex.RLock()
		defer p.mutex.RUnlock()
	}

	return len(p.m)
}

func (p *Map[K, V]) Empty() bool {
	return p.Size() == 0
}

func (p *Map[K, V]) Keys() []K {
	keys := make([]K, p.Size())

	if p.safe {
		p.mutex.RLock()
		defer p.mutex.RUnlock()
	}

	count := 0
	for key := range p.m {
		keys[count] = key
		count++
	}
	return keys
}

func (p *Map[K, V]) Values() []V {
	values := make([]V, p.Size())

	if p.safe {
		p.mutex.RLock()
		defer p.mutex.RUnlock()
	}

	count := 0
	for _, value := range p.m {
		values[count] = value
		count++
	}
	return values
}

func (p *Map[K, V]) Clear() {
	if p.safe {
		p.mutex.Lock()
		p.mutex.Unlock()
	}

	p.m = make(map[K]V)
}

func (p *Map[K, V]) String() string {
	return fmt.Sprintf("%v, safe = %v", p.m, p.safe)
}
