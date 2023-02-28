package cherryGops

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"github.com/google/gops/agent"
)

// Component gops 监听进程数据
type Component struct {
	cherryFacade.Component
	options agent.Options
}

func New(options ...agent.Options) *Component {
	component := &Component{}
	if len(options) > 0 {
		component.options = options[0]
	}
	return component
}

func (c *Component) Name() string {
	return "gops_component"
}

func (c *Component) Init() {
	if err := agent.Listen(c.options); err != nil {
		cherryLogger.Error(err)
	}
}

func (c *Component) OnAfterInit() {
}

func (c *Component) OnStop() {
	//agent.Close()
}
