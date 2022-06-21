package cherry

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
	cherryCommand "github.com/cherry-game/cherry/net/command"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherryProto "github.com/cherry-game/cherry/net/proto"
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
	_commands         = make(map[cherryPacket.Type]cherryCommand.ICommand)
	_handshakeData    = make(map[string]interface{})
	_heartbeat        = 60 * time.Second
	_connectors       []cherryFacade.IConnector
	_clusterComponent *cherryCluster.Component
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

	initHandler()
	initCluster()
	initConnector()
	initComponent()

	_thisApp.Startup()
}

func initHandler() {
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

func initComponent() {
	_thisApp.Register(_components...)
}

func initCluster() {
	if _thisApp.NodeMode() == Cluster {
		// register cluster component
		_clusterComponent = cherryCluster.NewComponent()
		_thisApp.Register(_clusterComponent)
	}
}

func initConnector() {
	if _thisApp.isFrontend == false {
		return
	}

	if len(_connectors) < 1 {
		panic("please call the cherry.RegisterConnector() method add IConnector.")
	}

	initCommand()
	initOnSession()

	for _, connector := range _connectors {
		// default setting
		if connector.IsSetListener() == false {
			connector.OnConnectListener(func(conn cherryFacade.INetConn) {
				// create agent
				agent := cherryAgent.NewAgent(_thisApp, conn, &cherryAgent.Options{
					Heartbeat: _heartbeat,
					Commands:  _commands,
				})

				// create new session
				newSession := cherrySession.Create(cherrySession.NextSID(), _thisApp.NodeId(), agent)
				// run agent
				agent.SetSession(newSession)
				agent.Run()
			})
		}

		RegisterComponent(connector)
	}
}

func initOnSession() {
	cherrySession.AddOnCreateListener(func(session *cherrySession.Session) (next bool) {
		session.Debugf("session create. [sid = %s, address = %s]",
			session.SID(),
			session.RemoteAddress(),
		)
		return true
	})

	cherrySession.AddOnCloseListener(func(session *cherrySession.Session) (next bool) {
		session.Debugf("session closed. [sid = %s, address = %s]",
			session.SID(),
			session.RemoteAddress(),
		)

		return true
	})
}

func initCommand() {
	if _, found := _commands[cherryPacket.Handshake]; found == false {
		if len(_handshakeData) < 1 {
			_handshakeData["heartbeat"] = _heartbeat.Seconds()
			_handshakeData["dict"] = cherryMessage.GetDictionary()
			_handshakeData["serializer"] = _thisApp.ISerializer.Name()
		}

		handshakeCommand := cherryCommand.NewHandshake(_thisApp, _handshakeData)
		RegisterCommand(handshakeCommand)
	}

	if _, found := _commands[cherryPacket.HandshakeAck]; found == false {
		handshakeAckCommand := cherryCommand.NewHandshakeACK()
		RegisterCommand(handshakeAckCommand)
	}

	if _, found := _commands[cherryPacket.Heartbeat]; found == false {
		heartbeatCommand := cherryCommand.NewHeartbeat(_thisApp)
		RegisterCommand(heartbeatCommand)
	}

	if _, found := _commands[cherryPacket.Data]; found == false {
		// connector forward message
		handDataCommand := cherryCommand.NewData(
			_thisApp,
			_handlerComponent.ProcessLocal,
			_clusterComponent.ForwardLocal,
		)
		RegisterCommand(handDataCommand)
	}
}

func SetSerializer(serializer cherryFacade.ISerializer) {
	_thisApp.SetSerializer(serializer)
}

func SetPacketCodec(codec cherryFacade.IPacketCodec) {
	_thisApp.SetPacketCodec(codec)
}

func SetHeartbeat(t time.Duration) {
	if t.Seconds() < 1 {
		t = 60 * time.Second
	}
	_heartbeat = t
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
	_connectors = append(_connectors, connector)
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

func GetConnectors() []cherryFacade.IConnector {
	return _connectors
}

func RPC(nodeId string, route string, arg proto.Message, reply proto.Message, timeout ...time.Duration) int32 {
	if reply != nil && reflect.TypeOf(reply).Kind() != reflect.Ptr {
		return cherryCode.RPCReplyParamsError
	}

	var requestTimeout time.Duration
	if len(timeout) > 0 {
		requestTimeout = timeout[0]
	}

	rsp := &cherryProto.Response{}
	_clusterComponent.Client().CallRemote(nodeId, route, arg, requestTimeout, rsp)
	if cherryCode.IsFail(rsp.Code) {
		return rsp.Code
	}

	if reply != nil {
		err := proto.Unmarshal(rsp.GetData(), reply)
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

func PostEvent(event cherryFacade.IEvent) {
	if _handlerComponent == nil {
		cherryLogger.Warnf("post event fail. handler component is nil.")
		return
	}
	_handlerComponent.PostEvent(event)
}
