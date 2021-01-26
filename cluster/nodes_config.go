package cherryCluster

import (
	"fmt"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	jsoniter "github.com/json-iterator/go"
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
	clusterNodes := config.Get("cluster")

	mode := clusterNodes.Get("mode").ToString()
	if mode == "" {
		cherryLogger.Warn("not found cluster->mode attribute in profile file.")
		return
	}

	if mode == "nodes" {
		n.loadNodesFromConfigFile(clusterNodes.Get("nodes"))
	} else {
		cherryLogger.Error("not implemented")
	}
}

func (n *NodesConfig) loadNodesFromConfigFile(nodesJson jsoniter.Any) {
	for _, nodeType := range nodesJson.Keys() {
		n.item[nodeType] = make(map[string]cherryInterfaces.INode)

		typeJson := nodesJson.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)
			nodeId := item.Get("node_id").ToString()

			n.item[nodeType][nodeId] = &Node{
				nodeId:     nodeId,
				address:    item.Get("address").ToString(),
				rpcAddress: item.Get("rpc_address").ToString(),
				settings:   item.Get("__settings__"),
				enabled:    item.Get("enabled").ToBool(),
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
	enabled    bool
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
