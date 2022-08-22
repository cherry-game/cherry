package main

import (
	"fmt"
	clog "github.com/cherry-game/cherry/logger"
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
	group   = cherrySession.NewGroup("all-users")
)

func joinRoom(session *cherrySession.Session, req *joinRoomRequest) error {
	broadcast := &newUserBroadcast{
		Content: fmt.Sprintf("user join: %v", req.Nickname),
	}

	if err := group.Broadcast("onNewUser", broadcast); err != nil {
		return err
	}

	return group.Add(session)
}

func newUser(s *cherrySession.Session, nickName string) error {
	atomic.AddInt64(&nextUid, 1)

	uid := nextUid
	if err := cherrySession.Bind(s.SID(), uid); err != nil {
		return err
	}

	var members []string
	for _, u := range users {
		members = append(members, u.nickname)
	}

	s.Push("onMembers", &existsMembersResponse{Members: strings.Join(members, ",")})

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

	return joinRoom(s, chat)
}

func stats(s *cherrySession.Session, uid int64) {
	// It's OK to use map without lock because of this service running in main thread
	user, found := users[uid]
	if !found {
		clog.Errorf("user not found: %v", uid)
		return
	}

	user.message++
	user.balance--

	s.Push("onBalance", &userBalanceResponse{CurrentBalance: user.balance})

	return
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
