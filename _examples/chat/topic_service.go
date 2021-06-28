package chat

import (
	session "github.com/cherry-game/cherry/net/session"
	"strings"
	"sync/atomic"
)

type User struct {
	session  *session.Session
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

func NewUser(s *session.Session, msg *NewUserRequest) error {
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
		nickname: msg.Nickname,
		gateId:   msg.GateUid,
		masterId: uid,
		balance:  1000,
	}

	users[uid] = user

	//chat := &JoinRoomRequest{
	//	Nickname:  msg.Nickname,
	//	GateUid:   msg.GateUid,
	//	MasterUid: uid,
	//}

	return nil //s.RPC("RoomService.JoinRoom", chat)
}
