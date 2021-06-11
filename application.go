package cherry

import (
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/extend/utils"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"os"
	"os/signal"
	"syscall"
)

// Application
type Application struct {
	facade.INode                       // current node info
	startTime    cherryTime.CherryTime // application start time
	running      bool                  // is running
	die          chan bool             // wait for end application
	components   []facade.IComponent   // all components
}

func (a *Application) ThisNode() facade.INode {
	return a.INode
}

func (a *Application) Running() bool {
	return a.running
}

func (a *Application) Find(name string) facade.IComponent {
	if name == "" {
		return nil
	}

	for _, component := range a.components {
		if component.Name() == name {
			return component
		}
	}
	return nil
}

// Remove remove component by name
func (a *Application) Remove(name string) facade.IComponent {
	if name == "" {
		return nil
	}

	var removeComponent facade.IComponent
	for i := 0; i < len(a.components); i++ {
		if a.components[i].Name() == name {
			removeComponent = a.components[i]
			a.components = append(a.components[:i], a.components[i+1:]...)
			i--
		}
	}
	return removeComponent
}

func (a *Application) All() []facade.IComponent {
	return a.components
}

func (a *Application) StartTime() string {
	return a.startTime.ToDateTimeFormat()
}

// Startup
func (a *Application) OnStartup(components ...facade.IComponent) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Error(r)
		}
	}()

	if a.running {
		cherryLogger.Errorf("[nodeId = %s] application has running.", a.NodeId())
		return
	}

	//is running
	a.running = true

	cherryLogger.Info("-------------------------------------------------")
	cherryLogger.Infof("[nodeId      = %s] application is starting...", a.NodeId())
	cherryLogger.Infof("[pid         = %d]", os.Getpid())
	cherryLogger.Infof("[startTime   = %s]", a.StartTime())
	cherryLogger.Infof("[profile     = %s]", cherryProfile.Name())
	cherryLogger.Infof("[profileDir  = %s]", cherryProfile.Dir())
	cherryLogger.Infof("[profileFile = %s]", cherryProfile.FileName())
	cherryLogger.Infof("[debug       = %v]", cherryProfile.Debug())
	cherryLogger.Infof("[logLevel    = %s]", cherryLogger.DefaultLogger().Level)
	cherryLogger.Infof("[stackLevel	 = %s]", cherryLogger.DefaultLogger().StackLevel)
	cherryLogger.Infof("[writeFile   = %v]", cherryLogger.DefaultLogger().EnableWriteFile)
	cherryLogger.Info("-------------------------------------------------")

	// add components & init
	for _, c := range components {
		if c == nil || c.Name() == "" {
			cherryLogger.Errorf("[component] is nil. component=%T", c)
			return
		}

		result := a.Find(c.Name())
		if result != nil {
			cherryLogger.Errorf("[component name = %s] is duplicate.", c.Name())
			return
		}

		a.components = append(a.components, c)
		cherryLogger.Debugf("[component = %s] is added.", c.Name())
	}

	// execute Load()
	for _, c := range a.components {
		c.Set(a)
		c.Init()
		cherryLogger.Debugf("[component = %s] -> OnInit().", c.Name())
	}

	//execute OnAfterInit()
	for _, c := range a.components {
		c.OnAfterInit()
		cherryLogger.Debugf("[component = %s] -> OnAfterInit().", c.Name())
	}

	cherryLogger.Infof("[nodeId = %s] application is running.", a.NodeId())
	cherryLogger.Info("-------------------------------------------------")
}

func (a *Application) OnShutdown(beforeStopFn ...func()) {
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	//case <-a.die:
	case <-sg:
		{
			//set running flag
			a.running = false
			cherryLogger.Infof("------- [nodeId = %s] application is shutting... -------", a.NodeId())

			cherryUtils.Try(func() {
				if beforeStopFn != nil {
					for _, f := range beforeStopFn {
						f()
					}
				}
			}, func(errString string) {
				cherryLogger.Warnf("[beforeStopFn] error = %s", errString)
			})

			//all components in reverse order
			for i := len(a.components) - 1; i >= 0; i-- {
				cherryUtils.Try(func() {
					a.components[i].OnBeforeStop()
					cherryLogger.Infof("[component = %s] -> OnBeforeStop().", a.components[i].Name())
				}, func(errString string) {
					cherryLogger.Warnf("[component = %s] -> OnBeforeStop(). error = %s", a.components[i].Name(), errString)
				})
			}

			for i := len(a.components) - 1; i >= 0; i-- {
				cherryUtils.Try(func() {
					a.components[i].OnStop()
					cherryLogger.Infof("[component = %s] -> OnStop().", a.components[i].Name())
				}, func(errString string) {
					cherryLogger.Warnf("[component = %s] -> OnStop(). error = %s", a.components[i].Name(), errString)
				})
			}

			cherryLogger.Infof("------- [nodeId = %s] application is shutdown... -------", a.NodeId())
		}
	}
}
