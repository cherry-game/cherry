package cherryCluster

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	cherryInterfaces "github.com/phantacix/cherry/interfaces"
	"github.com/phantacix/cherry/logger"
	"github.com/phantacix/cherry/utils"
)

var (
	nodesConfig = newNodesConfig() //nodesConfig config

)

type (

	// {key:nodeType,value:{key:nodeId,value:NodeConfig}}
	NodeMap map[string]map[string]cherryInterfaces.INode
)

type NodesConfig struct {
	item NodeMap
}

func LoadNodes(config jsoniter.Any) {
	nodesConfig.Load(config)
}

func Nodes() *NodesConfig {
	return nodesConfig
}

func newNodesConfig() *NodesConfig {
	return &NodesConfig{item: make(NodeMap)}
}

func (n *NodesConfig) Map() NodeMap {
	return n.item
}

func (n *NodesConfig) Get(nodeType, nodeId string) cherryInterfaces.INode {
	return n.item[nodeType][nodeId]
}

func (n *NodesConfig) GetType(nodeId string) (nodeType string, error error) {
	if nodeId == "" {
		return "", cherryUtils.Error("nodeId parameter is null.")
	}

	for nType, types := range n.item {
		for nId := range types {
			if nId == nodeId {
				return nType, nil
			}
		}
	}
	return "", cherryUtils.ErrorFormat("nodeId = %s not found. check profile config file please.", nodeId)
}

// Register nodesConfig config
func (n *NodesConfig) Load(config jsoniter.Any) {
	nodesJson := config.Get("nodes")
	for _, nodeType := range nodesJson.Keys() {
		n.item[nodeType] = make(map[string]cherryInterfaces.INode)

		typeJson := nodesJson.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)
			nodeId := item.Get("nodeId").ToString()

			n.item[nodeType][nodeId] = &Node{
				nodeId:     nodeId,
				address:    item.Get("address").ToString(),
				rpcAddress: item.Get("rpcAddress").ToString(),
				settings:   item.Get("__settings__"),
				isDisabled: false,
			}

			cherryLogger.Debugf("load node, %s", n.item[nodeType][nodeId])
		}
	}
}

// Node
type Node struct {
	nodeId     string
	address    string
	rpcAddress string
	settings   jsoniter.Any
	isDisabled bool
}

func (n *Node) NodeId() string {
	return n.nodeId
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

func (n *Node) IsDisabled() bool {
	return n.isDisabled
}

func (n *Node) String() string {
	return fmt.Sprintf("nodeId = %s, address = %s, rpcAddress = %s isDisabled = %v",
		n.nodeId,
		n.address,
		n.rpcAddress,
		n.isDisabled,
	)
}
