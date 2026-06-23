package cherryProfile

import (
	"fmt"
	"regexp"
	"strings"

	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
)

// Node is the concrete implementation of cfacade.INode.
// It holds a single cluster node's identity, network addresses, and
// per-node settings loaded from the profile config.
type Node struct {
	nodeID     string              // globally unique node identifier
	nodeType   string              // node type (e.g. "game", "gate", "map")
	address    string              // public listen address (for frontend nodes)
	rpcAddress string              // RPC listen address (reserved)
	settings   cfacade.ProfileJSON // per-node settings subtree
	enabled    bool                // whether this node is enabled
}

// NodeID returns the globally unique node identifier.
func (n *Node) NodeID() string {
	return n.nodeID
}

// NodeType returns the node type (e.g. "game", "gate", "map").
func (n *Node) NodeType() string {
	return n.nodeType
}

// Address returns the public listen address (for frontend nodes).
func (n *Node) Address() string {
	return n.address
}

// RpcAddress returns the RPC listen address (reserved).
func (n *Node) RpcAddress() string {
	return n.rpcAddress
}

// Settings returns the per-node settings subtree from the profile config.
func (n *Node) Settings() cfacade.ProfileJSON {
	return n.settings
}

// Enabled returns whether this node is enabled.
func (n *Node) Enabled() bool {
	return n.enabled
}

// String returns a human-readable representation of the node.
func (n *Node) String() string {
	return fmt.Sprintf(
		"nodeID=%s nodeType=%s address=%s rpcAddress=%s enabled=%t",
		n.nodeID, n.nodeType, n.address, n.rpcAddress, n.enabled,
	)
}

// GetNodeWithConfig searches the "node" section of the profile config for a
// node matching nodeID and returns it as an INode.
//
// Matching logic (via findNodeID):
//  1. Exact string match against the "node_id" field
//  2. Regex match if the config value looks like a regex (^...$)
//  3. Membership check if the config value is a JSON array of IDs
func GetNodeWithConfig(config *Config, nodeID string) (cfacade.INode, error) {
	nodeConfig := config.GetConfig("node")
	if nodeConfig.LastError() != nil {
		return nil, cerr.Error("\"nodes\" property not found in profile file")
	}

	for _, nodeType := range nodeConfig.Keys() {
		typeJson := nodeConfig.GetConfig(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.GetConfig(i)

			if !findNodeID(nodeID, item.GetConfig("node_id")) {
				continue
			}

			node := &Node{
				nodeID:     nodeID,
				nodeType:   nodeType,
				address:    item.GetString("address"),
				rpcAddress: item.GetString("rpc_address"),
				settings:   item.GetConfig("__settings__"),
				enabled:    item.GetBool("enabled"),
			}

			return node, nil
		}
	}

	return nil, cerr.Errorf("nodeID %s not found", nodeID)
}

// LoadNode looks up a node from the already-loaded global config.
// Must be called after Init(); panics if cfg.jsonConfig is nil.
func LoadNode(nodeID string) (cfacade.INode, error) {
	return GetNodeWithConfig(cfg.jsonConfig, nodeID)
}

// findNodeID checks whether the given nodeID matches the node_id config entry.
//
// It supports three match modes:
//  1. Exact string match
//  2. Regex match (detected by ^ prefix and $ suffix)
//  3. Array membership (when node_id is a JSON array of IDs)
func findNodeID(nodeID string, nodeIDJson cfacade.ProfileJSON) bool {
	configNodeID := nodeIDJson.ToString()
	if configNodeID == nodeID {
		return true
	}

	if isRegexNodeID(nodeID, configNodeID) {
		return true
	}

	for i := 0; i < nodeIDJson.Size(); i++ {
		if nodeIDJson.GetString(i) == nodeID {
			return true
		}
	}

	return false
}

// isRegexNodeID returns true if regexNodeID looks like a regex pattern
// (surrounded by ^ and $) and successfully matches nodeID.
func isRegexNodeID(nodeID, regexNodeID string) bool {
	if !strings.HasPrefix(regexNodeID, "^") || !strings.HasSuffix(regexNodeID, "$") {
		return false
	}

	regex, err := regexp.Compile(regexNodeID)
	if err != nil {
		return false
	}

	return regex.MatchString(nodeID)
}
