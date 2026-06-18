// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines the cluster-related abstractions:
//   - IDiscovery: node discovery service (member lookup and change notification)
//   - ICluster: cross-node messaging (publish and request/response)
package cherryFacade

import (
	"time"
)

type (
	// IDiscoveryComponent combines the component lifecycle with discovery capabilities.
	// Register implementations via cherryDiscovery.Register() to make them available
	// to the application builder.
	IDiscoveryComponent interface {
		IComponent
		IDiscovery
	}

	// IDiscovery is the discovery service interface used by business-level code.
	// It provides read access to cluster member information, the ability to update
	// the current node's settings and sync them to other nodes, and notification
	// callbacks for membership changes.
	//
	// Custom backends can implement this interface directly, or embed ComponentDefault
	// from net/discovery to reuse the member storage and listener notification logic.
	IDiscovery interface {
		Mode() string                                                 // discovery mode name (e.g. "default", "nats", "etcd")
		Map() map[string]IMember                                      // snapshot of all known members keyed by nodeID
		ListByType(nodeType string, filterNodeID ...string) []IMember // members of a given node type, excluding filterNodeID
		Random(nodeType string) (IMember, bool)                       // random member of the given node type; nil, false if none exist
		GetMember(nodeID string) (member IMember, found bool)         // lookup a member by nodeID; nil, false if not found
		UpdateSetting(key, value string)                              // update a single setting on THIS node and sync to other nodes
		UpdateSettings(settings map[string]string)                    // update multiple settings on THIS node and sync to other nodes
		OnAddMember(listener MemberListener)                          // register callback invoked after a member is added
		OnUpdateMember(listener MemberListener)                       // register callback invoked after a member is updated
		OnRemoveMember(listener MemberListener)                       // register callback invoked after a member is removed
	}

	// IMember represents a cluster member visible to business code.
	// Implementations may carry additional internal fields (e.g. heartbeat timestamp,
	// timeout configuration) that are not exposed through this interface.
	IMember interface {
		GetNodeID() string              // unique node identifier
		GetNodeType() string            // node type (e.g. "game", "gate", "map")
		GetAddress() string             // RPC address for cross-node communication
		GetSettings() map[string]string // arbitrary key-value metadata for this node
	}

	// MemberListener is called after a member is added, updated, or removed
	// from the local table. The listener receives the affected member as its argument.
	MemberListener func(member IMember)
)

type (
	// IClusterComponent combines the component lifecycle with cluster messaging capabilities.
	IClusterComponent interface {
		IComponent
		ICluster
	}

	// ICluster provides cross-node messaging for the Cherry Actor system.
	//
	// "Local" operations target actors within the same application process,
	// avoiding serialization overhead. "Remote" operations serialize the message
	// and route it through the cluster transport (NATS or similar).
	//
	// Publish methods are fire-and-forget. RequestRemote sends a message and
	// waits for a response, returning the raw payload and an error code.
	ICluster interface {
		Mode() string // cluster transport mode (e.g. "nats")

		// PublishLocal delivers a message to an actor on the local node.
		// The message is NOT serialized — it is passed in-process.
		PublishLocal(nodeID string, msg *Message) error

		// PublishRemote delivers a message to an actor on a remote node.
		// The message is serialized and sent over the cluster transport.
		PublishRemote(nodeID string, msg *Message) error

		// PublishRemoteType delivers a message to a random actor of the given node type.
		// Useful for load-balanced fan-out across a pool of workers.
		PublishRemoteType(nodeType string, msg *Message) error

		// RequestRemote sends a message to a remote actor and blocks until a response
		// is received or the optional timeout expires. Returns the response payload
		// and an error code. If no timeout is specified, a default is used.
		RequestRemote(nodeID string, msg *Message, timeout ...time.Duration) ([]byte, int32)
	}
)
