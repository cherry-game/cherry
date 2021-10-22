package cherry

import (
	"fmt"
	cherryConst "github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/extend/utils"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/profile"
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
		facade.INode
		facade.ISerializer
		facade.IPacketCodec
		isFrontend   bool
		nodeMode     NodeMode
		startTime    cherryTime.CherryTime // application start time
		running      int32                 // is running
		die          chan bool             // wait for end application
		components   []facade.IComponent   // all components
		onShutdownFn []func()              // on shutdown execute functions
	}
)

// NewApp create new application instance
func NewApp(profilePath, profileName, nodeId string) *Application {
	_, err := cherryProfile.Init(profilePath, profileName)
	if err != nil {
		panic(fmt.Sprintf("init profile fail. error = %s", err))
	}

	node, err := cherryProfile.LoadNode(nodeId)
	if err != nil {
		panic(err)
	}

	// set logger
	cherryLogger.SetNodeLogger(node)

	// print version info
	cherryLogger.Info(cherryConst.GetLOGO())

	app := &Application{
		INode:        node,
		startTime:    cherryTime.Now(),
		running:      0,
		die:          make(chan bool),
		ISerializer:  cherrySerializer.NewProtobuf(),
		IPacketCodec: cherryPacket.NewPomeloCodec(),
		isFrontend:   true,
		nodeMode:     Standalone,
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

func (a *Application) Register(components ...facade.IComponent) {
	if a.Running() {
		return
	}

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
	}
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

// Startup load components before startup
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

	cherryLogger.Info("-------------------------------------------------")
	cherryLogger.Infof("[nodeId      = %s] application is starting...", a.NodeId())
	cherryLogger.Infof("[nodeType    = %s]", a.NodeType())
	cherryLogger.Infof("[pid         = %d]", os.Getpid())
	cherryLogger.Infof("[startTime   = %s]", a.StartTime())
	cherryLogger.Infof("[profile     = %s]", cherryProfile.Name())
	cherryLogger.Infof("[profilePath = %s]", cherryProfile.Dir())
	cherryLogger.Infof("[profileFile = %s]", cherryProfile.FileName())
	cherryLogger.Infof("[debug       = %v]", cherryProfile.Debug())
	cherryLogger.Infof("[logLevel    = %s]", cherryLogger.DefaultLogger.Level)
	cherryLogger.Infof("[stackLevel  = %s]", cherryLogger.DefaultLogger.StackLevel)
	cherryLogger.Infof("[writeFile   = %v]", cherryLogger.DefaultLogger.EnableWriteFile)
	cherryLogger.Infof("[codec       = %v]", reflect.TypeOf(a.IPacketCodec))
	cherryLogger.Infof("[serializer  = %s]", a.ISerializer.Name())
	cherryLogger.Info("-------------------------------------------------")

	// add components
	a.Register(components...)

	// is running
	atomic.AddInt32(&a.running, 1)

	// component list
	for _, c := range a.components {
		c.Set(a)
		cherryLogger.Infof("[component = %s] is added.", c.Name())
	}
	cherryLogger.Info("-------------------------------------------------")

	// execute Init()
	for _, c := range a.components {
		cherryLogger.Infof("[component = %s] -> OnInit().", c.Name())
		c.Init()
	}
	cherryLogger.Info("-------------------------------------------------")

	//execute OnAfterInit()
	for _, c := range a.components {
		cherryLogger.Infof("[component = %s] -> OnAfterInit().", c.Name())
		c.OnAfterInit()
	}
	cherryLogger.Info("-------------------------------------------------")
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
			cherryLogger.Infof("[component = %s] -> OnBeforeStop().", a.components[i].Name())
			a.components[i].OnBeforeStop()
		}, func(errString string) {
			cherryLogger.Warnf("[component = %s] -> OnBeforeStop(). error = %s", a.components[i].Name(), errString)
		})
	}

	for i := len(a.components) - 1; i >= 0; i-- {
		cherryUtils.Try(func() {
			cherryLogger.Infof("[component = %s] -> OnStop().", a.components[i].Name())
			a.components[i].OnStop()
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
	if a.Running() {
		return
	}
	a.ISerializer = serializer
}

func (a *Application) SetPacketCodec(codec facade.IPacketCodec) {
	if a.Running() {
		return
	}
	a.IPacketCodec = codec
}
