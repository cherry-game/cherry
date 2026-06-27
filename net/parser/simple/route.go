package simple

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// Package-level routing state.
var (
	nodeRouteMap    = map[uint32]*NodeRoute{} // mid → target route
	onDataRouteFunc = DefaultDataRoute        // data routing handler
)

// NodeRoute describes the target actor and function for a given message id.
type (
	NodeRoute struct {
		NodeType string // target node type
		ActorID  string // target actor id
		FuncName string // target function name
	}

// DataRouteFunc is called to route a decoded message to the target actor.
	DataRouteFunc func(agent *Agent, msg *Message, route *NodeRoute)
)

// AddNodeRoute maps a mid (message id) to a NodeRoute for routing incoming messages.
func AddNodeRoute(mid uint32, nodeRoute *NodeRoute) {
	if nodeRoute == nil {
		return
	}

	nodeRouteMap[mid] = nodeRoute
}

// GetNodeRoute returns the NodeRoute for the given message id.
func GetNodeRoute(mid uint32) (*NodeRoute, bool) {
	routeActor, found := nodeRouteMap[mid]
	return routeActor, found
}

// DefaultDataRoute is the default message routing handler. It dispatches locally
// when the target node type matches the agent's node, or forwards to a random
// member of the target node type via the cluster.
func DefaultDataRoute(agent *Agent, msg *Message, route *NodeRoute) {
	session := agent.session
	session.SetMID(msg.MID)

	// current node
	if agent.NodeType() == route.NodeType {
		targetPath := cfacade.NewChildPath(agent.NodeID(), route.ActorID, session.Sid)
		LocalDataRoute(agent, session, msg, route, targetPath)
		return
	}

	if !session.IsBind() {
		clog.Warnf("[sid = %s,uid = %d] Session is not bind with UID. failed to forward message.[route = %+v]",
			agent.SID(),
			agent.UID(),
			route,
		)
		return
	}

	member, found := agent.Discovery().Random(route.NodeType)
	if !found {
		return
	}

	targetPath := cfacade.NewPath(member.GetNodeID(), route.ActorID)
	ClusterLocalDataRoute(agent, session, msg, route, member.GetNodeID(), targetPath)
}

// LocalDataRoute posts a message to a local actor on this node.
func LocalDataRoute(agent *Agent, session *cproto.Session, msg *Message, nodeRoute *NodeRoute, targetPath string) {
	message := cfacade.GetMessage()
	message.Source = session.AgentPath
	message.Target = targetPath
	message.FuncName = nodeRoute.FuncName
	message.Session = session
	message.Args = msg.Data

	agent.ActorSystem().PostLocal(message)
}

// ClusterLocalDataRoute publishes a message to a local actor on a remote node via the cluster.
func ClusterLocalDataRoute(agent *Agent, session *cproto.Session, msg *Message, nodeRoute *NodeRoute, nodeID, targetPath string) error {
	message := cfacade.GetMessage()
	message.Source = session.AgentPath
	message.Target = targetPath
	message.FuncName = nodeRoute.FuncName
	message.Session = session
	message.Args = msg.Data

	return agent.Cluster().PublishLocal(nodeID, message)
}
