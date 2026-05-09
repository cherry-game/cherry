package cherry

import (
	cfacade "github.com/cherry-game/cherry/facade"
	ccluster "github.com/cherry-game/cherry/net/cluster"
	cdiscovery "github.com/cherry-game/cherry/net/discovery"
	cherryDiscovery "github.com/cherry-game/cherry/net/discovery"
)

type (
	AppBuilder struct {
		*Application
		components []cfacade.IComponent
	}
)

func Configure(profileFilePath, nodeID string, isFrontend bool, mode NodeMode) *AppBuilder {
	appBuilder := &AppBuilder{
		Application: NewApp(profileFilePath, nodeID, isFrontend, mode),
		components:  make([]cfacade.IComponent, 0),
	}

	return appBuilder
}

func ConfigureNode(node cfacade.INode, isFrontend bool, mode NodeMode) *AppBuilder {
	appBuilder := &AppBuilder{
		Application: NewAppNode(node, isFrontend, mode),
		components:  make([]cfacade.IComponent, 0),
	}

	return appBuilder
}

func (p *AppBuilder) Startup() {
	app := p.Application

	if app.NodeMode() == Cluster {
		if app.cluster == nil {
			// Set deafult nats cluster
			app.cluster = ccluster.New()
		}
		app.Register(app.cluster)

		// Obtain the discovery service according to the configured mode
		app.discovery = cdiscovery.New()
		app.Register(app.discovery)
	}

	// Register custom components
	app.Register(p.components...)

	// startup
	app.Startup()
}

func (p *AppBuilder) Register(component ...cfacade.IComponent) {
	p.components = append(p.components, component...)
}

func (p *AppBuilder) AddActors(actors ...cfacade.IActorHandler) {
	p.actorSystem.Add(actors...)
}

func (p *AppBuilder) SetSerializer(serializer cfacade.ISerializer) {
	if serializer == nil {
		return
	}

	p.serializer = serializer
}

func (p *AppBuilder) SetDiscovery(discovery cfacade.IDiscoveryComponent) {
	if discovery == nil {
		return
	}

	cherryDiscovery.Register(discovery)
}

func (a *AppBuilder) SetCluster(cluster cfacade.IClusterComponent) {
	if cluster == nil {
		return
	}

	a.cluster = cluster
}

func (a *AppBuilder) SetNetParser(netParser cfacade.INetParser) {
	if a.Running() || netParser == nil {
		return
	}

	a.netParser = netParser
}
