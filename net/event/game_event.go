package cherryEvent

import cherryString "github.com/cherry-game/cherry/extend/string"

type GameEvent struct {
	EventName string
	Id        int64
}

func (g *GameEvent) Name() string {
	return g.EventName
}

func (g *GameEvent) UniqueId() string {
	return cherryString.Int64ToString(g.Id)
}

func (g *GameEvent) ActorId() int64 {
	return g.Id
}
