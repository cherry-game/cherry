package cherryEvent

type GameEvent struct {
	EventName string
	Id        string
}

func (g *GameEvent) Name() string {
	return g.EventName
}

func (g *GameEvent) UniqueId() string {
	return g.Id
}
