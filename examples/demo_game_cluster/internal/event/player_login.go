package event

type PlayerLogin struct {
	ActorId         string // actor id
	PlayerId        int64  // player id
	LoginDays       int    // 累计登录天数(每天+1)
	ContinueDays    int    // 连续登录天数(超过一天则归零)
	MaxContinueDays int    // 最大连续登录天数
}

func NewPlayerLogin(actorId string, playerId int64) PlayerLogin {
	event := PlayerLogin{
		ActorId:  actorId,
		PlayerId: playerId,
	}
	return event
}

func (PlayerLogin) Name() string {
	return PlayerLoginKey
}

func (p PlayerLogin) UniqueId() int64 {
	return p.PlayerId
}
