package cherryDiscovery

import (
	"fmt"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProfile "github.com/cherry-game/cherry/profile"
)

var (
	discoveryMap  = make(map[string]facade.IDiscovery)
	isInit        = false
	thisDiscovery facade.IDiscovery
)

func init() {
	RegisterDiscovery(&DiscoveryDefault{})
	RegisterDiscovery(&DiscoveryNats{})
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

func Init(app facade.IApplication, params ...interface{}) {
	if isInit {
		cherryLogger.Error("discovery is init.")
		return
	}
	isInit = true

	clusterConfig := cherryProfile.Config().Get("cluster")
	if clusterConfig.LastError() != nil {
		cherryLogger.Error("`cluster` property not found in profile file.")
		return
	}

	discoveryConfig := clusterConfig.Get("discovery")
	if discoveryConfig.LastError() != nil {
		cherryLogger.Error("`discovery` property not found in profile file.")
		return
	}

	mode := discoveryConfig.Get("mode").ToString()
	if mode == "" {
		cherryLogger.Error("`discovery->mode` property not found in profile file.")
		return
	}

	discovery, found := discoveryMap[mode]
	if discovery == nil || found == false {
		cherryLogger.Errorf("not found mode = %s property in discovery config.", mode)
		return
	}

	discovery.Init(app, discoveryConfig, params...)
	thisDiscovery = discovery
	cherryLogger.Infof("discovery init complete! [mode = %s]", mode)
}

func Discovery() facade.IDiscovery {
	return thisDiscovery
}

func List() []facade.IMember {
	return thisDiscovery.List()
}

func ListByType(nodeType string, filterNodeId ...string) []facade.IMember {
	return thisDiscovery.ListByType(nodeType, filterNodeId...)
}

func GetType(nodeId string) (nodeType string, err error) {
	return thisDiscovery.GetType(nodeId)
}

func GetMember(nodeId string) (member facade.IMember, found bool) {
	return thisDiscovery.GetMember(nodeId)
}

func AddMember(member facade.IMember) {
	thisDiscovery.AddMember(member)
}

func RemoveMember(nodeId string) {
	thisDiscovery.RemoveMember(nodeId)
}

func OnAddMember(listener facade.MemberListener) {
	thisDiscovery.OnAddMember(listener)
}

func OnRemoveMember(listener facade.MemberListener) {
	thisDiscovery.OnRemoveMember(listener)
}

func OnStop() {
	thisDiscovery.OnStop()
}
