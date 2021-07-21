package cherryCluster

import (
	facade "github.com/cherry-game/cherry/facade"
)

var (
	discoveryMap = make(map[string]facade.IDiscovery)
)

func init() {
	RegisterDiscovery(&DiscoveryNode{})
}

func RegisterDiscovery(discovery facade.IDiscovery) {
	if discovery == nil {
		panic("discovery is nil")
	}

	if discovery.Name() == "" {
		panic("discovery name is empty.")
	}

	discoveryMap[discovery.Name()] = discovery
}

func GetDiscovery(name string) facade.IDiscovery {
	return discoveryMap[name]
}
