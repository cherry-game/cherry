package actor

import (
	"context"
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/cherry/components/cron"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/event"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/game/db"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/game/sessions"
	cherrySlice "github.com/cherry-game/cherry/extend/slice"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	ch "github.com/cherry-game/cherry/net/handler"
	cm "github.com/cherry-game/cherry/net/message"
	cs "github.com/cherry-game/cherry/net/session"
)

var (
	// 角色进入游戏前的，前三个协议
	beforeLoginRoutes = []string{
		"game.actorHandler.select",
		"game.actorHandler.create",
		"game.actorHandler.enter",
	}
)

type (
	Handler struct {
		ch.Handler
	}
)

func (p *Handler) Name() string {
	return "actorHandler"
}

func (p *Handler) OnInit() {
	p.everyDayTimer()

	// 注册 session关闭的remote函数(网关触发连接断开后，会调用RPC发送该消息)
	p.AddRemote("sessionClose", p.sessionClose)
	// 注册 查看角色
	p.AddLocal("select", p.selectActor)
	// 注册 创建角色
	p.AddLocal("create", p.creatActor)
	// 注册 进入角色
	p.AddLocal("enter", p.enterActor)
	// 注册角色登陆事件
	p.AddEvent(event.ActorLoginKey, p.onLoginEvent)
	p.AddEvent(event.ActorCreateKey, p.onActorCreateEvent)

	p.AddBeforeFilter(func(ctx context.Context, session *cs.Session, message *cm.Message) bool {
		actorId := sessions.GetActorId(session)

		if actorId < 1 {
			if _, found := cherrySlice.StringIn(message.Route, beforeLoginRoutes); found == false {
				statusCode := &pb.Int32{
					Value: code.ActorNotLogin,
				}
				session.Kick(statusCode, true)
				return false
			}
		}

		return true
	})
}

// everyDayTimer 每日事件定时器
func (p *Handler) everyDayTimer() {
	cherryCron.AddEveryDayFunc(func() {
		// 零点触发每日重置事件
		idList := sessions.OnlineActorIds()
		//for _, actorId := range idList {
		//p.PostEvent(event.NewEveryDayReset(actorId, false))
		//}

		cherryLogger.Debugf("actor online count = %d", len(idList))
	}, 0, 0, 0)
}

// onLoginEvent 角色登陆事件处理
func (p *Handler) onLoginEvent(e cherryFacade.IEvent) {
	loginEvent, ok := e.(*event.ActorLogin)
	if ok == false {
		return
	}

	cherryLogger.Infof("[EVENT_ACTOR_LOGIN] [actorId = %d, onlineCount = %d]",
		loginEvent.ActorId(),
		sessions.UIDCount(),
	)
}

// onActorCreateEvent 测度创建角色事件
func (p *Handler) onActorCreateEvent(e cherryFacade.IEvent) {
	actorCreateEvent, ok := e.(*event.ActorCreate)
	if ok == false {
		return
	}
	cherryLogger.Infof("actor create event value = %v", actorCreateEvent)
}

func (p *Handler) OnStop() {
	cherryLogger.Infof("onlineCount = %d", sessions.UIDCount())
}

// sessionClose 接收角色session关闭处理
func (p *Handler) sessionClose(uid *pb.Int64) {
	actorId := sessions.UnBindSession(uid.Value)
	if actorId > 0 {
		// post actor logout event
		//p.PostEvent(event.NewActorLogout(actorId))
	}

	cherryLogger.Debugf("sessionClose [uid = %d, actorId = %d, onlineCount = %d]",
		uid.Value,
		actorId,
		sessions.UIDCount(),
	)
}

// selectActor 查询角色
func (p *Handler) selectActor(ctx context.Context, session *cs.Session, _ *pb.None) {
	response := &pb.ActorSelectResponse{}

	actorId := db.GetActorIdWithUID(sessions.ServerId(), session.UID())
	if actorId > 0 {
		// 游戏设定单服单角色，协议设计成可返回多角色
		actorTable, found := db.GetActorTable(actorId)
		if found {
			actorInfo := buildPBActor(actorTable)
			response.ActorList = append(response.ActorList, &actorInfo)
		}
	}

	session.Response(ctx, response)
}

// creatActor 创建角色
func (p *Handler) creatActor(session *cs.Session, req *pb.ActorCreateRequest) (*pb.ActorCreateResponse, int32) {
	if req.Gender > 1 {
		return nil, code.ActorCreateFail
	}

	// 检查角色名
	if len(req.ActorName) < 1 {
		return nil, code.ActorCreateFail
	}

	// 帐号是否已经在当前游戏服存在角色
	actorId := db.GetActorIdWithUID(sessions.ServerId(), session.UID())
	if actorId > 0 {
		return nil, code.ActorCreateFail
	}

	// 获取创角初始化配置
	actorInit, found := data.ActorInitConfig.Get(req.Gender)
	if found == false {
		return nil, code.ActorCreateFail
	}

	// 创建角色&添加角色初始的资产
	newActorTable, errCode := db.CreateActor(session, req.ActorName, sessions.ServerId(), actorInit)
	if code.IsFail(errCode) {
		return nil, errCode
	}

	// TODO 更新最后一次登陆的角色信息到中心节点

	// 抛出角色创建事件
	p.PostEvent(event.NewActorCreate(actorId, req.ActorName, req.Gender))

	actorInfo := buildPBActor(newActorTable)
	response := &pb.ActorCreateResponse{
		ActorInfo: &actorInfo,
	}

	return response, code.OK
}

// enterActor 进入角色
func (p *Handler) enterActor(ctx context.Context, session *cs.Session, req *pb.Int64) {
	if req.GetValue() < 1 { // actorId
		sessions.ResponseCode(ctx, session, code.ActorIdError)
		return
	}

	// 检查并查找该用户下的该角色
	actorTable, found := db.GetActorTable(req.GetValue())
	if found == false {
		sessions.ResponseCode(ctx, session, code.ActorIdError)
		return
	}

	actorId := req.Value

	// actorId bind session
	sessions.BindSession(actorTable.ActorId, session)

	// 这里改为客户端主动请求更佳，手游随时断线
	// [01]推送角色 道具数据
	//module.Item.ListPush(session, actorId)
	// [02]推送角色 英雄数据
	//module.Hero.ListPush(session, actorId)
	// [03]推送角色 成就数据
	//module.Achieve.CheckNewAndPush(actorId, true, true)
	// [04]推送角色 邮件数据
	//module.Mail.ListPush(session, actorId)

	// [99]最后推送 角色进入游戏响应结果
	response := &pb.ActorEnterResponse{}
	response.GuideMaps = make(map[int32]int32)
	response.GuideMaps[0] = 0
	session.Response(ctx, response)

	// 角色登录事件
	cherry.PostEvent(event.NewActorLogin(actorId))
}

func buildPBActor(actor *db.ActorTable) pb.Actor {
	return pb.Actor{
		ActorId:    actor.ActorId,
		ActorName:  actor.Name,
		ActorLevel: actor.Level,
		CreateTime: actor.CreateTime,
		ActorExp:   actor.Exp,
		Gender:     actor.Gender,
	}
}
