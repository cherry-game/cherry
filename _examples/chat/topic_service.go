package chat

import (
	cherryError "github.com/cherry-game/cherry/error"
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

func NewUser(s *cherrySession.Session, req *NewUserRequest) error {
	atomic.AddInt64(&nextUid, 1)

	uid := nextUid
	if err := s.Bind(uid); err != nil {
		return err
	}

	var members []string
	for _, u := range users {
		members = append(members, u.nickname)
	}

	err := s.Push("onMembers", &ExistsMembersResponse{Members: strings.Join(members, ",")})
	if err != nil {
		return err
	}

	user := &User{
		session:  s,
		nickname: req.Nickname,
		gateId:   req.GateUid,
		masterId: uid,
		balance:  1000,
	}

	users[uid] = user

	chat := &JoinRoomRequest{
		Nickname:  req.Nickname,
		GateUid:   req.GateUid,
		MasterUid: uid,
	}

	return s.RPC("gate.roomHandler.joinRoom", chat)
}

func Stats(s *cherrySession.Session, msg *MasterStats) error {
	// It's OK to use map without lock because of this service running in main thread
	user, found := users[msg.Uid]
	if !found {
		return cherryError.Errorf("User not found: %v", msg.Uid)
	}

	user.message++
	user.balance--
	return s.Push("onBalance", &UserBalanceResponse{user.balance})
}
