package player

import (
	cherryCron "github.com/cherry-game/cherry/components/cron"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/event"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/nodes/game/module/online"
	cfacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

type (
	// ActorPlayers 玩家总管理actor
	ActorPlayers struct {
		pomelo.ActorBase
	}
)

func (p *ActorPlayers) AliasID() string {
	return "player"
}

func (p *ActorPlayers) OnInit() {
	p.everyDayTimer()

	// 注册角色登陆事件
	p.Event().Register(event.PlayerLoginKey, p.onLoginEvent)
	p.Event().Register(event.PlayerCreateKey, p.onPlayerCreateEvent)
}

func (p *ActorPlayers) OnFindChild(msg *cfacade.Message) (cfacade.IActor, bool) {
	// 动态创建 player child actor
	childID := msg.TargetPath().ChildID
	childActor, err := p.Child().Create(childID, &actorPlayer{
		isOnline: true,
	})

	if err != nil {
		return nil, false
	}

	//记录childID，待玩家下线后根据策略销毁player child actor

	return childActor, true
}

// everyDayTimer 每日事件定时器
func (p *ActorPlayers) everyDayTimer() {
	cherryCron.AddEveryDayFunc(func() {
		// 零点触发每日重置事件
		//idList := sessions.OnlinePlayerIds()
		//for _, playerId := range idList {
		//p.Event().Post(event.NewEveryDayReset(playerId, false))
		//}
		//cherryLogger.Debugf("player online count = %d", len(idList))
	}, 0, 0, 0)
}

// onLoginEvent 玩家角色登陆事件处理
func (p *ActorPlayers) onLoginEvent(e cfacade.IEventData) {
	loginEvent, ok := e.(*event.PlayerLogin)
	if ok == false {
		return
	}

	cherryLogger.Infof("[onLoginEvent] [playerId = %d, onlineCount = %d]",
		loginEvent.PlayerId,
		online.Count(),
	)
}

// onPlayerCreateEvent 测度创建玩家事件
func (p *ActorPlayers) onPlayerCreateEvent(e cfacade.IEventData) {
	playerCreateEvent, ok := e.(*event.PlayerCreate)
	if ok == false {
		return
	}
	cherryLogger.Infof("player create event value = %v", playerCreateEvent)
}

func (p *ActorPlayers) OnStop() {
	cherryLogger.Infof("onlineCount = %d", online.Count())
}
