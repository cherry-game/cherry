package main

import (
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/net/session"
	"strings"
	"sync/atomic"
)

type user struct {
	session  *cherrySession.Session
	nickname string
	gateId   int64
	masterId int64
	balance  int64
	message  int
}

var (
	nextUid int64
	users   = make(map[int64]*user)
)

func newUser(s *cherrySession.Session, nickName string) error {
	atomic.AddInt64(&nextUid, 1)

	uid := nextUid
	if err := s.Bind(uid); err != nil {
		return err
	}

	var members []string
	for _, u := range users {
		members = append(members, u.nickname)
	}

	err := s.Push("onMembers", &existsMembersResponse{Members: strings.Join(members, ",")})
	if err != nil {
		return err
	}

	user := &user{
		session:  s,
		nickname: nickName,
		gateId:   uid,
		masterId: uid,
		balance:  1000,
	}

	users[uid] = user

	chat := &joinRoomRequest{
		Nickname: nickName,
	}

	return s.RPC("gate.roomHandler.joinRoom", chat)
}

func stats(s *cherrySession.Session, uid int64) error {
	// It's OK to use map without lock because of this service running in main thread
	user, found := users[uid]
	if !found {
		return cherryError.Errorf("user not found: %v", uid)
	}

	user.message++
	user.balance--
	return s.Push("onBalance", &userBalanceResponse{CurrentBalance: user.balance})
}

func disconnect(s *cherrySession.Session) (next bool) {
	if s.UID() < 1 {
		return true
	}

	uid := s.UID()
	delete(users, uid)

	s.Debug("user session disconnected.")

	return true
}
