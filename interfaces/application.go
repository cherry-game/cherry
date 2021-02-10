package cherryInterfaces

type IAppContext interface {
	Set(app IApplication)
	App() IApplication
}

type AppContext struct {
	app IApplication
}

func (b *AppContext) Set(app IApplication) {
	b.app = app
}

func (b *AppContext) App() IApplication {
	return b.app
}

type IApplication interface {
	INode                              // current node info
	Running() bool                     // is running
	Find(name string) IComponent       // find a component
	Remove(name string) IComponent     // remove a component
	All() []IComponent                 // all components
	Startup(components ...IComponent)  // startup
	Shutdown(beforeStopHook ...func()) // shutdown
}
