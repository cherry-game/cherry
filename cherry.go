package cherry

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
	cherryCommand "github.com/cherry-game/cherry/net/command"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	cherryDiscovery "github.com/cherry-game/cherry/net/discovery"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherryRouter "github.com/cherry-game/cherry/net/router"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"github.com/golang/protobuf/proto"
	"reflect"
	"time"
)

var (
	_thisApp    *Application
	_components []cherryFacade.IComponent
)

var (
	_commands           = make(map[cherryPacket.Type]cherryCommand.ICommand)
	_handshakeData      = make(map[string]interface{})
	_heartbeat          = 60 * time.Second
	_connector          cherryFacade.IConnector
	_connectorComponent *cherryConnector.Component
	_clusterComponent   *cherryCluster.Component
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
	initRegisterComponent()
	initClusterComponent()
	initConnectorComponent()

	_thisApp.Startup()
}

func initHandlerComponent() {
	// register handler component
	_handlerComponent = cherryHandler.NewComponent(_handlerOpts...)

	if _thisApp.isFrontend {
		// add session handler for frontend node
		_handlerComponent.Register2Group(&cherryHandler.SessionHandler{})
	}

	for _, group := range _handlerGroups {
		_handlerComponent.Register(group)
	}

	// add handler component
	_thisApp.Register(_handlerComponent)
}

func initRegisterComponent() {
	_thisApp.Register(_components...)
}

func initClusterComponent() {
	// register cluster component
	if _thisApp.NodeMode() == Cluster {
		_clusterComponent = cherryCluster.NewComponent(_handlerComponent)
		_thisApp.Register(_clusterComponent)
	}
}

func initConnectorComponent() {
	if _thisApp.isFrontend == false {
		return
	}

	if _connector == nil {
		panic("call the cherry.RegisterConnector() method add IConnector.")
	}

	if len(_handshakeData) < 1 {
		_handshakeData["heartbeat"] = _heartbeat.Seconds()
	}

	initCommand()

	_connectorComponent = cherryConnector.NewComponent(_connector)

	cherrySession.AddOnCreateListener(func(session *cherrySession.Session) (next bool) {
		session.Debugf("session create. [nodeId = %s, sid = %s, address = %s]",
			_thisApp.NodeId(),
			session.SID(),
			session.RemoteAddress(),
		)
		return true
	})

	cherrySession.AddOnCloseListener(func(session *cherrySession.Session) (next bool) {
		session.Debugf("session closed. [nodeId = %s, sid = %s, address = %s]",
			_thisApp.NodeId(),
			session.SID(),
			session.RemoteAddress(),
		)

		return true
	})

	agentOptions := &cherryAgent.Options{
		Heartbeat: _heartbeat,
		Commands:  _commands,
	}

	// new client connect
	_connectorComponent.OnConnect(func(conn cherryFacade.INetConn) {
		// create agent
		agent := cherryAgent.NewAgent(_thisApp, conn, agentOptions)
		// create new session
		agent.Session = cherrySession.Create(cherrySession.NextSID(), _thisApp.NodeId(), agent)
		// run agent
		agent.Run()
	})

	_thisApp.Register(_connectorComponent)
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
	handDataCommand := cherryCommand.NewData(
		_thisApp,
		_handlerComponent.ProcessLocal,
		_clusterComponent.ForwardLocal,
	)
	RegisterCommand(handDataCommand)
}

func SetSerializer(serializer cherryFacade.ISerializer) {
	_thisApp.SetSerializer(serializer)
}

func SetPacketCodec(codec cherryFacade.IPacketCodec) {
	_thisApp.SetPacketCodec(codec)
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

func AddNodeRouter(nodeType string, routingFunc cherryRouter.RoutingFunc) {
	cherryRouter.AddRoute(nodeType, routingFunc)
}

func GetCluster() *cherryCluster.Component {
	return _clusterComponent
}

func GetRPCClient() cherryFacade.RPCClient {
	return _clusterComponent.Client()
}

func GetConnector() *cherryConnector.Component {
	return _connectorComponent
}

func RPC(nodeId string, route string, arg proto.Message, reply proto.Message, timeout ...time.Duration) (code int32) {
	if reply != nil && reflect.TypeOf(reply).Kind() != reflect.Ptr {
		return cherryCode.RPCReplyParamsError
	}

	var requestTimeout time.Duration
	if len(timeout) > 0 {
		requestTimeout = timeout[0]
	}

	callResult := _clusterComponent.Client().CallRemote(nodeId, route, arg, requestTimeout)
	if cherryCode.IsFail(callResult.Code) {
		return code
	}

	if reply != nil {
		err := proto.Unmarshal(callResult.GetData(), reply)
		if err != nil {
			return cherryCode.RPCUnmarshalError
		}
	}

	return cherryCode.OK
}

func RPCByRoute(route string, arg proto.Message, reply proto.Message, timeout ...time.Duration) int32 {
	rt, err := cherryMessage.DecodeRoute(route)
	if err != nil {
		cherryLogger.Warnf("[RPCByRoute] decode route fail.. [error = %s]", err)
		return cherryCode.RPCRouteDecodeError
	}

	member, err := cherryRouter.Route(context.Background(), rt.NodeType(), nil)
	if err != nil {
		cherryLogger.Warnf("[RPCByRoute]get node router is fail. [route = %s] [error = %s]", route, err)
		return cherryCode.RPCRouteHashError
	}

	return RPC(member.GetNodeId(), route, arg, reply, timeout...)
}

func RPCAsync(nodeId string, route string, arg proto.Message) {
	if nodeId == "" {
		decode, err := cherryMessage.DecodeRoute(route)
		if err != nil {
			cherryLogger.Warnf("[RPCAsync] decode route fail. [route = %s]", route)
			return
		}

		member, err := cherryRouter.Route(context.Background(), decode.NodeType(), nil)
		if err != nil {
			cherryLogger.Warnf("[RPCAsync] get node router is fail. [route = %s] [error = %s]", route, err)
			return
		}

		nodeId = member.GetNodeId()
	}

	_clusterComponent.Client().CallRemoteAsync(nodeId, route, arg)
}

func RPCAsyncByRoute(route string, arg proto.Message) {
	RPCAsync("", route, arg)
}

func RPCAsyncByNodeType(nodeType string, route string, arg proto.Message, filterNodeId ...string) {
	members := cherryDiscovery.ListByType(nodeType, filterNodeId...)

	for _, member := range members {
		RPCAsync(member.GetNodeId(), route, arg)
	}
}
