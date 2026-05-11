package cherryFacade

import (
	"time"
)

type (
	IDiscoveryComponent interface {
		IComponent
		IDiscovery
	}

	// IDiscovery discovery service interface
	IDiscovery interface {
		Mode() string                                                 //
		Map() map[string]IMember                                      // get member list
		ListByType(nodeType string, filterNodeID ...string) []IMember // get member list by node type
		Random(nodeType string) (IMember, bool)                       // random member by node type
		GetType(nodeID string) (nodeType string, err error)           // get node type by node id
		GetMember(nodeID string) (member IMember, found bool)         // get member
		AddMember(member IMember)                                     // add member
		RemoveMember(nodeID string)                                   // remove member
		OnAddMember(listener MemberListener)                          // add member listener
		OnRemoveMember(listener MemberListener)                       // remove member listener
	}

	IMember interface {
		GetNodeID() string
		GetNodeType() string
		GetAddress() string
		GetSettings() map[string]string
	}

	MemberListener func(member IMember) // member add/remove listener
)

type (
	IClusterComponent interface {
		IComponent
		ICluster
	}

	ICluster interface {
		Mode() string                                                                              // cluster implement mode
		PublishLocal(nodeID string, msg *Message) error                                            // publish local message
		PublishRemote(nodeID string, msg *Message) error                                           // publish remote message
		PublishRemoteType(nodeType string, msg *Message) error                                     // publish remote message by node type
		RequestRemote(nodeID string, msg *Message, timeout ...time.Duration) ([]byte, int32)       // request remote message
	}
)
