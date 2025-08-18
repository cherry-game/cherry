package pomelo

import (
	ccode "github.com/cherry-game/cherry/code"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// DefaultDataRoute 默认的消息路由
func DefaultDataRoute(agent *Agent, route *pmessage.Route, msg *pmessage.Message) {
	session := BuildSession(agent, msg)

	// current node
	if agent.NodeType() == route.NodeType() {
		targetPath := cfacade.NewChildPath(agent.NodeID(), route.HandleName(), session.Sid)
		LocalDataRoute(agent, session, route, msg, targetPath)
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

	targetPath := cfacade.NewPath(member.GetNodeID(), route.HandleName())
	errCode := ClusterLocalDataRoute(agent, session, route, msg, member.GetNodeID(), targetPath)
	if ccode.IsFail(errCode) {
		clog.Warnf("[sid = %s,uid = %d,route = %s] cluster local data error. errCode = %v",
			agent.SID(),
			agent.UID(),
			msg.Route,
			errCode,
		)
	}
}

func LocalDataRoute(agent *Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message, targetPath string) {
	message := cfacade.GetMessage()
	message.Source = session.AgentPath
	message.Target = targetPath
	message.FuncName = route.Method()
	message.Session = session
	message.Args = msg.Data

	agent.ActorSystem().PostLocal(&message)
}

func ClusterLocalDataRoute(agent *Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message, nodeID, targetPath string) int32 {
	clusterPacket := cproto.GetClusterPacket()
	clusterPacket.SourcePath = session.AgentPath
	clusterPacket.TargetPath = targetPath
	clusterPacket.FuncName = route.Method()
	clusterPacket.Session = session   // agent session
	clusterPacket.ArgBytes = msg.Data // packet -> message -> data

	return agent.Cluster().PublishLocal(nodeID, clusterPacket)
}

func BuildSession(agent *Agent, msg *pmessage.Message) *cproto.Session {
	agent.session.Mid = uint32(msg.ID)
	return agent.session
}
