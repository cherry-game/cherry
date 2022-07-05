package cherry

import (
	"fmt"
	cconst "github.com/cherry-game/cherry/const"
	ctime "github.com/cherry-game/cherry/extend/time"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cpacket "github.com/cherry-game/cherry/net/packet"
	cserializer "github.com/cherry-game/cherry/net/serializer"
	cprofile "github.com/cherry-game/cherry/profile"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"syscall"
)

const (
	Cluster    NodeMode = 1 // 集群模式
	Standalone NodeMode = 2 // 单机模式
)

type (
	NodeMode byte

	Application struct {
		cfacade.INode
		cfacade.ISerializer
		cfacade.IPacketCodec
		isFrontend   bool
		nodeMode     NodeMode
		startTime    ctime.CherryTime     // application start time
		running      int32                // is running
		dieChan      chan bool            // wait for end application
		components   []cfacade.IComponent // all components
		onShutdownFn []func()             // on shutdown execute functions
	}
)

// NewApp create new application instance
func NewApp(profilePath, profileName, nodeId string) *Application {
	if err := cprofile.Init(profilePath, profileName); err != nil {
		panic(fmt.Sprintf("init profile fail. error = %s", err))
	}

	node, err := cprofile.LoadNode(nodeId)
	if err != nil {
		panic(err)
	}

	// set logger
	clog.SetNodeLogger(node)

	// print version info
	clog.Info(cconst.GetLOGO())

	app := &Application{
		INode:        node,
		ISerializer:  cserializer.NewProtobuf(),
		IPacketCodec: cpacket.NewPomeloCodec(),
		isFrontend:   true,
		nodeMode:     Standalone,
		startTime:    ctime.Now(),
		running:      0,
		dieChan:      make(chan bool),
	}

	return app
}

func (a *Application) IsFrontend() bool {
	return a.isFrontend
}

func (a *Application) NodeMode() NodeMode {
	return a.nodeMode
}

func (a *Application) Running() bool {
	return a.running > 0
}

func (a *Application) DieChan() chan bool {
	return a.dieChan
}

func (a *Application) Register(components ...cfacade.IComponent) {
	if a.Running() {
		return
	}

	for _, c := range components {
		if c == nil || c.Name() == "" {
			clog.Errorf("[component = %T] name is nil", c)
			return
		}

		result := a.Find(c.Name())
		if result != nil {
			clog.Errorf("[component name = %s] is duplicate.", c.Name())
			return
		}

		a.components = append(a.components, c)
	}
}

func (a *Application) Find(name string) cfacade.IComponent {
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
func (a *Application) Remove(name string) cfacade.IComponent {
	if name == "" {
		return nil
	}

	var removeComponent cfacade.IComponent
	for i := 0; i < len(a.components); i++ {
		if a.components[i].Name() == name {
			removeComponent = a.components[i]
			a.components = append(a.components[:i], a.components[i+1:]...)
			i--
		}
	}
	return removeComponent
}

func (a *Application) All() []cfacade.IComponent {
	return a.components
}

func (a *Application) StartTime() string {
	return a.startTime.ToDateTimeFormat()
}

// Startup load components before startup
func (a *Application) Startup(components ...cfacade.IComponent) {
	defer func() {
		if r := recover(); r != nil {
			clog.Error(r)
		}
	}()

	if a.Running() {
		clog.Error("application has running.")
		return
	}

	defer func() {
		clog.Flush()
	}()

	clog.Info("-------------------------------------------------")
	clog.Infof("[nodeId			= %s] application is starting...", a.NodeId())
	clog.Infof("[nodeType		= %s]", a.NodeType())
	clog.Infof("[pid			= %d]", os.Getpid())
	clog.Infof("[startTime		= %s]", a.StartTime())
	clog.Infof("[profile		= %s]", cprofile.Name())
	clog.Infof("[profilePath 	= %s]", cprofile.Dir())
	clog.Infof("[profileFile 	= %s]", cprofile.FileName())
	clog.Infof("[cherryLogLevel	= %s]", cprofile.LogLevel())
	clog.Infof("[isDebug		= %v]", cprofile.Debug())
	clog.Infof("[logLevel		= %s]", clog.DefaultLogger.Level)
	clog.Infof("[stackLevel		= %s]", clog.DefaultLogger.StackLevel)
	clog.Infof("[writeFile		= %v]", clog.DefaultLogger.EnableWriteFile)
	clog.Infof("[codec			= %v]", reflect.TypeOf(a.IPacketCodec))
	clog.Infof("[serializer		= %s]", a.ISerializer.Name())
	clog.Info("-------------------------------------------------")

	// add components
	a.Register(components...)

	// component list
	for _, c := range a.components {
		c.Set(a)
		clog.Infof("[component = %s] is added.", c.Name())
	}
	clog.Info("-------------------------------------------------")

	// execute Init()
	for _, c := range a.components {
		clog.Infof("[component = %s] -> OnInit().", c.Name())
		c.Init()
	}
	clog.Info("-------------------------------------------------")

	//execute OnAfterInit()
	for _, c := range a.components {
		clog.Infof("[component = %s] -> OnAfterInit().", c.Name())
		c.OnAfterInit()
	}
	clog.Info("-------------------------------------------------")
	spendTime := a.startTime.DiffInMillisecond(ctime.Now())
	clog.Infof("[spend time = %dms] application is running.", spendTime)
	clog.Info("-------------------------------------------------")

	// set application is running
	atomic.AddInt32(&a.running, 1)

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	case <-a.dieChan:
		clog.Info("invoke shutdown().")
	case s := <-sg:
		clog.Infof("receive shutdown signal = %v.", s)
	}

	// stop status
	atomic.StoreInt32(&a.running, 0)

	clog.Info("------- application will shutdown -------")

	cutils.Try(func() {
		if a.onShutdownFn != nil {
			for _, f := range a.onShutdownFn {
				f()
			}
		}

	}, func(errString string) {
		clog.Warnf("[onShutdownFn] error = %s", errString)
	})

	//all components in reverse order
	for i := len(a.components) - 1; i >= 0; i-- {
		cutils.Try(func() {
			clog.Infof("[component = %s] -> OnBeforeStop().", a.components[i].Name())
			a.components[i].OnBeforeStop()
		}, func(errString string) {
			clog.Warnf("[component = %s] -> OnBeforeStop(). error = %s", a.components[i].Name(), errString)
		})
	}

	for i := len(a.components) - 1; i >= 0; i-- {
		cutils.Try(func() {
			clog.Infof("[component = %s] -> OnStop().", a.components[i].Name())
			a.components[i].OnStop()
		}, func(errString string) {
			clog.Warnf("[component = %s] -> OnStop(). error = %s", a.components[i].Name(), errString)
		})
	}

	clog.Info("------- application has been shutdown... -------")
}

func (a *Application) OnShutdown(fn ...func()) {
	a.onShutdownFn = append(a.onShutdownFn, fn...)
}

func (a *Application) Shutdown() {
	a.dieChan <- true
}

func (a *Application) SetSerializer(serializer cfacade.ISerializer) {
	if a.Running() {
		return
	}
	a.ISerializer = serializer
}

func (a *Application) SetPacketCodec(codec cfacade.IPacketCodec) {
	if a.Running() {
		return
	}
	a.IPacketCodec = codec
}
