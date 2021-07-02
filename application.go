package cherry

import (
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/extend/utils"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"syscall"
)

// Application
type Application struct {
	facade.INode
	facade.ISerializer
	facade.IPacketCodec
	startTime    cherryTime.CherryTime // application start time
	running      int32                 // is running
	die          chan bool             // wait for end application
	components   []facade.IComponent   // all components
	onShutdownFn []func()              // on shutdown execute functions
}

func (a *Application) Running() bool {
	return a.running > 0
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
func (a *Application) Startup(components ...facade.IComponent) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Error(r)
		}
	}()

	if a.Running() {
		cherryLogger.Errorf("[nodeId = %s] application has running.", a.NodeId())
		return
	}

	defer func() {
		cherryLogger.Flush()
	}()

	// is running
	atomic.AddInt32(&a.running, 1)

	cherryLogger.Info("-------------------------------------------------")
	cherryLogger.Infof("[nodeId      = %s] application is starting...", a.NodeId())
	cherryLogger.Infof("[pid         = %d]", os.Getpid())
	cherryLogger.Infof("[startTime   = %s]", a.StartTime())
	cherryLogger.Infof("[profile     = %s]", cherryProfile.Name())
	cherryLogger.Infof("[profileDir  = %s]", cherryProfile.Dir())
	cherryLogger.Infof("[profileFile = %s]", cherryProfile.FileName())
	cherryLogger.Infof("[debug       = %v]", cherryProfile.Debug())
	cherryLogger.Infof("[logLevel    = %s]", cherryLogger.DefaultLogger.Level)
	cherryLogger.Infof("[stackLevel  = %s]", cherryLogger.DefaultLogger.StackLevel)
	cherryLogger.Infof("[writeFile   = %v]", cherryLogger.DefaultLogger.EnableWriteFile)
	cherryLogger.Infof("[codec       = %v]", reflect.TypeOf(a.IPacketCodec))
	cherryLogger.Infof("[serializer  = %s]", a.ISerializer.Name())
	cherryLogger.Info("-------------------------------------------------")

	// add components & init
	for _, c := range components {
		if c == nil || c.Name() == "" {
			cherryLogger.Errorf("[component = %T] name is nil", c)
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
		cherryLogger.Debugf("[component = %s] -> OnInit().", c.Name())
		c.Init()
	}

	//execute OnAfterInit()
	for _, c := range a.components {
		cherryLogger.Debugf("[component = %s] -> OnAfterInit().", c.Name())
		c.OnAfterInit()
	}

	cherryLogger.Infof("[nodeId = %s] application is running.", a.NodeId())
	cherryLogger.Info("-------------------------------------------------")

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	case <-a.die:
		cherryLogger.Infof("[nodeId = %s] -> invoke shutdown().", a.NodeId())
	case <-sg:
		cherryLogger.Infof("[nodeId = %s] -> receive shutdown signal = %v.", a.NodeId(), sg)
	}

	// stop status
	atomic.StoreInt32(&a.running, 0)

	cherryLogger.Infof("------- [nodeId = %s] application will shutdown -------", a.NodeId())

	cherryUtils.Try(func() {
		if a.onShutdownFn != nil {
			for _, f := range a.onShutdownFn {
				f()
			}
		}

	}, func(errString string) {
		cherryLogger.Warnf("[onShutdownFn] error = %s", errString)
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

	cherryLogger.Infof("------- [nodeId = %s] application has been shutdown... -------", a.NodeId())
}

func (a *Application) OnShutdown(fn ...func()) {
	a.onShutdownFn = append(a.onShutdownFn, fn...)
}

func (a *Application) Shutdown() {
	a.die <- true
}

func (a *Application) SetSerializer(serializer facade.ISerializer) {
	a.ISerializer = serializer
}

func (a *Application) SetPacketCodec(codec facade.IPacketCodec) {
	a.IPacketCodec = codec
}
