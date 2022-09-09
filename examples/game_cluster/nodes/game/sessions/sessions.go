package sessions

import (
	"context"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	cherryString "github.com/cherry-game/cherry/extend/string"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryContext "github.com/cherry-game/cherry/net/context"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"sync"
	"sync/atomic"
)

var (
	currentServerId int32 = 0 // 当前游戏节点的serverId
)

var (
	actorIdMap        = sync.Map{} // key:actorId, value:*Session
	uidMap            = sync.Map{} // key:UID, value:actorId
	onlineCount int64 = 0          // online account count
)

func SetServerId(node string) {
	id, ok := cherryString.ToInt32(node)
	if ok == false {
		panic("node id value is not int32")
	}

	currentServerId = id
}

func ServerId() int32 {
	return currentServerId
}

func BindSession(actorId int64, session *cherrySession.Session) {
	if actorId > 0 && session != nil {
		actorIdMap.Store(actorId, session)
		uidMap.Store(session.UID(), actorId)
		atomic.AddInt64(&onlineCount, 1)
	}
}

func GetActorId(session *cherrySession.Session) int64 {
	return GetActorIdByUID(session.UID())
}

func GetActorIdByUID(uid cherryFacade.UID) int64 {
	actorIdValue, found := uidMap.Load(uid)
	if found == false {
		return 0
	}

	return actorIdValue.(int64)
}

func GetSession(actorId int64) (*cherrySession.Session, bool) {
	sessionValue, found := actorIdMap.Load(actorId)
	if found == false {
		return nil, false
	}

	return sessionValue.(*cherrySession.Session), true
}

func UnBindSession(uid cherryFacade.UID) int64 {
	actorIdValue, found := uidMap.LoadAndDelete(uid)
	if found == false {
		return 0
	}

	actorId := actorIdValue.(int64)
	actorIdMap.Delete(actorId)

	atomic.AddInt64(&onlineCount, -1)

	return actorId
}

func OnlineActorIds() []int64 {
	var list []int64

	actorIdMap.Range(func(key, value interface{}) bool {
		actorId, ok := key.(int64)
		if ok {
			list = append(list, actorId)
		}
		return true
	})

	return list
}

func UIDCount() int64 {
	return onlineCount
}

func Push(actorId int64, route string, v interface{}) {
	session, found := GetSession(actorId)
	if found == false {
		cherryLogger.Debugf("push fail,session not found. [actorId = %d, route = %s, val = %+v]",
			actorId, route, v,
		)
		return
	}

	session.Push(route, v)
}

func ResponseCode(ctx context.Context, session *cherrySession.Session, statusCode int32) {
	msgId := cherryContext.GetMessageId(ctx)
	isError := false
	if code.IsFail(statusCode) {
		isError = true
	}

	req := &pb.Int32{
		Value: statusCode,
	}
	session.ResponseMID(msgId, req, isError)
}
