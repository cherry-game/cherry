package cherryDiscovery

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
)

var (
	discoveryMap  = make(map[string]cfacade.IDiscovery)
	thisDiscovery cfacade.IDiscovery
)

func init() {
	RegisterDiscovery(&DiscoveryDefault{})
	RegisterDiscovery(&DiscoveryNATS{})
	RegisterDiscovery(&DiscoveryETCD{})
}

func RegisterDiscovery(discovery cfacade.IDiscovery) {
	if discovery == nil {
		clog.Fatal("discovery object is nil")
		return
	}

	if discovery.Name() == "" {
		clog.Fatalf("discovery name is empty. %T", discovery)
		return
	}

	discoveryMap[discovery.Name()] = discovery
}

func Init(app cfacade.IApplication) {
	if app.Running() {
		clog.Error("node is running, execute init() fail!")
		return
	}

	discoveryConfig := cprofile.GetConfig("cluster").GetConfig("discovery")
	if discoveryConfig.LastError() != nil {
		clog.Error("`cluster` property not found in profile file.")
		return
	}

	mode := discoveryConfig.GetString("mode")
	if mode == "" {
		clog.Error("`discovery->mode` property not found in profile file.")
		return
	}

	discovery, found := discoveryMap[mode]
	if discovery == nil || found == false {
		clog.Errorf("mode = %s property not found in discovery config.", mode)
		return
	}
	thisDiscovery = discovery
	thisDiscovery.Init(app)

	clog.Infof("discovery init complete! [mode = %s]", mode)
}

func Discovery() cfacade.IDiscovery {
	return thisDiscovery
}

func List() []cfacade.IMember {
	return thisDiscovery.List()
}

func ListByType(nodeType string, filterNodeId ...string) []cfacade.IMember {
	return thisDiscovery.ListByType(nodeType, filterNodeId...)
}

func GetType(nodeId string) (nodeType string, err error) {
	return thisDiscovery.GetType(nodeId)
}

func GetMember(nodeId string) (member cfacade.IMember, found bool) {
	return thisDiscovery.GetMember(nodeId)
}

func AddMember(member cfacade.IMember) {
	thisDiscovery.AddMember(member)
}

func RemoveMember(nodeId string) {
	thisDiscovery.RemoveMember(nodeId)
}

func OnAddMember(listener cfacade.MemberListener) {
	thisDiscovery.OnAddMember(listener)
}

func OnRemoveMember(listener cfacade.MemberListener) {
	thisDiscovery.OnRemoveMember(listener)
}

func OnStop() {
	thisDiscovery.OnStop()
}
