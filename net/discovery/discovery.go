// Package cherryDiscovery provides cluster node discovery for the Cherry framework.
//
// It supports multiple discovery backends via the IDiscovery interface:
//   - "default" mode: reads node topology from the profile config file (dev/test use)
//   - "nats" mode: master-based discovery over NATS messaging
//   - "etcd" mode: distributed discovery via etcd (maintained in a separate repository)
//
// Custom backends can be registered via Register() and selected via the "cluster.discovery.mode"
// property in the profile configuration file.
package cherryDiscovery

import (
	cerror "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
)

// discoveryMap holds all registered discovery component constructors, keyed by mode name.
var (
	discoveryMap = make(map[string]cfacade.IDiscoveryComponent)
)

func init() {
	Register(&ComponentDefault{})
	Register(&ComponentMaster{})
	// etcd mode is maintained in a separate repository (cherry-game/components/etcd)
	// to avoid pulling etcd client dependencies into the core framework.
}

// Register adds a discovery component implementation to the registry.
// Panics via log.Fatal if component is nil or its Mode() returns empty.
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

// New creates a discovery component based on the profile configuration.
// It reads the mode from "cluster.discovery.mode" and instantiates
// the corresponding registered component. Panics via log.Fatal on
// missing config or unknown mode.
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

// GetDiscovery looks up a registered discovery component by mode name.
func GetDiscovery(mode string) (cfacade.IDiscoveryComponent, error) {
	value, found := discoveryMap[mode]
	if !found {
		return nil, cerror.Errorf("`cluster` mode not found. mode = %s", mode)
	}

	return value, nil
}

// GetMode reads the discovery mode from the profile configuration.
// The config path is "cluster.discovery.mode".
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
