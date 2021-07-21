package cherryCluster

import (
	cherryConst "github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProfile "github.com/cherry-game/cherry/profile"
)

// Component 集群组件
// rpc server
// rpc client
// session 管理
// 节点获取
type Component struct {
	facade.Component
	mode      string
	discovery facade.IDiscovery
	rpcClient *rpcClient
}

func NewComponent() *Component {
	return &Component{}
}

func (c *Component) Name() string {
	return cherryConst.ClusterComponent
}

func (c *Component) Init() {
	clusterJson := cherryProfile.Config().Get("cluster")
	if clusterJson.LastError() != nil {
		cherryLogger.Error("`cluster` property not found in profile file.")
		return
	}

	discoveryJson := clusterJson.Get("discovery")
	if discoveryJson.LastError() != nil {
		cherryLogger.Error("`discovery` property not found in profile file.")
		return
	}

	c.mode = discoveryJson.Get("mode").ToString()
	if c.mode == "" {
		cherryLogger.Error("`discovery->mode` property not found in profile file.")
		return
	}

	c.discovery = GetDiscovery(c.mode)
	if c.discovery == nil {
		cherryLogger.Errorf("not found mode = %s property in discovery map.", c.mode)
		return
	}

	c.discovery.Init(c.App(), discoveryJson)
}

func (c *Component) OnStop() {

}

func (c *Component) Discovery() facade.IDiscovery {
	return c.discovery
}

func (c *Component) OnAddMember(listener facade.MemberListener) {
	c.Discovery().OnAddMember(listener)
}

func (c *Component) OnRemoveMember(listener facade.MemberListener) {
	c.Discovery().OnRemoveMember(listener)
}
