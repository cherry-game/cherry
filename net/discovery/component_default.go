package cherryDiscovery

import (
	"math/rand"
	"sync"

	cerr "github.com/cherry-game/cherry/error"
	cslice "github.com/cherry-game/cherry/extend/slice"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
)

// Config keys used to parse node information from the profile file.
const (
	ConfigKeyNode       = "node"
	ConfigKeyNodeID     = "node_id"
	ConfigKeyRPCAddress = "rpc_address"
	ConfigKeySettings   = "__settings__"
)

// ComponentDefault is a file-based discovery implementation for development/testing.
// It reads node information directly from the profile configuration file (profile.json->node).
//
// Fields:
//   - memberMap: stores all known cluster members keyed by nodeID (sync.Map for concurrent access)
//   - onAddListener: callbacks invoked when a new member is added via AddMember()
//   - onUpdateListener: callbacks invoked when an existing member is updated via UpdateMember()
//   - onRemoveListener: callbacks invoked when a member is removed via RemoveMember()
//
// This implementation is designed as a composable base: other backends (nats, etcd)
// embed ComponentDefault to reuse the member storage and listener notification logic.
type ComponentDefault struct {
	cfacade.Component
	memberMap        sync.Map                 // key:nodeID, value:cfacade.IMember
	onAddListener    []cfacade.MemberListener // listeners for member add events
	onUpdateListener []cfacade.MemberListener // listeners for member update events
	onRemoveListener []cfacade.MemberListener // listeners for member remove events
}

func (*ComponentDefault) Name() string {
	return "discovery"
}

// Init loads node configuration from the profile file.
// sync.Map is usable in its zero value, no explicit initialization needed.
func (n *ComponentDefault) Init() {
	n.loadConfig()
}

func (n *ComponentDefault) Mode() string {
	return "default"
}

// loadConfig parses the "node" section of the profile file and populates memberMap.
// Each node entry must have: node_id, rpc_address, and optional __settings__.
// Duplicate nodeIDs or empty nodeIDs within a node type will skip that entry.
func (n *ComponentDefault) loadConfig() {
	nodeConfig := cprofile.GetConfig(ConfigKeyNode)
	if nodeConfig.LastError() != nil {
		clog.Errorf("`%s` property not found in profile file.", ConfigKeyNode)
		return
	}

	for _, nodeType := range nodeConfig.Keys() {
		typeJson := nodeConfig.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)

			nodeID := item.Get(ConfigKeyNodeID).ToString()
			if nodeID == "" {
				clog.Errorf("nodeID is empty in nodeType = %s", nodeType)
				break
			}

			if _, found := n.GetMember(nodeID); found {
				clog.Errorf("nodeType = %s, nodeID = %s, duplicate nodeID", nodeType, nodeID)
				break
			}

			member := &cproto.Member{
				NodeID:   nodeID,
				NodeType: nodeType,
				Address:  item.Get(ConfigKeyRPCAddress).ToString(),
				Settings: make(map[string]string),
			}

			settings := item.Get(ConfigKeySettings)
			for _, key := range settings.Keys() {
				member.Settings[key] = settings.Get(key).ToString()
			}

			n.memberMap.Store(member.NodeID, member)
		}
	}
}

// Map returns a snapshot of all known members as a plain map.
func (n *ComponentDefault) Map() map[string]cfacade.IMember {
	memberMap := map[string]cfacade.IMember{}

	n.memberMap.Range(func(key, value any) bool {
		if member, ok := value.(cfacade.IMember); ok {
			memberMap[member.GetNodeID()] = member
		}
		return true
	})

	return memberMap
}

// ListByType returns members of the given nodeType, excluding any nodes listed in filterNodeID.
func (n *ComponentDefault) ListByType(nodeType string, filterNodeID ...string) []cfacade.IMember {
	var memberList []cfacade.IMember

	n.memberMap.Range(func(key, value any) bool {
		member := value.(cfacade.IMember)
		if member.GetNodeType() == nodeType {
			if _, ok := cslice.StringIn(member.GetNodeID(), filterNodeID); !ok {
				memberList = append(memberList, member)
			}
		}

		return true
	})

	return memberList
}

// Random returns a random member of the given nodeType.
// Returns nil, false if no members of that type exist.
func (n *ComponentDefault) Random(nodeType string) (cfacade.IMember, bool) {
	memberList := n.ListByType(nodeType)
	memberLen := len(memberList)

	if memberLen < 1 {
		return nil, false
	}

	if memberLen == 1 {
		return memberList[0], true
	}

	return memberList[rand.Intn(memberLen)], true
}

// GetType returns the node type for the given nodeID.
func (n *ComponentDefault) GetType(nodeID string) (nodeType string, err error) {
	member, found := n.GetMember(nodeID)
	if !found {
		return "", cerr.Errorf("nodeID = %s not found.", nodeID)
	}
	return member.GetNodeType(), nil
}

// GetMember looks up a member by nodeID. Returns nil, false if not found or nodeID is empty.
func (n *ComponentDefault) GetMember(nodeID string) (cfacade.IMember, bool) {
	if nodeID == "" {
		return nil, false
	}

	value, found := n.memberMap.Load(nodeID)
	if !found {
		return nil, false
	}

	return value.(cfacade.IMember), found
}

// UpdateSetting updates a single setting on the local node's member entry
// and notifies OnUpdateMember listeners. If the local node is not in the
// member map (e.g. before Init completes), the call is silently ignored.
//
// Backends with a transport layer (nats, etcd) override this method to also
// broadcast the change to other nodes.
func (n *ComponentDefault) UpdateSetting(key, value string) {
	if n.App() == nil {
		return
	}

	nodeID := n.App().NodeID()
	if nodeID == "" {
		return
	}

	raw, found := n.memberMap.Load(nodeID)
	if !found {
		return
	}

	member, ok := raw.(*cproto.Member)
	if !ok {
		return
	}

	member.UpdateSetting(key, value)

	for _, listener := range n.onUpdateListener {
		listener(member)
	}
}

// UpdateSettings updates multiple settings on the local node's member entry
// and notifies OnUpdateMember listeners once. If the local node is not in the
// member map, the call is silently ignored.
//
// Backends with a transport layer (nats, etcd) override this method to also
// broadcast the change to other nodes.
func (n *ComponentDefault) UpdateSettings(settings map[string]string) {
	if n.App() == nil {
		return
	}

	nodeID := n.App().NodeID()
	if nodeID == "" {
		return
	}

	raw, found := n.memberMap.Load(nodeID)
	if !found {
		return
	}

	member, ok := raw.(*cproto.Member)
	if !ok {
		return
	}

	member.UpdateSettings(settings)

	for _, listener := range n.onUpdateListener {
		listener(member)
	}
}

// OnAddMember registers a listener that is called after a member is added via AddMember().
// Nil listeners are silently ignored.
func (n *ComponentDefault) OnAddMember(listener cfacade.MemberListener) {
	if listener == nil {
		return
	}
	n.onAddListener = append(n.onAddListener, listener)
}

// OnUpdateMember registers a listener that is called when an existing member is updated via UpdateMember().
// Nil listeners are silently ignored.
func (n *ComponentDefault) OnUpdateMember(listener cfacade.MemberListener) {
	if listener == nil {
		return
	}
	n.onUpdateListener = append(n.onUpdateListener, listener)
}

// OnRemoveMember registers a listener that is called after a member is removed via RemoveMember().
// Nil listeners are silently ignored.
func (n *ComponentDefault) OnRemoveMember(listener cfacade.MemberListener) {
	if listener == nil {
		return
	}
	n.onRemoveListener = append(n.onRemoveListener, listener)
}

// AddMember inserts a member into the local member table.
// If the member already exists, logs a debug message but still notifies add listeners.
// Otherwise stores the new member and notifies all onAddListener callbacks.
func (n *ComponentDefault) AddMember(member cfacade.IMember) {
	_, isDuplicate := n.memberMap.LoadOrStore(member.GetNodeID(), member)
	if isDuplicate {
		clog.Debugf("Add Duplicate Member. [member = %s]", member)
	} else {
		clog.Debugf("Add Member. [ member = %s]", member)
	}

	for _, listener := range n.onAddListener {
		listener(member)
	}
}

// UpdateMember updates an existing member in the local member table.
// If the member doesn't exist, it is stored but no update listeners are notified.
// If the member exists, the stored value is replaced and onUpdateListener callbacks fire.
func (n *ComponentDefault) UpdateMember(member *cproto.Member) {
	value, loaded := n.memberMap.LoadOrStore(member.NodeID, member)
	if loaded {
		member := value.(cfacade.IMember)
		clog.Debugf("Update member. [member = %s]", member)

		for _, listener := range n.onUpdateListener {
			listener(member)
		}
	}
}

// RemoveMember deletes a member from the local member table by nodeID.
// If the member existed, notifies all onRemoveListener callbacks.
func (n *ComponentDefault) RemoveMember(nodeID string) {
	value, loaded := n.memberMap.LoadAndDelete(nodeID)
	if loaded {
		member := value.(cfacade.IMember)
		clog.Debugf("Remove member. [member = %s]", member)

		for _, listener := range n.onRemoveListener {
			listener(member)
		}
	}
}
