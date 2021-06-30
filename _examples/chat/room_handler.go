package chat

import (
	"fmt"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
)

type RoomHandler struct {
	cherryHandler.Handler
	group *cherrySession.Group
}

func (r *RoomHandler) Name() string {
	return "roomHandler"
}

func (r *RoomHandler) OnInit() {
	r.group = cherrySession.NewGroup("all-users")

	r.RegisterLocal("joinRoom", r.joinRoom)
	r.RegisterLocal("syncMessage", r.syncMessage)

	r.AddOnClose(r.userDisconnected)
}

func (r *RoomHandler) joinRoom(session *cherrySession.Session, _ *cherryMessage.Message, req *JoinRoomRequest) error {
	if err := session.Bind(req.MasterUid); err != nil {
		return err
	}

	broadcast := &NewUserBroadcast{
		Content: fmt.Sprintf("user join: %v", req.Nickname),
	}

	if err := r.group.Broadcast("onNewUser", broadcast); err != nil {
		return err
	}

	return r.group.Add(session)
}

func (r *RoomHandler) syncMessage(s *cherrySession.Session, _ *cherryMessage.Message, req *SyncMessage) error {
	// Send an RPC to master server to stats
	if err := Stats(s, &MasterStats{Uid: s.UID()}); err != nil {
		return err
	}

	// Sync message to all members in this room
	return r.group.Broadcast("onMessage", req)
}

func (r *RoomHandler) userDisconnected(session *cherrySession.Session) (next bool) {
	if err := r.group.Leave(session); err != nil {
		session.Debugf("Remove user from group failed. err=%s", err)
		return true
	}

	session.Debug("User session disconnected")
	return true
}
