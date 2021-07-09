package cherryFacade

type IComponent interface {
	IAppContext
	Name() string
	Init()
	OnAfterInit()
	OnBeforeStop()
	OnStop()
}

// Component base component
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
