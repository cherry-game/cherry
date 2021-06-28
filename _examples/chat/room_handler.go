package chat

import (
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	session "github.com/cherry-game/cherry/net/session"
)

type Room struct {
}

type RoomHandler struct {
	cherryHandler.Handler
	rooms map[int]*Room
}

func (r *RoomHandler) Name() string {
	return "RoomService"
}

func (r *RoomHandler) OnInit() {
	r.RegisterLocal("SyncMessage", r.SyncMessage)
}

func (r *RoomHandler) SyncMessage(s *session.Session, msg *SyncMessage) error {
	// Send an RPC to master server to stats
	if err := s.RPC("TopicService.Stats", &MasterStats{Uid: s.UID()}); err != nil {
		return err
	}

	// Sync message to all members in this room
	//return rs.group.Broadcast("onMessage", msg)
	return nil
}
