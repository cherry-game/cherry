package cherryActor

import cfacade "github.com/cherry-game/cherry/facade"

var (
	Name = "actor_component"
)

type Component struct {
	cfacade.Component
	*System
	actorHandlers []cfacade.IActorHandler
}

func New() *Component {
	return &Component{
		System: NewSystem(),
	}
}

func (c *Component) Name() string {
	return Name
}

func (c *Component) Init() {
	c.System.SetApp(c.App())
}

func (c *Component) OnAfterInit() {
	// Register actor
	for _, actor := range c.actorHandlers {
		c.CreateActor(actor.AliasID(), actor)
	}
}

func (c *Component) OnStop() {
	c.System.Stop()
}

func (c *Component) Add(actors ...cfacade.IActorHandler) {
	c.actorHandlers = append(c.actorHandlers, actors...)
}
