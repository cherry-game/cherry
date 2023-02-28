package rpcGame

import (
	"fmt"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/pb"
	sessionKey "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/session_key"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

const (
	playerActor = "player"
)

const (
	sessionClose = "sessionClose"
)

// SessionClose 如果session已登录，则调用rpcGame.SessionClose() 告知游戏服
func SessionClose(app cfacade.IApplication, session *cproto.Session) {
	nodeId := session.GetString(sessionKey.ServerID)
	if nodeId == "" {
		clog.Warnf("Get server id fail. session = %s", session.Sid)
		return
	}

	targetPath := fmt.Sprintf("%s.%s.%s", nodeId, playerActor, session.Sid)
	app.ActorSystem().Call("", targetPath, sessionClose, &pb.Int64{
		Value: session.Uid,
	})

	//clog.Infof("send close session to game node. [node = %s, uid = %d]", nodeId, session.Uid)
}
