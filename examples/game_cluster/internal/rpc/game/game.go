package rpcGame

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	sessionKey "github.com/cherry-game/cherry/examples/game_cluster/internal/session_key"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

const (
	actorHandler = "game.actorHandler."
)

const (
	sessionClose = actorHandler + "sessionClose"
)

// SessionClose 如果session已登录，则调用rpcGame.SessionClose() 告知游戏服
func SessionClose(session *cherrySession.Session) {
	if session.IsBind() == false {
		return
	}

	gameNodeId := sessionKey.GetNodeId(session)
	if gameNodeId == "" {
		return
	}

	req := &pb.Int64{
		Value: session.UID(),
	}

	cherry.PublishRemote(gameNodeId, sessionClose, req)
	cherryLogger.Infof("send close session to game node. [node = %s, uid = %d]", gameNodeId, session.UID())
}
