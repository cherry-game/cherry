package event

type PlayerLogout struct {
	ActorId  string // actor id
	PlayerId int64  // player id
}

func NewPlayerLogout(actorId string, playerId int64) PlayerLogout {
	event := PlayerLogout{
		ActorId:  actorId,
		PlayerId: playerId,
	}
	return event
}

func (PlayerLogout) Name() string {
	return PlayerLogoutKey
}

func (p PlayerLogout) UniqueId() int64 {
	return p.PlayerId
}
