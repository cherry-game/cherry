package cherryInterfaces

//IComponent
type IComponent interface {
	IAppContext

	//Name unique components name
	Name() string

	Init()

	AfterInit()

	BeforeStop()

	Stop()
}

// BaseComponent
type BaseComponent struct {
	AppContext
}

func (*BaseComponent) Name() string {
	panic("implement `Name()` please.")
}

func (*BaseComponent) Init() {
}

func (*BaseComponent) AfterInit() {
}

func (*BaseComponent) BeforeStop() {
}

func (*BaseComponent) Stop() {
}
