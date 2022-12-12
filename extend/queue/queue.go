package cherryQueue

import (
	"container/list"
	"sync"
)

type Queue struct {
	mutex sync.RWMutex
	list  *list.List
}

func NewQueue() *Queue {
	return &Queue{
		mutex: sync.RWMutex{},
		list:  list.New(),
	}
}

func (p *Queue) Pop() (interface{}, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Empty() {
		return nil, false
	}

	element := p.list.Front()
	if element == nil {
		return nil, false
	}

	p.list.Remove(element)
	return element.Value, true
}

func (p *Queue) Push(val interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.list.PushBack(val)
}

func (p *Queue) Len() int {
	return p.list.Len()
}

func (p *Queue) Empty() bool {
	return p.list.Len() == 0
}

func (p *Queue) Clear() {
	p.mutex.Lock()
	p.list.Init()
	p.mutex.Unlock()
}
