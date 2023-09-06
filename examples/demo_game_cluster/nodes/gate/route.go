package gate

import (
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/pb"
	sessionKey "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/session_key"
	cslice "github.com/cherry-game/cherry/extend/slice"
	cstring "github.com/cherry-game/cherry/extend/string"
	cfacade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	cproto "github.com/cherry-game/cherry/net/proto"
)

var (
	// 客户端连接后，必需先执行第一条协议，进行token验证后，才能进行后续的逻辑
	firstRouteName = "gate.user.login"

	// 角色进入游戏时的前三个协议
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
func onPomeloDataRoute(agent *pomelo.Agent, route *pmessage.Route, msg *pmessage.Message) {
	session := pomelo.BuildSession(agent, msg)

	// agent没有"用户登录",且请求不是第一条协议，则踢掉agent，断开连接
	if !session.IsBind() && msg.Route != firstRouteName {
		agent.Kick(notLoginRsp, true)
		return
	}

	if agent.NodeType() == route.NodeType() {
		targetPath := cfacade.NewChildPath(agent.NodeId(), route.HandleName(), session.Sid)
		pomelo.LocalDataRoute(agent, session, route, msg, targetPath)
	} else {
		gameNodeRoute(agent, session, route, msg)
	}
}

// gameNodeRoute 实现agent路由消息到游戏节点
func gameNodeRoute(agent *pomelo.Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message) {
	if !session.IsBind() {
		return
	}

	// 如果agent没有完成"角色登录",则禁止转发到game节点
	if !session.Contains(sessionKey.PlayerID) {
		// 如果不是角色登录协议则踢掉agent
		if found := cslice.StringInSlice(msg.Route, beforeLoginRoutes); !found {
			agent.Kick(notLoginRsp, true)
			return
		}
	}

	serverId := session.GetString(sessionKey.ServerID)
	if serverId == "" {
		return
	}

	childId := cstring.ToString(session.Uid)
	targetPath := cfacade.NewChildPath(serverId, route.HandleName(), childId)
	pomelo.ClusterLocalDataRoute(agent, session, route, msg, serverId, targetPath)
}
