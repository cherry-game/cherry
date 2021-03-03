package cherry

import (
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"os"
	"os/signal"
	"syscall"
)

// Application
type Application struct {
	cherryInterfaces.INode                               // current node info
	startTime              cherryTime.CherryTime         // application start time
	running                bool                          // is running
	die                    chan bool                     // wait for end application
	components             []cherryInterfaces.IComponent // all components
}

func (a *Application) ThisNode() cherryInterfaces.INode {
	return a.INode
}

func (a *Application) Running() bool {
	return a.running
}

func (a *Application) Find(name string) cherryInterfaces.IComponent {
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
func (a *Application) Remove(name string) cherryInterfaces.IComponent {
	if name == "" {
		return nil
	}

	var removeComponent cherryInterfaces.IComponent
	for i := 0; i < len(a.components); i++ {
		if a.components[i].Name() == name {
			removeComponent = a.components[i]
			a.components = append(a.components[:i], a.components[i+1:]...)
			i--
		}
	}
	return removeComponent
}

func (a *Application) All() []cherryInterfaces.IComponent {
	return a.components
}

func (a *Application) StartTime() string {
	return a.startTime.ToDateTimeFormat()
}

// Startup
func (a *Application) Startup(components ...cherryInterfaces.IComponent) {
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
	cherryLogger.Infof("[nodeId 	= %s] application is starting...", a.NodeId())
	cherryLogger.Infof("[profile 	= %s]", cherryProfile.Name())
	cherryLogger.Infof("[configDir 	= %s]", cherryProfile.Dir())
	cherryLogger.Infof("[configFile = %s]", cherryProfile.FileName())
	cherryLogger.Infof("[debug 		= %v]", cherryProfile.Debug())
	cherryLogger.Infof("[startTime 	= %s]", a.StartTime())
	cherryLogger.Infof("[pid	 	= %d]", os.Getpid())
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

	// execute Init()
	for _, c := range a.components {
		c.Set(a)
		c.Init()
		cherryLogger.Debugf("[component = %s] -> Init().", c.Name())
	}

	//execute AfterInit()
	for _, c := range a.components {
		c.AfterInit()
		cherryLogger.Debugf("[component = %s] -> AfterInit().", c.Name())
	}

	cherryLogger.Infof("[nodeId = %s] application is running. startTime = %s", a.NodeId(), a.StartTime())
	cherryLogger.Info("-------------------------------------------------")
}

func (a *Application) Shutdown(beforeStopFn ...func()) {
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	//case <-a.die:
	case <-sg:
		{
			//set running flag
			a.running = false
			cherryLogger.Infof("------- [nodeId = %s] application is shutting... -------", a.NodeId())

			if beforeStopFn != nil {
				for _, f := range beforeStopFn {
					f()
				}
			}

			//all components in reverse order
			for i := len(a.components) - 1; i >= 0; i-- {
				cherryUtils.Try(func() {
					a.components[i].BeforeStop()
					cherryLogger.Debugf("[component = %s] -> BeforeStop().", a.components[i].Name())
				}, func(errString string) {
					cherryLogger.Debugf("[component = %s] -> BeforeStop(). error = %s", a.components[i].Name(), errString)
				})
			}

			for i := len(a.components) - 1; i >= 0; i-- {
				cherryUtils.Try(func() {
					a.components[i].Stop()
					cherryLogger.Debugf("[component = %s] -> Stop().", a.components[i].Name())
				}, func(errString string) {
					cherryLogger.Debugf("[component = %s] -> Stop(). error = %s", a.components[i].Name(), errString)
				})
			}

			cherryLogger.Infof("------- [nodeId = %s] application is shutdown... -------", a.NodeId())
		}
	}
}

// filter(filter IHandlerfilter)  before after  这个过滤器可以放到Acceptor里去
// globalFilter()  全局过滤器，也可以放Acceptor里去
// rpcBefore()  rpc过滤器
// addCrons
// removeCrons
// rpc UserRpc
// sysrpc SysRpc
