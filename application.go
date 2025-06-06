package cherry

import (
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	cconst "github.com/cherry-game/cherry/const"
	ctime "github.com/cherry-game/cherry/extend/time"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	cserializer "github.com/cherry-game/cherry/net/serializer"
	cprofile "github.com/cherry-game/cherry/profile"
)

const (
	Cluster    NodeMode = 1 // 集群模式
	Standalone NodeMode = 2 // 单机模式
)

type (
	NodeMode byte

	Application struct {
		cfacade.INode
		isFrontend   bool
		nodeMode     NodeMode
		startTime    ctime.CherryTime     // application start time
		running      int32                // is running
		dieChan      chan bool            // wait for end application
		onShutdownFn []func()             // on shutdown execute functions
		components   []cfacade.IComponent // all components
		serializer   cfacade.ISerializer  // serializer
		discovery    cfacade.IDiscovery   // discovery component
		cluster      cfacade.ICluster     // cluster component
		actorSystem  *cactor.Component    // actor system
		netParser    cfacade.INetParser   // net packet parser
	}
)

// NewApp create new application instance
func NewApp(profileFilePath, nodeID string, isFrontend bool, mode NodeMode) *Application {
	node, err := cprofile.Init(profileFilePath, nodeID)
	if err != nil {
		panic(err)
	}

	return NewAppNode(node, isFrontend, mode)
}

func NewAppNode(node cfacade.INode, isFrontend bool, mode NodeMode) *Application {
	// set logger
	clog.SetNodeLogger(node)

	// print version info
	clog.Info(cconst.GetLOGO())

	app := &Application{
		INode:       node,
		serializer:  cserializer.NewProtobuf(),
		isFrontend:  isFrontend,
		nodeMode:    mode,
		startTime:   ctime.Now(),
		running:     0,
		dieChan:     make(chan bool),
		actorSystem: cactor.New(),
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

// Remove component by name
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

func (a *Application) OnShutdown(fn ...func()) {
	a.onShutdownFn = append(a.onShutdownFn, fn...)
}

// Startup load components before startup
func (a *Application) Startup() {
	defer func() {
		if r := recover(); r != nil {
			clog.Error(r)
		}
	}()

	if a.Running() {
		clog.Error("Application has running.")
		return
	}

	defer func() {
		clog.Flush()
	}()

	// register actor system
	a.Register(a.actorSystem)

	// add connector component
	if a.netParser != nil {
		for _, connector := range a.netParser.Connectors() {
			a.Register(connector)
		}
	}

	clog.Info("-------------------------------------------------")
	clog.Infof("[nodeID      = %s] application is starting...", a.NodeID())
	clog.Infof("[nodeType    = %s]", a.NodeType())
	clog.Infof("[pid         = %d]", os.Getpid())
	clog.Infof("[startTime   = %s]", a.StartTime())
	clog.Infof("[profilePath = %s]", cprofile.Path())
	clog.Infof("[profileName = %s]", cprofile.Name())
	clog.Infof("[env         = %s]", cprofile.Env())
	clog.Infof("[debug       = %v]", cprofile.Debug())
	clog.Infof("[printLevel  = %s]", cprofile.PrintLevel())
	clog.Infof("[logLevel    = %s]", clog.DefaultLogger.LogLevel)
	clog.Infof("[stackLevel  = %s]", clog.DefaultLogger.StackLevel)
	clog.Infof("[writeFile   = %v]", clog.DefaultLogger.EnableWriteFile)
	clog.Infof("[serializer  = %s]", a.serializer.Name())
	clog.Info("-------------------------------------------------")

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

	// execute OnAfterInit()
	for _, c := range a.components {
		clog.Infof("[component = %s] -> OnAfterInit().", c.Name())
		c.OnAfterInit()
	}

	// load net packet parser
	if a.isFrontend {
		if a.netParser == nil {
			clog.Panic("net packet parser is nil.")
		}
		a.netParser.Load(a)
	}

	clog.Info("-------------------------------------------------")
	clog.Infof("[spend time = %dms] application is running.", a.startTime.NowDiffMillisecond())
	clog.Info("-------------------------------------------------")

	// set application is running
	atomic.AddInt32(&a.running, 1)

	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	select {
	case <-a.dieChan:
		clog.Info("invoke shutdown().")
	case s := <-sg:
		clog.Infof("receive shutdown signal = %v.", s)
	}

	// stop status
	atomic.StoreInt32(&a.running, 0)

	clog.Info("------- application will shutdown -------")

	if a.onShutdownFn != nil {
		for _, f := range a.onShutdownFn {
			cutils.Try(func() {
				f()
			}, func(errString string) {
				clog.Warnf("[onShutdownFn] error = %s", errString)
			})
		}
	}

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

func (a *Application) Shutdown() {
	a.dieChan <- true
}

func (a *Application) Serializer() cfacade.ISerializer {
	return a.serializer
}

func (a *Application) Discovery() cfacade.IDiscovery {
	return a.discovery
}

func (a *Application) Cluster() cfacade.ICluster {
	return a.cluster
}

func (a *Application) ActorSystem() cfacade.IActorSystem {
	return a.actorSystem
}

func (a *Application) StartTime() string {
	return a.startTime.ToDateTimeFormat()
}

func (a *Application) SetSerializer(serializer cfacade.ISerializer) {
	if a.Running() || serializer == nil {
		return
	}

	a.serializer = serializer
}

func (a *Application) SetDiscovery(discovery cfacade.IDiscovery) {
	if a.Running() || discovery == nil {
		return
	}

	a.discovery = discovery
}

func (a *Application) SetCluster(cluster cfacade.ICluster) {
	if a.Running() || cluster == nil {
		return
	}

	a.cluster = cluster
}

func (a *Application) SetNetParser(netParser cfacade.INetParser) {
	if a.Running() || netParser == nil {
		return
	}

	a.netParser = netParser
}
