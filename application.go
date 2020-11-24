package cherry

import (
	"github.com/cherry-game/cherry/cluster"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/utils"
	"os"
	"os/signal"
	"syscall"
)

// Application
type Application struct {
	nodeId     string                        // current node id
	nodeType   string                        // current node type
	startTime  int64                         // application start time
	running    bool                          // is running
	die        chan bool                     // wait for end application
	components []cherryInterfaces.IComponent // all components
}

func (a *Application) NodeId() string {
	return a.nodeId
}

func (a *Application) NodeType() string {
	return a.nodeType
}

func (a *Application) ThisNode() cherryInterfaces.INode {
	return cherryCluster.Nodes().Get(a.nodeType, a.nodeId)
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

// Startup
func (a *Application) Startup(components ...cherryInterfaces.IComponent) {
	if a.running {
		cherryLogger.Errorf("[nodeId = %s] application has running.", a.nodeId)
		return
	}

	//is running
	a.running = true

	cherryLogger.Infof("[nodeId = %s] application is starting.", a.nodeId)

	// add components & init
	for _, c := range components {
		if c == nil || c.Name() == "" {
			cherryLogger.Error("[component] is nil.")
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
		cherryLogger.Debugf("[component = %s] executed Init().", c.Name())
	}

	//execute AfterInit()
	for _, c := range a.components {
		c.AfterInit()
		cherryLogger.Debugf("[component = %s] executed AfterInit().", c.Name())
	}

	stringTime := cherryUtils.Timer.UnixTimeToString(a.startTime)

	cherryLogger.Infof("[nodeId = %s] application is running. startTime = %s", a.nodeId, stringTime)
	cherryLogger.Info("-----------------------------------------")
}

//
func (a *Application) Shutdown(beforeStopHook ...func()) {
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	case <-a.die:
	case <-sg:
		{
			//set running flag
			a.running = false

			cherryLogger.Infof("------- [nodeId = %s] application is stopping... -------", a.NodeId())

			if beforeStopHook != nil {
				for _, f := range beforeStopHook {
					f()
				}
			}

			//all components in reverse order
			for i := len(a.components) - 1; i >= 0; i-- {
				a.components[i].BeforeStop()
				cherryLogger.Debugf("[component = %s] executed BeforeStop().", a.components[i].Name())
			}

			for i := len(a.components) - 1; i >= 0; i-- {
				a.components[i].Stop()
				cherryLogger.Debugf("[component = %s] executed Stop().", a.components[i].Name())
			}

			cherryLogger.Infof("------- [nodeId = %s] application is stopped... -------", a.NodeId())

			close(a.die)
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
