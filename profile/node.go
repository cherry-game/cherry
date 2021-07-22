package cherryProfile

import (
	"fmt"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
	jsoniter "github.com/json-iterator/go"
)

// Node node info
type Node struct {
	nodeId     string
	nodeType   string
	address    string
	rpcAddress string
	settings   jsoniter.Any
	enabled    bool
}

func (n *Node) NodeId() string {
	return n.nodeId
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

func (n *Node) Settings() jsoniter.Any {
	return n.settings
}

func (n *Node) Enabled() bool {
	return n.enabled
}

const stringFormat = "nodeId = %s, nodeType = %s, address = %s, rpcAddress = %s, enabled = %v"

func (n *Node) String() string {
	return fmt.Sprintf(stringFormat,
		n.nodeId,
		n.nodeType,
		n.address,
		n.rpcAddress,
		n.enabled,
	)
}

func LoadNode(nodeId string) (cherryFacade.INode, error) {
	nodeJson := Config().Get("node")
	if nodeJson.LastError() != nil {
		return nil, cherryError.Error("`nodes` property not found in profile file.")
	}

	for _, nodeType := range nodeJson.Keys() {
		typeJson := nodeJson.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)

			if item.Get("node_id").ToString() != nodeId {
				continue
			}

			node := &Node{
				nodeId:     nodeId,
				nodeType:   nodeType,
				address:    item.Get("address").ToString(),
				rpcAddress: item.Get("rpc_address").ToString(),
				settings:   item.Get("__settings__"),
				enabled:    item.Get("enabled").ToBool(),
			}

			return node, nil
		}
	}

	return nil, cherryError.Errorf("nodeId = %s not found.", nodeId)
}
