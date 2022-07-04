package cherryCluster

import (
	cconst "github.com/cherry-game/cherry/const"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cdiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cnats "github.com/cherry-game/cherry/net/cluster/nats"
)

type Component struct {
	cfacade.Component
	cfacade.RPCClient
	cfacade.RPCServer
}

func NewComponent() *Component {
	return &Component{}
}

func (c *Component) Name() string {
	return cconst.ClusterComponent
}

func (c *Component) Init() {
	cnats.Init()

	c.RPCClient = NewRPCClient(c)

	server := NewNatsRPCServer(c, c.RPCClient, 32767)
	server.Init()
	c.RPCServer = server

	// init discovery
	cdiscovery.Init(c.App())
}

func (c *Component) OnStop() {
	clog.Infof("cluster component stopping.")
	c.RPCClient.OnStop()
	c.RPCServer.OnStop()
	cdiscovery.OnStop()
	clog.Infof("cluster component on stop.")
}
