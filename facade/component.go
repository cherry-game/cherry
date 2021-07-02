package cherryFacade

type IComponent interface {
	IAppContext
	Name() string
	Init()
	OnAfterInit()
	OnBeforeStop()
	OnStop()
}

// Component
type Component struct {
	AppContext
}

func (*Component) Name() string {
	return ""
}

func (*Component) Init() {
}

func (*Component) OnAfterInit() {
}

func (*Component) OnBeforeStop() {
}

func (*Component) OnStop() {
}
