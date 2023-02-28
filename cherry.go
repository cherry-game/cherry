package cherry

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	ccluster "github.com/cherry-game/cherry/net/cluster"
	cdiscovery "github.com/cherry-game/cherry/net/discovery"
)

type (
	AppBuilder struct {
		*Application
		components []cfacade.IComponent
		netParser  cfacade.INetParser
	}
)

func Configure(profileFilePath, nodeId string, isFrontend bool, mode NodeMode) *AppBuilder {
	app := NewApp(profileFilePath, nodeId, isFrontend, mode)
	appBuilder := &AppBuilder{
		Application: app,
		components:  make([]cfacade.IComponent, 0),
		netParser:   nil,
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

	if p.isFrontend {
		if p.netParser == nil {
			clog.Panic("gate is nil.")
		}
		p.netParser.Load(app)
	}

	app.Startup()
}

func (p *AppBuilder) Register(component ...cfacade.IComponent) {
	p.components = append(p.components, component...)
}

func (p *AppBuilder) AddActors(actorHandlers ...cfacade.IActorHandler) {
	for _, handler := range actorHandlers {
		p.actorSystem.CreateActor(handler.AliasID(), handler)
	}
}

func (p *AppBuilder) NetParser() cfacade.INetParser {
	return p.netParser
}

func (p *AppBuilder) SetNetParser(parser cfacade.INetParser) {
	p.netParser = parser
}

func (p *AppBuilder) SetActorInvoke(local cfacade.InvokeFunc, remote cfacade.InvokeFunc) {
	p.actorSystem.SetLocalInvoke(local)
	p.actorSystem.SetRemoteInvoke(remote)
}
