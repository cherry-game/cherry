package player

import (
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/event"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/pb"
	sessionKey "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/session_key"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/nodes/game/db"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/nodes/game/module/online"
	cstring "github.com/cherry-game/cherry/extend/string"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

type (
	// actorPlayer 每位登录的玩家对应一个子actor
	actorPlayer struct {
		pomelo.ActorBase
		isOnline bool // 玩家是否在线
		playerId int64
		uid      int64
	}
)

func (p *actorPlayer) OnInit() {
	clog.Debugf("[actorPlayer] path = %s init!", p.PathString())

	// 注册 session关闭的remote函数(网关触发连接断开后，会调用RPC发送该消息)
	p.Remote().Register("sessionClose", p.sessionClose)

	p.Local().Register("select", p.playerSelect) // 注册 查看角色
	p.Local().Register("create", p.playerCreate) // 注册 创建角色
	p.Local().Register("enter", p.playerEnter)   // 注册 进入角色
}

func (p *actorPlayer) OnStop() {
	clog.Debugf("[actorPlayer] path = %s exit!", p.PathString())
}

// sessionClose 接收角色session关闭处理
func (p *actorPlayer) sessionClose() {
	online.UnBindPlayer(p.uid)
	p.isOnline = false
	p.Exit()

	logoutEvent := event.NewPlayerLogout(p.ActorID(), p.playerId)
	p.PostEvent(&logoutEvent)
}

// playerSelect 玩家查询角色列表
func (p *actorPlayer) playerSelect(session *cproto.Session, _ *pb.None) {
	response := &pb.PlayerSelectResponse{}

	playerId := db.GetPlayerIdWithUID(session.Uid)
	if playerId > 0 {
		// 游戏设定单服单角色，协议设计成可返回多角色
		playerTable, found := db.GetPlayerTable(playerId)
		if found {
			playerInfo := buildPBPlayer(playerTable)
			response.List = append(response.List, &playerInfo)
		}
	}

	p.Response(session, response)
}

// playerCreate 玩家创角
func (p *actorPlayer) playerCreate(session *cproto.Session, req *pb.PlayerCreateRequest) {
	if req.Gender > 1 {
		p.ResponseCode(session, code.PlayerCreateFail)
		return
	}

	// 检查玩家昵称
	if len(req.PlayerName) < 1 {
		p.ResponseCode(session, code.PlayerCreateFail)
		return
	}

	// 帐号是否已经在当前游戏服存在角色
	if db.GetPlayerIdWithUID(session.Uid) > 0 {
		p.ResponseCode(session, code.PlayerCreateFail)
		return
	}

	// 获取创角初始化配置
	playerInitRow, found := data.PlayerInitConfig.Get(req.Gender)
	if found == false {
		p.ResponseCode(session, code.PlayerCreateFail)
		return
	}

	// 创建角色&添加角色初始的资产
	serverId := session.GetInt32(sessionKey.ServerID)
	newPlayerTable, errCode := db.CreatePlayer(session, req.PlayerName, serverId, playerInitRow)
	if code.IsFail(errCode) {
		p.ResponseCode(session, errCode)
		return
	}

	// TODO 更新最后一次登陆的角色信息到中心节点

	// 抛出角色创建事件
	playerCreateEvent := event.NewPlayerCreate(newPlayerTable.PlayerId, req.PlayerName, req.Gender)
	p.PostEvent(&playerCreateEvent)

	playerInfo := buildPBPlayer(newPlayerTable)
	response := &pb.PlayerCreateResponse{
		Player: &playerInfo,
	}

	p.Response(session, response)
}

// playerEnter 玩家进入游戏
func (p *actorPlayer) playerEnter(session *cproto.Session, req *pb.Int64) {
	playerId := req.Value
	if playerId < 1 {
		p.ResponseCode(session, code.PlayerIdError)
		return
	}

	// 检查并查找该用户下的该角色
	playerTable, found := db.GetPlayerTable(req.GetValue())
	if found == false {
		p.ResponseCode(session, code.PlayerIdError)
		return
	}

	// 保存进入游戏的玩家对应的agentPath
	online.BindPlayer(playerId, playerTable.UID, session.AgentPath)

	// 设置网关节点session的PlayerID属性
	p.Call(session.ActorPath(), "setSession", &pb.StringKeyValue{
		Key:   sessionKey.PlayerID,
		Value: cstring.ToString(playerId),
	})

	p.uid = playerTable.UID
	p.playerId = playerTable.PlayerId
	p.isOnline = true // 设置为在线状态

	// 这里改为客户端主动请求更佳
	// [01]推送角色 道具数据
	//module.Item.ListPush(session, playerId)
	// [02]推送角色 英雄数据
	//module.Hero.ListPush(session, playerId)
	// [03]推送角色 成就数据
	//module.Achieve.CheckNewAndPush(playerId, true, true)
	// [04]推送角色 邮件数据
	//module.Mail.ListPush(session, playerId)

	// [99]最后推送 角色进入游戏响应结果
	response := &pb.PlayerEnterResponse{}
	response.GuideMaps = map[int32]int32{}

	p.Response(session, response)

	// 角色登录事件
	loginEvent := event.NewPlayerLogin(p.ActorID(), playerId)
	p.PostEvent(&loginEvent)
}

func buildPBPlayer(playerTable *db.PlayerTable) pb.Player {
	return pb.Player{
		PlayerId:   playerTable.PlayerId,
		PlayerName: playerTable.Name,
		Level:      playerTable.Level,
		CreateTime: playerTable.CreateTime,
		Exp:        playerTable.Exp,
		Gender:     playerTable.Gender,
	}
}
