package event

type PlayerCreate struct {
	BaseEvent
	PlayerId   int64
	PlayerName string
	Gender     int32
}

func NewPlayerCreate(playerId int64, playerName string, gender int32) PlayerCreate {
	event := PlayerCreate{
		BaseEvent: BaseEvent{
			name: PlayerCreateKey,
		},
		PlayerId:   playerId,
		PlayerName: playerName,
		Gender:     gender,
	}
	return event
}
