package simple

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	"go.uber.org/zap/zapcore"
)

var (
	nodeRouteMap    = map[uint32]*NodeRoute{}
	onDataRouteFunc = DefaultDataRoute
)

type (
	NodeRoute struct {
		NodeType string
		ActorID  string
		FuncName string
	}

	DataRouteFunc func(agent *Agent, msg *Message, route *NodeRoute)
)

func AddNodeRoute(mid uint32, nodeRoute *NodeRoute) {
	if nodeRoute == nil {
		return
	}

	nodeRouteMap[mid] = nodeRoute
}

func GetNodeRoute(mid uint32) (*NodeRoute, bool) {
	routeActor, found := nodeRouteMap[mid]
	return routeActor, found
}

func DefaultDataRoute(agent *Agent, msg *Message, route *NodeRoute) {
	session := agent.session.Copy()
	session.Mid = msg.MID

	// current node
	if agent.NodeType() == route.NodeType {
		targetPath := cfacade.NewChildPath(agent.NodeId(), route.ActorID, session.Sid)
		LocalDataRoute(agent, &session, msg, route, targetPath)
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

	targetPath := cfacade.NewPath(member.GetNodeId(), route.ActorID)
	ClusterLocalDataRoute(agent, &session, msg, route, member.GetNodeId(), targetPath)
}

func LocalDataRoute(agent *Agent, session *cproto.Session, msg *Message, nodeRoute *NodeRoute, targetPath string) {
	message := cfacade.GetMessage()
	message.Source = session.AgentPath
	message.Target = targetPath
	message.FuncName = nodeRoute.FuncName
	message.Session = session
	message.Args = msg.Data

	agent.ActorSystem().PostLocal(message)
}

func ClusterLocalDataRoute(agent *Agent, session *cproto.Session, msg *Message, nodeRoute *NodeRoute, nodeID, targetPath string) {
	clusterPacket := cproto.GetClusterPacket()
	clusterPacket.SourcePath = session.AgentPath
	clusterPacket.TargetPath = targetPath
	clusterPacket.FuncName = nodeRoute.FuncName
	clusterPacket.Session = session   // agent session
	clusterPacket.ArgBytes = msg.Data // packet -> message -> data

	err := agent.Cluster().PublishLocal(nodeID, clusterPacket)
	if err != nil {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Publish local fail. [nodeID = %s, target = %s, funcName = %s, error = %s]",
				agent.SID(),
				agent.UID(),
				nodeID,
				clusterPacket.TargetPath,
				clusterPacket.FuncName,
				err,
			)
		}
	}
}
