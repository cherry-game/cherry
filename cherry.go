package cherry

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
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
	_components       []cherryFacade.IComponent
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

func SetSerializer(serializer cherryFacade.ISerializer) {
	_thisApp.SetSerializer(serializer)
}

func SetPacketCodec(codec cherryFacade.IPacketCodec) {
	_thisApp.SetPacketCodec(codec)
}

func Run(isFrontend bool, nodeMode NodeMode) {
	if _thisApp == nil {
		panic("please call the configure function first.")
	}

	if _thisApp.Running() {
		return
	}

	_thisApp.isFrontend = isFrontend
	_thisApp.nodeMode = nodeMode

	initHandlerComponent()
	initClusterComponent()
	initConnectorComponent()

	_thisApp.Startup()
}

func initHandlerComponent() {
	// register handler component
	_handlerComponent = cherryHandler.NewComponent(_handlerOpts...)
	for _, group := range _handlerGroups {
		_handlerComponent.Register(group)
	}
	_thisApp.Register(_handlerComponent)

	// add developer registered components
	_thisApp.Register(_components...)
}

func initClusterComponent() {
	// register cluster component
	if _thisApp.NodeMode() == Cluster {
		_clusterComponent = cherryCluster.NewComponent()
		_thisApp.Register(_clusterComponent)

		//配置 cluster 消息路由

		// cluster 收到rpc forward函数的消息
		_clusterComponent.OnForward(func(msg *cherryProto.Message) {
			session, found := cherrySession.GetBySID(msg.Sid)
			if found {
				// cluster接收的消息，转到handler统一处理
				_handlerComponent.PostMessage(session, &cherryMessage.Message{
					Type:  cherryMessage.Type(msg.GetMsgType()),
					ID:    uint(msg.Id),
					Route: msg.Route,
					Data:  msg.Data,
					Error: false,
				})
			}
		})

		// TODO setting remote handler
		_handlerComponent.RemoteHandler = _clusterComponent.SendUserMessage
	}
}

func initConnectorComponent() {
	if _connector == nil {
		return
	}

	if len(_handshakeData) < 1 {
		_handshakeData["heartbeat"] = _heartbeat.Seconds()
	}

	initCommand()

	connectorComponent := cherryConnector.NewComponent(_connector)
	stat := connectorComponent.ConnectStat

	cherrySession.AddOnCreateListener(func(session *cherrySession.Session) (next bool) {
		stat.IncreaseConn()

		session.Debugf("session on create. address[%s], state[%s]",
			session.RemoteAddress(),
			stat.PrintInfo(),
		)
		return true
	})

	cherrySession.AddOnCreateListener(func(session *cherrySession.Session) (next bool) {
		stat.DecreaseConn()
		session.Debugf("session on closed. address[%s], state[%s]",
			session.RemoteAddress(),
			stat.PrintInfo(),
		)

		//这会产生问题， 如果是前端服务器，则需要这么通知
		//后端服务器，不必要通知
		if _connector != nil && _clusterComponent != nil {
			_clusterComponent.SendCloseSession(session.SID())
		}

		return true
	})

	agentOptions := &cherryAgent.Options{
		Heartbeat: _heartbeat,
		Commands:  _commands,
	}

	// TODO rpc send interface
	if _clusterComponent != nil {
		//如果是集群，则转发到这里
		agentOptions.RPCHandler = _clusterComponent.SendSysMessage
	} else {
		// 如果是单机，则转到这里
		agentOptions.RPCHandler = _handlerComponent.PostMessage
	}

	// new client connect
	connectorComponent.OnConnect(func(conn cherryFacade.INetConn) {
		// create agent
		agent := cherryAgent.NewAgent(_thisApp, conn, agentOptions)

		// create new session
		session := cherrySession.Create(_thisApp.NodeId(), agent)
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

	// TODO connector forward message
	handDataCommand := cherryCommand.NewData(_handlerComponent.PostMessage)
	RegisterCommand(handDataCommand)
}

func SetHeartbeat(time time.Duration) {
	_heartbeat = time
}

func SetHandshake(key string, value interface{}) {
	_handshakeData[key] = value
}

func SetDictionary(dict map[string]uint16) {
	cherryMessage.SetDictionary(dict)
}

func SetMessageCompression(compression bool) {
	cherryMessage.SetDataCompression(compression)
}

func SetOnShutdown(fn ...func()) {
	_thisApp.OnShutdown(fn...)
}

func SetHandlerOptions(opts ...cherryHandler.Option) {
	_handlerOpts = append(_handlerOpts, opts...)
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
