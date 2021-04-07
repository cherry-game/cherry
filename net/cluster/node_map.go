package cherryCluster

import (
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	jsoniter "github.com/json-iterator/go"
)

var (
	nodesConfig = NodeMap{}
)

type (
	// {key:nodeType,value:{key:nodeId,value:NodeConfig}}
	NodeMap map[string]map[string]cherryInterfaces.INode
)

// Load nodesConfig config
func Load(config jsoniter.Any) error {
	clusterNodes := config.Get("cluster")

	if clusterNodes.LastError() != nil {
		return cherryUtils.Error("cluster attribute not found in profile file.")
	}

	mode := clusterNodes.Get("mode").ToString()
	if mode == "" {
		return cherryUtils.Error("cluster->mode attribute not found in profile file.")
	}

	if mode == "nodes" {
		loadNodesFromConfigFile(clusterNodes.Get("nodes"))
	} else {
		return cherryUtils.Error("not implemented")
	}

	return nil
}

func Map() *NodeMap {
	return &nodesConfig
}

func GetNode(nodeId string) (node cherryInterfaces.INode, err error) {
	nodeType, err := GetType(nodeId)
	if err != nil {
		return nil, err
	}

	return Get(nodeType, nodeId)
}

func Get(nodeType, nodeId string) (cherryInterfaces.INode, error) {
	node, found := nodesConfig[nodeType][nodeId]
	if found {
		return node, nil
	}
	return nil, cherryUtils.Errorf("node not found. nodeType=%s, nodeId=%s", nodeType, nodeId)
}

func GetType(nodeId string) (nodeType string, error error) {
	if nodeId == "" {
		return "", cherryUtils.Error("nodeId parameter is null.")
	}

	for nType, types := range nodesConfig {
		for nId := range types {
			if nId == nodeId {
				return nType, nil
			}
		}
	}
	return "", cherryUtils.Errorf("nodeId = %s not found. check profile config file please.", nodeId)
}

func loadNodesFromConfigFile(nodesJson jsoniter.Any) {
	for _, nodeType := range nodesJson.Keys() {
		nodesConfig[nodeType] = make(map[string]cherryInterfaces.INode)

		typeJson := nodesJson.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)
			nodeId := item.Get("node_id").ToString()

			nodesConfig[nodeType][nodeId] = &Node{
				nodeId:     nodeId,
				nodeType:   nodeType,
				address:    item.Get("address").ToString(),
				rpcAddress: item.Get("rpc_address").ToString(),
				settings:   item.Get("__settings__"),
				enabled:    item.Get("enabled").ToBool(),
			}
		}
	}
}
