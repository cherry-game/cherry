package cherryDiscovery

import (
	facade "github.com/cherry-game/cherry/facade"
)

// DiscoveryEtcd etcd方式发现服务
type DiscoveryEtcd struct {
	DiscoveryDefault
}

func (p *DiscoveryEtcd) Name() string {
	return "etcd"
}

func (p *DiscoveryEtcd) Init(_ facade.IApplication) {

}

func (p *DiscoveryEtcd) OnStop() {

}
