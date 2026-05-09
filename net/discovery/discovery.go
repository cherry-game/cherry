package cherryDiscovery

import (
	cerror "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
)

var (
	discoveryMap = make(map[string]cfacade.IDiscoveryComponent)
)

func init() {
	Register(&ComponentDefault{})
	Register(&ComponentMaster{})
	//RegisterDiscovery(&DiscoveryETCD{})
}

func Register(component cfacade.IDiscoveryComponent) {
	if component == nil {
		clog.Fatal("Discovery component is nil")
		return
	}

	if component.Mode() == "" {
		clog.Fatalf("Discovery mode is empty. %T", component)
		return
	}

	discoveryMap[component.Mode()] = component
}

func New() cfacade.IDiscoveryComponent {
	mode, err := GetMode()
	if err != nil {
		clog.Fatal(err)
		return nil
	}

	component, err := GetDiscovery(mode)
	if err != nil {
		clog.Fatal(err)
		return nil
	}

	return component
}

func GetDiscovery(mode string) (cfacade.IDiscoveryComponent, error) {
	value, found := discoveryMap[mode]
	if !found {
		return nil, cerror.Errorf("`cluster` mode not found. mode = %s", mode)
	}

	return value, nil
}

func GetMode() (string, error) {
	config := cprofile.GetConfig("cluster").GetConfig("discovery")
	if config.LastError() != nil {
		return "", cerror.Error("`cluster` property not found in profile file.")
	}

	mode := config.GetString("mode")
	if mode == "" {
		return "", cerror.Error("`discovery->mode` property not found in profile file.")
	}

	return mode, nil
}
