package cherryCluster

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/facade"
	jsoniter "github.com/json-iterator/go"
)

var (
	nodesConfig = cherryFacade.NodeMap{}
)

// Load nodesConfig config
func Load(config jsoniter.Any) error {
	clusterConfig := config.Get("cluster")

	if clusterConfig.LastError() != nil {
		return cherryError.Error("cluster attribute not found in profile file.")
	}

	mode := clusterConfig.Get("mode").ToString()
	if mode == "" {
		return cherryError.Error("cluster->mode attribute not found in profile file.")
	}

	if mode == "nodes" {
		loadNodesFromConfigFile(clusterConfig.Get("nodes"))
	} else {
		panic(fmt.Sprintf("mode = %s not implemented", mode))
	}

	return nil
}

func Map() *cherryFacade.NodeMap {
	return &nodesConfig
}

func GetNode(nodeId string) (node cherryFacade.INode, err error) {
	nodeType, err := GetType(nodeId)
	if err != nil {
		return nil, err
	}

	return Get(nodeType, nodeId)
}

func Get(nodeType, nodeId string) (cherryFacade.INode, error) {
	node, found := nodesConfig[nodeType][nodeId]
	if found {
		return node, nil
	}
	return nil, cherryError.Errorf("node not found. nodeType=%s, nodeId=%s", nodeType, nodeId)
}

func GetType(nodeId string) (nodeType string, error error) {
	if nodeId == "" {
		return "", cherryError.Error("nodeId parameter is null.")
	}

	for nType, types := range nodesConfig {
		for nId := range types {
			if nId == nodeId {
				return nType, nil
			}
		}
	}
	return "", cherryError.Errorf("nodeId = %s not found. check profile config file please.", nodeId)
}

func loadNodesFromConfigFile(nodesJson jsoniter.Any) {
	for _, nodeType := range nodesJson.Keys() {
		nodesConfig[nodeType] = make(map[string]cherryFacade.INode)

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
