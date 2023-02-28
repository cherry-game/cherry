package cherryFacade

type (
	IComponent interface {
		Name() string
		App() IApplication
		IComponentLifecycle
	}

	IComponentLifecycle interface {
		Set(app IApplication)
		Init()
		OnAfterInit()
		OnBeforeStop()
		OnStop()
	}
)

// Component base component
type Component struct {
	app IApplication
}

func (*Component) Name() string {
	return ""
}

func (p *Component) App() IApplication {
	return p.app
}

func (p *Component) Set(app IApplication) {
	p.app = app
}

func (*Component) Init() {
}

func (*Component) OnAfterInit() {
}

func (*Component) OnBeforeStop() {
}

func (*Component) OnStop() {
}
