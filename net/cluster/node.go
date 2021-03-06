package cherryCluster

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

// Node
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

func (n *Node) String() string {
	return fmt.Sprintf("nodeId = %s, address = %s, rpcAddress = %s enabled = %v",
		n.nodeId,
		n.address,
		n.rpcAddress,
		n.enabled,
	)
}
