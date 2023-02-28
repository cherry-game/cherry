package event

type PlayerLogin struct {
	BaseEvent
	PlayerId        int64
	LoginDays       int //累计登录天数(每天+1)
	ContinueDays    int //连续登录天数(超过一天则归零)
	MaxContinueDays int //最大连续登录天数
}

func NewPlayerLogin(playerId int64) PlayerLogin {
	event := PlayerLogin{
		BaseEvent: BaseEvent{
			name: PlayerLoginKey,
		},
		PlayerId: playerId,
	}
	return event
}
