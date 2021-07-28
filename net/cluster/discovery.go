package cherryCluster

import (
	"fmt"
	facade "github.com/cherry-game/cherry/facade"
)

var (
	discoveryMap = make(map[string]facade.IDiscovery)
)

func init() {
	RegisterDiscovery(&DiscoveryDefault{})
	RegisterDiscovery(&DiscoveryMaster{})
}

func RegisterDiscovery(discovery facade.IDiscovery) {
	if discovery == nil {
		panic("discovery is nil")
	}

	if discovery.Name() == "" {
		panic(fmt.Sprintf("discovery name is empty. %T", discovery))
	}

	discoveryMap[discovery.Name()] = discovery
}

func GetDiscovery(name string) facade.IDiscovery {
	return discoveryMap[name]
}
