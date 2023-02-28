package cherryActor

import (
	cherryMap "github.com/cherry-game/cherry/extend/map"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"strings"
)

type actorChild struct {
	thisActor   *Actor
	childActors *cherryMap.Map[string, *Actor] // key:childActorId, value:*actor
}

func newChild(thisActor *Actor) actorChild {
	return actorChild{
		thisActor:   thisActor,
		childActors: cherryMap.NewMap[string, *Actor](true),
	}
}

func (p *actorChild) onStop() {
	for _, id := range p.childActors.Keys() {
		if childActor, found := p.childActors.Get(id); found {
			p.childActors.Remove(id)
			childActor.Exit()
		}
	}

	//p.childActors = nil
	p.thisActor = nil
}

func (p *actorChild) Create(childID string, handler cfacade.IActorHandler) (cfacade.IActor, error) {
	if p.thisActor.path.IsChild() {
		return nil, ErrForbiddenCreateChildActor
	}

	if strings.TrimSpace(childID) == "" {
		return nil, ErrActorIDIsNil
	}

	if thisActor, found := p.childActors.Get(childID); found {
		return thisActor, nil
	}

	childActor, err := newActor(p.thisActor.ActorID(), childID, handler, p.thisActor.system)
	if err != nil {
		return nil, err
	}

	p.childActors.Put(childID, &childActor)
	go childActor.run()

	return &childActor, nil
}

func (p *actorChild) Get(childID string) (cfacade.IActor, bool) {
	return p.childActors.Get(childID)
}

func (p *actorChild) Remove(childID string) {
	_, found := p.childActors.Remove(childID)
	if !found {
		clog.Warnf("[Remove] [childID = %s] get child Actor fail. ", childID)
	}
}

func (p *actorChild) Each(fn func(cfacade.IActor)) {
	for _, id := range p.childActors.Keys() {
		thisActor, found := p.childActors.Get(id)
		if found {
			fn(thisActor)
		}
	}
}
