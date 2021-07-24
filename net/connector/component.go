package cherryConnector

import (
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
)

type (
	// Component (连接器组件适用于前端节点)
	Component struct {
		facade.Component
		ConnectStat *ConnectStat
		connector   facade.IConnector
	}
)

func NewComponent(connector facade.IConnector) *Component {
	return &Component{
		ConnectStat: &ConnectStat{},
		connector:   connector,
	}
}

func (c *Component) Name() string {
	return cherryConst.ConnectorComponent
}

func (c *Component) OnConnect(listener ...facade.OnConnectListener) {
	c.connector.OnConnect(listener...)
}

func (c *Component) OnStart() {
}

func (c *Component) OnAfterInit() {
	if c.connector != nil {
		go c.connector.OnStart()
	}
}

func (c *Component) OnStop() {
	if c.connector != nil {
		c.connector.OnStop()
	}
}
