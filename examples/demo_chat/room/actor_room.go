package main

import (
	"fmt"
	"github.com/cherry-game/cherry/examples/demo_chat/protocol"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
	"sync/atomic"
	"time"
)

var (
	_nextUID int64
)

func newUID() int64 {
	return atomic.AddInt64(&_nextUID, 1)
}

type (
	actorRoom struct {
		cactor.Base
		userMap map[int64]*User // uid,user
	}

	User struct {
		uid      cfacade.UID
		nickname string
		balance  int64
		message  int
	}
)

func (p *actorRoom) AliasID() string {
	return "room"
}

func (p *actorRoom) OnInit() {
	p.userMap = make(map[int64]*User)

	p.Local().Register("login", p.login)
	p.Local().Register("syncMessage", p.syncMessage)

	p.Remote().Register("exit", p.exit)
}

func (*actorRoom) OnLocalReceived(_ *cfacade.Message) (bool, bool) {
	// 当接收local消息时，直接在当前actor执行(不再路由到子actor)
	return false, true
}

func (p *actorRoom) login(session *cproto.Session, req *protocol.LoginRequest) {
	clog.Debugf("nickname = %s", req.Nickname)

	if session.IsBind() {
		return
	}

	agent, found := pomelo.GetAgent(session.Sid)
	if !found {
		return
	}

	uid := newUID()
	agent.Bind(uid)

	user := &User{
		uid:      uid,
		nickname: req.Nickname,
		balance:  1000,
		message:  0,
	}

	p.userMap[uid] = user

	// 广播其他用户，有新用户进入房间
	newUserRequest := &protocol.NewUserBroadcast{
		Content: fmt.Sprintf("user join: %+v", req),
	}
	p.broadcast("onNewUser", newUserRequest)

	agent.Response(session, &protocol.LoginResponse{})
}

func (p *actorRoom) exit(req *protocol.Int64) {
	if req.Value < 1 {
		return
	}
	delete(p.userMap, req.Value)
}

func (p *actorRoom) syncMessage(session *cproto.Session, req *protocol.SyncMessage) {
	user, found := p.userMap[session.Uid]
	if !found {
		clog.Errorf("user not found: %v", session.Uid)
		return
	}

	user.message++
	user.balance--

	// 有新消息，广播给其他用户
	p.broadcast("onMessage", req)

	// 扣减当前用户的余额，并通知客户端
	agent, found := pomelo.GetAgent(session.Sid)
	if found {
		agent.Push("onBalance", &protocol.UserBalanceResponse{
			CurrentBalance: user.balance,
		})
	}

	// 测试往log节点写日志
	req.PacketTime = time.Now().UnixNano()
	rsp := &protocol.WriteResponse{}
	err := p.CallWait("log-1.log", "write", req, rsp)
	clog.Debugf("log->write() %v %v", rsp, err)
}

func (p *actorRoom) broadcast(route string, v interface{}) {
	for uid := range p.userMap {
		if agent, ok := pomelo.GetAgentWithUID(uid); ok {
			agent.Push(route, v)
		}
	}
}
