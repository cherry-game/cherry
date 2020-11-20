package cherryInterfaces

type IAppContext interface {
	Set(ctx IApplication)
	App() IApplication
}

type AppContext struct {
	ctx IApplication
}

func (b *AppContext) Set(ctx IApplication) {
	b.ctx = ctx
}

func (b *AppContext) App() IApplication {
	return b.ctx
}

type IApplication interface {
	//NodeId current nodeId
	NodeId() string

	//NodeType current nodeType
	NodeType() string

	//ThisNode current node info
	ThisNode() INode

	//Running is running
	Running() bool

	//Find find a component
	Find(name string) IComponent

	//Remove remove a component
	Remove(name string) IComponent

	//All all components
	All() []IComponent

	//PostEvent
	PostEvent(e IEvent)

	//Startup
	Startup(components ...IComponent)

	//Shutdown
	Shutdown(beforeStopHook ...func())
}
