package pomelo

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
)

// DefaultDataRoute default message route handler
func DefaultDataRoute(agent *Agent, route *pmessage.Route, msg *pmessage.Message) {
	// current node
	if agent.NodeType() == route.NodeType() {
		targetPath := cfacade.NewChildPath(agent.NodeID(), route.HandleName(), agent.SID())
		LocalDataRoute(agent, route, msg, targetPath)
		return
	}

	if !agent.IsBind() {
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
	err := ClusterLocalDataRoute(agent, route, msg, member.GetNodeID(), targetPath)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d,route = %s] cluster local data error. err = %v",
			agent.SID(),
			agent.UID(),
			msg.Route,
			err,
		)
	}
}

func LocalDataRoute(agent *Agent, route *pmessage.Route, msg *pmessage.Message, targetPath string) {
	message := cfacade.GetMessage()
	message.Source = agent.AgentPath()
	message.Target = targetPath
	message.FuncName = route.Method()
	message.Session = agent.NewSession(msg.ID)
	message.Args = msg.Data

	agent.ActorSystem().PostLocal(message)
}

func ClusterLocalDataRoute(agent *Agent, route *pmessage.Route, msg *pmessage.Message, nodeID, targetPath string) error {
	message := cfacade.GetMessage()
	message.Source = agent.AgentPath()
	message.Target = targetPath
	message.FuncName = route.Method()
	message.Session = agent.NewSession(msg.ID)
	message.Args = msg.Data

	return agent.Cluster().PublishLocal(nodeID, message)
}
