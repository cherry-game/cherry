package cherryDiscovery

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProfile "github.com/cherry-game/cherry/profile"
)

var (
	discoveryMap  = make(map[string]facade.IDiscovery)
	thisDiscovery facade.IDiscovery
)

func init() {
	RegisterDiscovery(&DiscoveryDefault{})
	RegisterDiscovery(&DiscoveryNATS{})
	RegisterDiscovery(&DiscoveryEtcd{})
}

func RegisterDiscovery(discovery facade.IDiscovery) {
	if discovery == nil {
		cherryLogger.Fatal("discovery object is nil")
		return
	}

	if discovery.Name() == "" {
		cherryLogger.Fatalf("discovery name is empty. %T", discovery)
		return
	}

	discoveryMap[discovery.Name()] = discovery
}

func Init(app facade.IApplication) {
	if app.Running() {
		cherryLogger.Error("node is running, execute init() fail!")
		return
	}

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
		cherryLogger.Errorf("mode = %s property not found in discovery config.", mode)
		return
	}
	thisDiscovery = discovery
	thisDiscovery.Init(app)

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
