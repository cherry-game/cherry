package pomelo

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	ppacket "github.com/cherry-game/cherry/net/parser/pomelo/packet"
	cproto "github.com/cherry-game/cherry/net/proto"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap/zapcore"
	"time"
)

type (
	Command struct {
		writeBacklog    int
		sysData         map[string]interface{}
		heartbeatTime   time.Duration
		handshakeBytes  []byte
		heartbeatBytes  []byte
		onPacketFuncMap map[ppacket.Type]PacketFunc
		onDataRouteFunc DataRouteFunc
	}

	PacketFunc    func(agent *Agent, packet *ppacket.Packet)
	DataRouteFunc func(agent *Agent, route *pmessage.Route, msg *pmessage.Message)
)

var (
	cmd = Command{
		writeBacklog:    64,
		sysData:         make(map[string]interface{}),
		heartbeatTime:   60 * time.Second,
		handshakeBytes:  make([]byte, 0),
		heartbeatBytes:  make([]byte, 0),
		onPacketFuncMap: make(map[ppacket.Type]PacketFunc, 4),
		onDataRouteFunc: DefaultDataRoute,
	}
)

func (p *Command) init(app cfacade.IApplication) {
	p.sysData["heartbeat"] = p.heartbeatTime.Seconds()
	p.sysData["dict"] = pmessage.GetDictionary()
	p.sysData["serializer"] = app.Serializer().Name()

	handShakeData := map[string]interface{}{
		"code": 200,
		"sys":  p.sysData,
	}

	clog.Infof("[initCommand] Handshake data = %v", handShakeData)

	handShakeJsonData, err := jsoniter.Marshal(handShakeData)
	if err != nil {
		clog.Error(err)
		return
	}

	p.handshakeBytes, err = ppacket.Encode(ppacket.Handshake, handShakeJsonData)
	if err != nil {
		clog.Error(err)
		return
	}

	p.heartbeatBytes, err = ppacket.Encode(ppacket.Heartbeat, nil)
	if err != nil {
		clog.Error(err)
		return
	}

	p.onPacketFuncMap[ppacket.Handshake] = handshakeCommand
	p.onPacketFuncMap[ppacket.HandshakeAck] = handshakeACKCommand
	p.onPacketFuncMap[ppacket.Heartbeat] = heartbeatCommand
	p.onPacketFuncMap[ppacket.Data] = dataCommand
}

func handshakeCommand(agent *Agent, _ *ppacket.Packet) {
	agent.SetState(AgentWaitAck)
	agent.SendRaw(cmd.handshakeBytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Request handshake. [address = %s]",
			agent.SID(),
			agent.UID(),
			agent.RemoteAddr(),
		)
	}
}

func handshakeACKCommand(agent *Agent, _ *ppacket.Packet) {
	agent.SetState(AgentWorking)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] request handshakeACK. [address = %s]",
			agent.SID(),
			agent.UID(),
			agent.RemoteAddr(),
		)
	}
}

func heartbeatCommand(agent *Agent, _ *ppacket.Packet) {
	agent.SendRaw(cmd.heartbeatBytes)
}

func dataCommand(agent *Agent, pkg *ppacket.Packet) {
	if agent.State() != AgentWorking {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Data State is not working. [state = %d]",
				agent.SID(),
				agent.UID(),
				agent.State(),
			)
		}
		return
	}

	msg, err := pmessage.Decode(pkg.Data())
	if err != nil {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Data message decode error. [data = %s, error = %s]",
				agent.SID(),
				agent.UID(),
				pkg.Data(),
				err,
			)
		}
		return
	}

	route, err := pmessage.DecodeRoute(msg.Route)
	if err != nil {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Data Message decode route error. [data = %s, error = %s]",
				agent.SID(),
				agent.UID(),
				pkg.Data(),
				err,
			)
		}
		return
	}

	cmd.onDataRouteFunc(agent, route, &msg)
}

// DefaultDataRoute 默认的消息路由
func DefaultDataRoute(agent *Agent, route *pmessage.Route, msg *pmessage.Message) {
	session := BuildSession(agent, msg)

	// current node
	if agent.NodeType() == route.NodeType() {
		targetPath := cfacade.NewChildPath(agent.NodeId(), route.HandleName(), session.Sid)
		LocalDataRoute(agent, &session, route, msg, targetPath)
		return
	}

	if !session.IsBind() {
		clog.Warnf("[sid = %s,uid = %d] Session is not bind with UID. failed to forward message.[route = %s]",
			agent.SID(),
			agent.UID(),
			msg.Route,
		)
		return
	}

	member, found := agent.Discovery().Random(route.NodeType())
	if !found {
		return
	}

	targetPath := cfacade.NewPath(member.GetNodeId(), route.HandleName())
	ClusterLocalDataRoute(agent, &session, route, msg, member.GetNodeId(), targetPath)
}

func LocalDataRoute(agent *Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message, targetPath string) {
	message := cfacade.GetMessage()
	message.Source = session.AgentPath
	message.Target = targetPath
	message.FuncName = route.Method()
	message.Session = session
	message.Args = msg.Data

	agent.ActorSystem().PostLocal(message)
}

func ClusterLocalDataRoute(agent *Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message, nodeId, targetPath string) {
	clusterPacket := cproto.GetClusterPacket()
	clusterPacket.SourcePath = session.AgentPath
	clusterPacket.TargetPath = targetPath
	clusterPacket.FuncName = route.Method()
	clusterPacket.Session = session   // agent session
	clusterPacket.ArgBytes = msg.Data // packet -> message -> data

	err := agent.Cluster().PublishLocal(nodeId, clusterPacket)
	if err != nil {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Publish local fail. [nodeId = %s, target = %s, funcName = %s, error = %s]",
				agent.SID(),
				agent.UID(),
				nodeId,
				clusterPacket.TargetPath,
				clusterPacket.FuncName,
				err,
			)
		}
	}
}

func BuildSession(agent *Agent, msg *pmessage.Message) cproto.Session {
	session := agent.session.Copy()
	session.Mid = uint32(msg.ID)
	return session
}
