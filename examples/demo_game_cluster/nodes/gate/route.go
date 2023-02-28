package gate

import (
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/pb"
	sessionKey "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/session_key"
	cslice "github.com/cherry-game/cherry/extend/slice"
	cstring "github.com/cherry-game/cherry/extend/string"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	cproto "github.com/cherry-game/cherry/net/proto"
)

var (
	// 客户端连接后，必需先执行第一条协议，进行token验证后，才能进行后续的逻辑
	firstRouteName = "gate.user.login"

	// 角色进入游戏前的，前三个协议
	beforeLoginRoutes = []string{
		"game.player.select", //查询玩家角色
		"game.player.create", //玩家创建角色
		"game.player.enter",  //玩家角色进入游戏
	}

	notLoginRsp = &pb.Int32{
		Value: code.PlayerDenyLogin,
	}
)

// onDataRoute 数据路由规则
//
// 登录逻辑:
// 1.(建立连接)客户端建立连接，服务端对应创建一个agent用于处理玩家消息,actorID == sid
// 2.(用户登录)客户端进行帐号登录验证，通过uid绑定当前sid
// 3.(角色登录)客户端通过'beforeLoginRoutes'中的协议完成角色登录
func onDataRoute(agent *pomelo.Agent, route *pmessage.Route, msg *pmessage.Message) {
	// agent没有"用户登录",且请求不是第一条协议，则踢掉agent，断开连接
	if !agent.Session().IsBind() && msg.Route != firstRouteName {
		agent.Kick(notLoginRsp, true)
		return
	}

	session := pomelo.BuildSession(agent, msg)
	if agent.App().NodeType() == route.NodeType() {
		pomelo.LocalDataRoute(agent, &session, route, msg)
		return
	}

	gameNodeRoute(agent, &session, route, msg)
}

// gameNodeRoute 实现agent路由消息到游戏节点的函数
func gameNodeRoute(agent *pomelo.Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message) {
	// agent没有完成"角色登录",则禁止转发到game节点
	if !session.Contains(sessionKey.PlayerID) {
		// 如果不是角色登录协议则踢掉agent
		_, found := cslice.StringIn(msg.Route, beforeLoginRoutes)
		if !found {
			agent.Kick(notLoginRsp, true)
			return
		}
	}

	// 获取绑定的ServerID，进行消息转发
	serverId := session.GetString(sessionKey.ServerID)
	member, found := agent.App().Discovery().GetMember(serverId)
	if !found {
		clog.Warnf("[sid = %s,uid = %d] Find node fail. [route = %s]",
			agent.SID(),
			agent.UID(),
			msg.Route,
		)
		return
	}

	childID := cstring.ToString(agent.UID())

	clusterPacket := cproto.GetClusterPacket()
	clusterPacket.SourcePath = session.AgentPath
	clusterPacket.TargetPath = cfacade.NewPath(member.GetNodeId(), route.HandleName(), childID)
	clusterPacket.FuncName = route.Method()
	clusterPacket.Session = session   // agent session
	clusterPacket.ArgBytes = msg.Data // packet -> message -> data

	pomelo.PublishClusterLocal(agent, member.GetNodeId(), clusterPacket)
}
