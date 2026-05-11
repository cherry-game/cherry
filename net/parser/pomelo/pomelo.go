package pomelo

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// DefaultDataRoute default message route handler
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
	err := ClusterLocalDataRoute(agent, session, route, msg, member.GetNodeID(), targetPath)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d,route = %s] cluster local data error. err = %v",
			agent.SID(),
			agent.UID(),
			msg.Route,
			err,
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

	agent.ActorSystem().PostLocal(message)
}

func ClusterLocalDataRoute(agent *Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message, nodeID, targetPath string) error {
	message := cfacade.GetMessage()
	message.Source = session.AgentPath
	message.Target = targetPath
	message.FuncName = route.Method()
	message.Session = session
	message.Args = msg.Data

	return agent.Cluster().PublishLocal(nodeID, message)
}

func BuildSession(agent *Agent, msg *pmessage.Message) *cproto.Session {
	agent.session.SetMID(uint32(msg.ID))

	return agent.session
}
