package cherryDiscovery

import (
	facade "github.com/cherry-game/cherry/facade"
)

// DiscoveryETCD etcd方式发现服务
type DiscoveryETCD struct {
	DiscoveryDefault
}

func (p *DiscoveryETCD) Name() string {
	return "etcd"
}

func (p *DiscoveryETCD) Init(_ facade.IApplication) {

}

func (p *DiscoveryETCD) OnStop() {

}
