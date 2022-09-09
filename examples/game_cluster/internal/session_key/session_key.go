package sessionKey

import cs "github.com/cherry-game/cherry/net/session"

const (
	ServerID     = "server_id"     // int32 游戏服务器ID
	OpenID       = "open_id"       // string 第三方登陆sdk的用户唯一标识
	PlatformType = "platform_type" // int32 平台类型id
	PID          = "pid"           // int32 平台id
	ActorID      = "actor_id"      // int64 角色id
	UID          = "uid"           // int64 用户id
)

func GetNodeId(session *cs.Session) string {
	return session.GetString(ServerID)
}

func GetServerId(session *cs.Session) int32 {
	return session.GetInt32(ServerID)
}

func GetOpenId(session *cs.Session) string {
	return session.GetString(OpenID)
}

func GetPID(session *cs.Session) int32 {
	return session.GetInt32(PID)
}
