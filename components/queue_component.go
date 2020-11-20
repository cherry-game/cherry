package cherryComponents

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
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
	cherryLogger.Infof("[component = %s] is executed Init() method.", s.Name())
}

func (s *QueueComponent) AfterInit() {
	cherryLogger.Infof("[component = %s] is executed AfterInit() method.", s.Name())
}

func (s *QueueComponent) Test() string {
	return "call mocks()"
}

func (s *QueueComponent) BeforeStop() {
	cherryLogger.Infof("[component = %s] is executed BeforeStop() method.", s.Name())
}

func (s *QueueComponent) Stop() {
	cherryLogger.Infof("[component = %s] is executed Stop() method.", s.Name())
}
