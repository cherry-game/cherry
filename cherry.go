package cherry

import (
	cfacade "github.com/cherry-game/cherry/facade"
	cactor "github.com/cherry-game/cherry/net/actor"
	ccluster "github.com/cherry-game/cherry/net/cluster"
	cdiscovery "github.com/cherry-game/cherry/net/discovery"
)

type (
	AppBuilder struct {
		*Application
		components  []cfacade.IComponent
		actorSystem *cactor.Component
	}
)

func Configure(profileFilePath, nodeId string, isFrontend bool, mode NodeMode) *AppBuilder {
	app := NewApp(profileFilePath, nodeId, isFrontend, mode)
	appBuilder := &AppBuilder{
		Application: app,
		components:  make([]cfacade.IComponent, 0),
		actorSystem: cactor.New(),
	}

	return appBuilder
}

func (p *AppBuilder) Startup() {
	app := p.Application

	if app.NodeMode() == Cluster {
		discovery := cdiscovery.New()
		app.SetDiscovery(discovery)
		app.Register(discovery)

		cluster := ccluster.New()
		app.SetCluster(cluster)
		app.Register(cluster)
	}

	// Register custom components
	app.Register(p.components...)

	app.SetActorSystem(p.actorSystem)
	app.Register(p.actorSystem)

	if app.netParser != nil {
		for _, connector := range app.netParser.Connectors() {
			app.Register(connector)
		}
	}

	// startup
	app.Startup()
}

func (p *AppBuilder) Register(component ...cfacade.IComponent) {
	p.components = append(p.components, component...)
}

func (p *AppBuilder) AddActors(actors ...cfacade.IActorHandler) {
	p.actorSystem.Add(actors...)
}

func (p *AppBuilder) NetParser() cfacade.INetParser {
	return p.netParser
}

func (p *AppBuilder) SetNetParser(parser cfacade.INetParser) {
	p.netParser = parser
}
