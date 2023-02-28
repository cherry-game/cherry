package main

import (
	"github.com/cherry-game/cherry/examples/demo_chat/protocol"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
	"sync/atomic"
)

var (
	_nextUID int64
)

func nextUID() int64 {
	return atomic.AddInt64(&_nextUID, 1)
}

type (

	// actorUser 每建立一个新的客户端连接，系统会创建一个新的actorUser进行关联
	actorUser struct {
		cactor.Base
		agent *pomelo.Agent // 新连接建立时对应的agent对象
	}
)

func newActorUser(agent *pomelo.Agent) actorUser {
	newUser := actorUser{
		agent: agent,
	}
	agent.AddOnClose(newUser.onClose)
	return newUser
}

func (h *actorUser) OnInit() {
	h.Local().Register("login", h.login)
}

func (h *actorUser) login(session *cproto.Session, req *protocol.LoginRequest) {
	clog.Debugf("nickname = %s", req.Nickname)

	if session.IsBind() {
		return
	}

	req.UID = nextUID()
	h.agent.Bind(req.UID)

	// 通知room actor，有新用户加入
	h.Call(".room", "join", req)

	h.agent.Response(session, &protocol.LoginResponse{})
}

func (h *actorUser) onClose(agent *pomelo.Agent) {
	session := agent.Session()
	if !session.IsBind() {
		return
	}

	// 发送玩家断开连接的消息给room actor
	req := &protocol.Int64{
		Value: session.Uid,
	}

	h.Call(".room", "exit", req)

	clog.Debugf("[sid = %s,uid = %d] session disconnected.", session.Sid, session.Uid)
}
