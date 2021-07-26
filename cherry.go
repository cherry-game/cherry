package cherry

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
	cherryCommand "github.com/cherry-game/cherry/net/command"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"time"
)

var (
	_thisApp *Application
)

var (
	_serializer       cherryFacade.ISerializer
	_codec            cherryFacade.IPacketCodec
	_components       []cherryFacade.IComponent
	_sessionComponent *cherrySession.Component
	_clusterComponent *cherryCluster.Component
)

var (
	_commands      = make(map[cherryPacket.Type]cherryCommand.ICommand)
	_handshakeData = make(map[string]interface{})
	_heartbeat     = 60 * time.Second
	_connector     cherryFacade.IConnector
)

var (
	_handlerOpts      []cherryHandler.Option
	_handlerGroups    []*cherryHandler.HandlerGroup
	_handlerComponent *cherryHandler.Component
)

func App() *Application {
	return _thisApp
}

func Configure(profilePath, profileName, nodeId string) cherryFacade.IApplication {
	_thisApp = NewApp(profilePath, profileName, nodeId)
	return _thisApp
}

func Run() {
	if _thisApp == nil {
		panic("please call the configure function first.")
	}

	if _thisApp.Running() {
		return
	}

	if _codec != nil {
		_thisApp.SetPacketCodec(_codec)
	}

	if _serializer != nil {
		_thisApp.SetSerializer(_serializer)
	}

	// register session component
	_sessionComponent = cherrySession.NewComponent()
	_thisApp.Register(_sessionComponent)

	// register handler component
	_handlerComponent = cherryHandler.NewComponent(_handlerOpts...)
	for _, group := range _handlerGroups {
		_handlerComponent.Register(group)
	}
	_thisApp.Register(_handlerComponent)

	// add developer registered components
	_thisApp.Register(_components...)

	// register cluster component
	if _clusterComponent != nil {
		_thisApp.Register(_clusterComponent)
		//配置 cluster 消息路由
	}

	// register connector component
	if _connector != nil {
		initConnectorComponent()
	}

	_thisApp.Startup()
}

func initConnectorComponent() {
	if len(_handshakeData) < 1 {
		_handshakeData["heartbeat"] = _heartbeat.Seconds()
	}

	initCommand()

	connectorComponent := cherryConnector.NewComponent(_connector)
	stat := connectorComponent.ConnectStat

	_sessionComponent.AddOnCreate(func(session *cherrySession.Session) (next bool) {
		stat.IncreaseConn()

		session.Debugf("session on create. address[%s], state[%s]",
			session.RemoteAddress(),
			stat.PrintInfo(),
		)
		return true
	})

	_sessionComponent.AddOnClose(func(session *cherrySession.Session) (next bool) {
		stat.DecreaseConn()
		session.Debugf("session on closed. address[%s], state[%s]",
			session.RemoteAddress(),
			stat.PrintInfo(),
		)
		return true
	})

	agentOptions := &cherryAgent.Options{
		Heartbeat: _heartbeat,
		Command:   _commands,
	}

	if _clusterComponent != nil {
		// set rpc handler TODO clusterComponent提供接口
		agentOptions.RPCHandler = _clusterComponent.SendUserMessage
	} else {
		agentOptions.RPCHandler = _handlerComponent.PostMessage
	}

	// new client connect
	connectorComponent.OnConnect(func(conn cherryFacade.INetConn) {
		// create agent
		agent := cherryAgent.NewAgent(_thisApp, conn, agentOptions)

		// create new session
		session := _sessionComponent.Create(agent)
		agent.Session = session
		// run agent
		agent.Run()
	})

	_thisApp.Register(connectorComponent)
}

func initCommand() {
	if len(_commands) > 0 {
		return
	}

	// default values
	handshakeCommand := cherryCommand.NewHandshake(_thisApp, _handshakeData)
	RegisterCommand(handshakeCommand)

	handshakeAckCommand := cherryCommand.NewHandshakeACK()
	RegisterCommand(handshakeAckCommand)

	heartbeatCommand := cherryCommand.NewHeartbeat(_thisApp)
	RegisterCommand(heartbeatCommand)

	handDataCommand := cherryCommand.NewData(_handlerComponent.PostMessage)
	RegisterCommand(handDataCommand)
}

func SetHandlerOptions(opts ...cherryHandler.Option) {
	_handlerOpts = append(_handlerOpts, opts...)
}

func SetSerializer(serializer cherryFacade.ISerializer) {
	_serializer = serializer
}

func SetPacketCodec(codec cherryFacade.IPacketCodec) {
	_codec = codec
}

func SetHeartbeat(time time.Duration) {
	_heartbeat = time
}

func SetHandshake(key string, value interface{}) {
	_handshakeData[key] = value
}

func SetDictionary(dict map[string]uint16) {
	if _thisApp.Running() {
		return
	}

	cherryMessage.SetDictionary(dict)
}

func SetMessageCompression(compression bool) {
	if _thisApp.Running() {
		return
	}

	cherryMessage.SetDataCompression(compression)
}

func SetOnShutdown(fn ...func()) {
	_thisApp.OnShutdown(fn...)
}

func RegisterHandler(handler ...cherryFacade.IHandler) {
	handlerGroup := cherryHandler.NewGroupWithHandler(handler...)
	_handlerGroups = append(_handlerGroups, handlerGroup)
}

func RegisterHandlerGroup(group ...*cherryHandler.HandlerGroup) {
	_handlerGroups = append(_handlerGroups, group...)
}

func RegisterComponent(component ...cherryFacade.IComponent) {
	_components = append(_components, component...)
}

func RegisterConnector(connector cherryFacade.IConnector) {
	_connector = connector
}

func RegisterCommand(command cherryCommand.ICommand) {
	_commands[command.GetType()] = command
}

func RegisterCluster() {
	_clusterComponent = cherryCluster.NewComponent()
}
