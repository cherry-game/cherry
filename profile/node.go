package cherryProfile

import (
	"fmt"

	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
)

// Node node info
type Node struct {
	nodeID     string
	nodeType   string
	address    string
	rpcAddress string
	settings   cfacade.ProfileJSON
	enabled    bool
}

func (n *Node) NodeID() string {
	return n.nodeID
}

func (n *Node) NodeType() string {
	return n.nodeType
}

func (n *Node) Address() string {
	return n.address
}

func (n *Node) RpcAddress() string {
	return n.rpcAddress
}

func (n *Node) Settings() cfacade.ProfileJSON {
	return n.settings
}

func (n *Node) Enabled() bool {
	return n.enabled
}

const stringFormat = "nodeID = %s, nodeType = %s, address = %s, rpcAddress = %s, enabled = %v"

func (n *Node) String() string {
	return fmt.Sprintf(stringFormat,
		n.nodeID,
		n.nodeType,
		n.address,
		n.rpcAddress,
		n.enabled,
	)
}

func GetNodeWithConfig(config *Config, nodeID string) (cfacade.INode, error) {
	nodeConfig := config.GetConfig("node")
	if nodeConfig.LastError() != nil {
		return nil, cerr.Error("`nodes` property not found in profile file.")
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

	return nil, cerr.Errorf("nodeID = %s not found.", nodeID)
}

func LoadNode(nodeID string) (cfacade.INode, error) {
	return GetNodeWithConfig(cfg.jsonConfig, nodeID)
}

func findNodeID(nodeID string, nodeIDJson cfacade.ProfileJSON) bool {
	if nodeIDJson.ToString() == nodeID {
		return true
	}

	for i := 0; i < nodeIDJson.Size(); i++ {
		if nodeIDJson.GetString(i) == nodeID {
			return true
		}
	}

	return false
}
