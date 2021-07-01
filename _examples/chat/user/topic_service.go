package user

import (
	"github.com/cherry-game/cherry/_examples/chat/proto"
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/session"
	"strings"
	"sync/atomic"
)

type User struct {
	session  *cherrySession.Session
	nickname string
	gateId   int64
	masterId int64
	balance  int64
	message  int
}

var (
	nextUid int64
	users   = make(map[int64]*User)
)

func NewUser(s *cherrySession.Session, nickName string) error {
	atomic.AddInt64(&nextUid, 1)

	uid := nextUid
	if err := s.Bind(uid); err != nil {
		return err
	}

	var members []string
	for _, u := range users {
		members = append(members, u.nickname)
	}

	err := s.Push("onMembers", &proto.ExistsMembersResponse{Members: strings.Join(members, ",")})
	if err != nil {
		return err
	}

	user := &User{
		session:  s,
		nickname: nickName,
		gateId:   uid,
		masterId: uid,
		balance:  1000,
	}

	users[uid] = user

	chat := &proto.JoinRoomRequest{
		Nickname: nickName,
	}

	return s.RPC("gate.roomHandler.joinRoom", chat)
}

func Stats(s *cherrySession.Session, uid int64) error {
	// It's OK to use map without lock because of this service running in main thread
	user, found := users[uid]
	if !found {
		return cherryError.Errorf("User not found: %v", uid)
	}

	user.message++
	user.balance--
	return s.Push("onBalance", &proto.UserBalanceResponse{CurrentBalance: user.balance})
}

func Disconnect(s *cherrySession.Session) (next bool) {
	if s.UID() < 1 {
		return true
	}

	uid := s.UID()
	delete(users, uid)

	cherryLogger.Infof("User session disconnected UID[%d]", s.UID())

	return true
}
