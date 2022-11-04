package cherryCron

import (
	"github.com/cherry-game/cherry"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/robfig/cron/v3"
)

const (
	Name = "cron_component"
)

type Component struct {
	cfacade.Component
}

// Name unique components name
func (*Component) Name() string {
	return Name
}

func (p *Component) Init() {
	Start()
	clog.Info("cron component init.")
}

func (p *Component) OnStop() {
	Stop()
	clog.Infof("cron component is stopped.")
}

func RegisterComponent(opts ...cron.Option) {
	if len(opts) > 0 {
		Init(opts...)
	}

	cherry.RegisterComponent(&Component{})
}
