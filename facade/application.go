// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines application-level abstractions:
//   - INode: node identity and configuration
//   - IApplication: the application container — component registry, lifecycle, service accessors
//   - ProfileJSON: typed profile/config reader
package cherryFacade

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type (
	// INode represents a cluster node's identity and configuration.
	INode interface {
		NodeID() string        // globally unique node ID
		NodeType() string      // node type (e.g. "game", "gate", "map")
		Address() string       // public listen address (for frontend nodes)
		RpcAddress() string    // RPC listen address (reserved)
		Settings() ProfileJSON // node settings from profile
		Enabled() bool         // whether this node is enabled
	}

	IApplication interface {
		INode
		Running() bool                     // whether the application is running
		DieChan() chan bool                // closed when application shuts down
		IsFrontend() bool                  // whether this is a frontend node
		Register(components ...IComponent) // register components (before startup)
		Find(name string) IComponent       // find a component by name
		Remove(name string) IComponent     // remove a component by name
		All() []IComponent                 // list all registered components
		OnShutdown(fn ...func())           // register shutdown hooks (called in registration order)
		Startup()                          // start the application (blocks until shutdown)
		Shutdown()                         // shut down the application
		Serializer() ISerializer           // message serializer
		Discovery() IDiscovery             // node discovery service
		Cluster() ICluster                 // cluster messaging service
		ActorSystem() IActorSystem         // Actor system
	}

	// ProfileJSON is a typed profile/config reader backed by jsoniter.
	ProfileJSON interface {
		jsoniter.Any
		GetConfig(path ...interface{}) ProfileJSON
		GetString(path interface{}, defaultVal ...string) string
		GetBool(path interface{}, defaultVal ...bool) bool
		GetInt(path interface{}, defaultVal ...int) int
		GetInt32(path interface{}, defaultVal ...int32) int32
		GetInt64(path interface{}, defaultVal ...int64) int64
		GetDuration(path interface{}, defaultVal ...time.Duration) time.Duration
		Unmarshal(ptrVal interface{}) error
	}
)
