package event

type PlayerCreate struct {
	PlayerId   int64
	PlayerName string
	Gender     int32
}

func NewPlayerCreate(playerId int64, playerName string, gender int32) PlayerCreate {
	event := PlayerCreate{
		PlayerId:   playerId,
		PlayerName: playerName,
		Gender:     gender,
	}
	return event
}

func (PlayerCreate) Name() string {
	return PlayerCreateKey
}

func (p PlayerCreate) UniqueId() int64 {
	return p.PlayerId
}
