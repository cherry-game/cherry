package cherryEvents

type GameEvent struct {
	Name string
	Id   string
}

func (g *GameEvent) EventName() string {
	return g.Name
}

func (g *GameEvent) SetEvent(name string) {
	g.Name = name
}

func (g *GameEvent) UniqueId() string {
	return g.Id
}

func (g *GameEvent) SetUnique(uniqueId string) {
	g.Id = uniqueId
}
