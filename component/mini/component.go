package cherryMini

import (
	cc "github.com/cherry-game/cherry/const"
	cf "github.com/cherry-game/cherry/facade"
)

// Component mini wrapper component
type Component struct {
	cf.Component
	name           string
	initFunc       func(cf.IApplication)
	afterInitFunc  func(cf.IApplication)
	beforeStopFunc func(cf.IApplication)
	stopFunc       func(cf.IApplication)
}

func New(name string) *Component {
	return &Component{
		name: name,
	}
}

func (p *Component) Name() string {
	return cc.MiniComponent + p.name
}

func (p *Component) Init() {
	if p.initFunc != nil {
		p.initFunc(p.App())
	}
}

func (p *Component) OnAfterInit() {
	if p.afterInitFunc != nil {
		p.afterInitFunc(p.App())
	}
}

func (p *Component) OnBeforeStop() {
	if p.beforeStopFunc != nil {
		p.beforeStopFunc(p.App())
	}
}

func (p *Component) OnStop() {
	if p.stopFunc != nil {
		p.stopFunc(p.App())
	}
}

func (p *Component) SetInit(fn func(cf.IApplication)) *Component {
	p.initFunc = fn
	return p
}

func (p *Component) SetAfterInit(fn func(cf.IApplication)) *Component {
	p.afterInitFunc = fn
	return p
}

func (p *Component) SetBeforeStop(fn func(cf.IApplication)) *Component {
	p.beforeStopFunc = fn
	return p
}

func (p *Component) SetStop(fn func(cf.IApplication)) *Component {
	p.stopFunc = fn
	return p
}
