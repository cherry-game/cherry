// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines the component system:
//   - IComponent: named pluggable unit with lifecycle hooks
//   - IComponentLifecycle: Set → Init → OnAfterInit (startup), OnBeforeStop → OnStop (shutdown)
//   - Component: embeddable base with no-op lifecycle stubs
package cherryFacade

type (
	// IComponent is a named, lifecycle-aware unit registered with the Application.
	// Components are initialized in registration order and stopped in reverse order.
	IComponent interface {
		Name() string            // unique name within the Application
		App() IApplication       // the owning Application
		IComponentLifecycle
	}

	// IComponentLifecycle defines the startup and shutdown hooks for a component.
	// Startup: Set(app) → Init() → OnAfterInit()
	// Shutdown: OnBeforeStop() → OnStop()
	IComponentLifecycle interface {
		Set(app IApplication) // receive the owning Application reference
		Init()                // initialise internal state
		OnAfterInit()         // called after ALL components have been Init'd
		OnBeforeStop()        // called before shutdown begins
		OnStop()              // release resources
	}
)

// Component is an embeddable base that provides no-op lifecycle methods.
// Embed this in your component struct to satisfy IComponent without
// implementing every hook.
type Component struct {
	app IApplication
}

// Name returns the component's unique name.
// Override this to return a non-empty string; the base implementation returns "".
func (*Component) Name() string {
	return ""
}

// App returns the Application this component is registered with.
// Available after Set() is called during startup.
func (p *Component) App() IApplication {
	return p.app
}

// Set stores the Application reference. Called by the framework once
// before Init(), in registration order.
func (p *Component) Set(app IApplication) {
	p.app = app
}

// Init is called after Set(), before any other component is initialized.
// Override to allocate resources that other components may depend on.
func (*Component) Init() {
}

// OnAfterInit is called after ALL components have been Init'd.
// Override to access other components via p.App().Find(...).
func (*Component) OnAfterInit() {
}

// OnBeforeStop is called before shutdown begins.
// Override to close connections or flush pending writes.
func (*Component) OnBeforeStop() {
}

// OnStop is the final shutdown hook. Called after all components'
// OnBeforeStop have run. Override to release resources.
func (*Component) OnStop() {
}
