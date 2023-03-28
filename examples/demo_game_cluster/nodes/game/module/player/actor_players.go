package player

import (
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/event"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/nodes/game/module/online"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	"time"
)

type (
	// ActorPlayers 玩家总管理actor
	ActorPlayers struct {
		pomelo.ActorBase
		childExitTime time.Duration
	}
)

func (p *ActorPlayers) AliasID() string {
	return "player"
}

func (p *ActorPlayers) OnInit() {
	p.childExitTime = time.Minute * 30

	// 注册角色登陆事件
	p.Event().Register(event.PlayerLoginKey, p.onLoginEvent)
	p.Event().Register(event.PlayerLogoutKey, p.onLogoutEvent)
	p.Event().Register(event.PlayerCreateKey, p.onPlayerCreateEvent)
	p.Remote().Register("checkChild", p.checkChild)

	p.Timer().Add(1*time.Second, p.everySecondTimer)
}

func (p *ActorPlayers) OnFindChild(msg *cfacade.Message) (cfacade.IActor, bool) {
	// 动态创建 player child actor
	childID := msg.TargetPath().ChildID
	childActor, err := p.Child().Create(childID, &actorPlayer{
		isOnline: false,
	})

	if err != nil {
		return nil, false
	}

	return childActor, true
}

// cron
func (p *ActorPlayers) everySecondTimer() {
	p.Call(p.PathString(), "checkChild", nil)
}

func (p *ActorPlayers) checkChild() {
	// 扫描所有玩家actor
	p.Child().Each(func(iActor cfacade.IActor) {
		child, ok := iActor.(*actorPlayer)
		if !ok {
			return
		}

		if child.isOnline {
			return
		}

		// 玩家下线，并且超过childExitTime时间没有收发消息，则退出actor
		deadline := time.Now().Add(-p.childExitTime).Unix()
		if child.LastAt() < deadline {
			child.Exit() //actor退出
		}
	})
}

// onLoginEvent 玩家登陆事件处理
func (p *ActorPlayers) onLoginEvent(e cfacade.IEventData) {
	evt, ok := e.(*event.PlayerLogin)
	if ok == false {
		return
	}

	clog.Infof("[PlayerLoginEvent] [playerId = %d, onlineCount = %d]",
		evt.PlayerId,
		online.Count(),
	)
}

// onLoginEvent 玩家登出事件处理
func (p *ActorPlayers) onLogoutEvent(e cfacade.IEventData) {
	evt, ok := e.(event.PlayerLogout)
	if !ok {
		return
	}

	clog.Infof("[PlayerLogoutEvent] [playerId = %d, onlineCount = %d]",
		evt.PlayerId,
		online.Count(),
	)
}

// onPlayerCreateEvent 玩家创建事件
func (p *ActorPlayers) onPlayerCreateEvent(e cfacade.IEventData) {
	evt, ok := e.(*event.PlayerCreate)
	if !ok {
		return
	}

	clog.Infof("[PlayerCreateEvent] [%+v]", evt)
}

func (p *ActorPlayers) OnStop() {
	clog.Infof("onlineCount = %d", online.Count())
}
