package cherryMini

import (
	cc "github.com/cherry-game/cherry/const"
	cf "github.com/cherry-game/cherry/facade"
)

type (
	appFunc func(cf.IApplication)
	Options func(options *options)

	options struct {
		initFunc       appFunc
		afterInitFunc  appFunc
		beforeStopFunc appFunc
		stopFunc       appFunc
	}

	// Component mini wrapper component
	Component struct {
		cf.Component
		name string
		options
	}
)

func New(name string, opts ...Options) *Component {
	comp := &Component{
		name: name,
	}

	// fill options
	for _, opt := range opts {
		opt(&comp.options)
	}

	return comp
}

func (p *Component) Name() string {
	return cc.MiniComponentPrefix + p.name
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

func WithInitFunc(fn func(cf.IApplication)) Options {
	return func(options *options) {
		options.initFunc = fn
	}
}

func WithAfterInit(fn func(cf.IApplication)) Options {
	return func(options *options) {
		options.afterInitFunc = fn
	}
}

func WithBeforeStop(fn func(cf.IApplication)) Options {
	return func(options *options) {
		options.beforeStopFunc = fn
	}
}

func WithStop(fn func(cf.IApplication)) Options {
	return func(options *options) {
		options.stopFunc = fn
	}
}
