package cherryQueue

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/interfaces"
)

//QueueComponent queue component
type QueueComponent struct {
	cherryInterfaces.BaseComponent
	Abc string
}

//NewQueue
func NewQueue() *QueueComponent {
	return &QueueComponent{
		Abc: "abc mocks string",
	}
}

func (s *QueueComponent) Name() string {
	return cherryConst.QueueComponent
}

func (s *QueueComponent) Init() {
}

func (s *QueueComponent) AfterInit() {
}

func (s *QueueComponent) BeforeStop() {
}

func (s *QueueComponent) Stop() {
}

func (s *QueueComponent) Test() string {
	return "call mocks()"
}
