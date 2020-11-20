package cherryRPC

import (
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/utils"
	"hash/crc32"
	"math"
	"math/rand"
)

// Calculate route info and return an appropriate node id.
func DefaultRoute(session cherryInterfaces.ISession, msg RpcMsg, context RouteContextClass, cb Callback) {
	list := context.GetNodesByType(msg.NodeType())
	if list == nil || len(list) < 1 {
		cb(cherryUtils.ErrorFormat("can not find node info for type:%s", msg.NodeType()), "")
		return
	}

	hash := crc32.ChecksumIEEE([]byte(string(session.UID())))
	index := int(hash) % len(list)
	cb(nil, list[index].Id)
}

// Random algorithm for calculating node id.
func RandomRoute(client RpcClient, nodeType string, msg RpcMsg, cb Callback) {
	list := client.NodeMap[nodeType]
	if list == nil || len(list) < 1 {
		cb(cherryUtils.ErrorFormat("rpc servers not exist with nodeType:%s", nodeType), "")
		return
	}

	index := int(math.Floor(float64(rand.Int() * len(list))))
	cb(nil, list[index])
}

func RoundRobinRoute(client RpcClient, nodeType string, msg RpcMsg, cb Callback) {
	list := client.NodeMap[nodeType]
	if list == nil || len(list) < 1 {
		cb(cherryUtils.ErrorFormat("rpc servers not exist with nodeType:%s", nodeType), "")
		return
	}

	index := client.RoundRobinParam[nodeType]
	if index == 0 {
		index += 1
	}

	if index == math.MaxInt32 {
		index = 0
	}

	client.RoundRobinParam[nodeType] = index
}

func WeightRoundRoute(client RpcClient, nodeType string, msg RpcMsg, cb Callback) {
	list := client.NodeMap[nodeType]
	if list == nil || len(list) < 1 {
		cb(cherryUtils.ErrorFormat("rpc servers not exist with nodeType:%s", nodeType), "")
		return
	}

	var index = -1
	var weight = 0

	if _, found := client.WeightRoundParam[nodeType]; found == false {
		index = client.WeightRoundParam[nodeType].Index
		weight = client.WeightRoundParam[nodeType].Weight
	}

	var getMaxWeight = func() int {
		var maxWeight = -1
		for _, s := range list {
			server := client.Nodes[s]
			if server.weight > maxWeight {
				maxWeight = server.weight
			}
		}
		return maxWeight
	}

	for {
		index = (index + 1) % len(list)
		if index == 0 {
			weight = weight - 1
			if weight <= 0 {
				weight = getMaxWeight()
				if weight <= 0 {
					cb(cherryUtils.Error("rpc weight round route get invalid weight."), "")
					break
				}
			}
		}

		server := client.Nodes[list[index]]
		if server.weight >= weight {
			client.WeightRoundParam[nodeType] = struct {
				Index  int
				Weight int
			}{Index: index, Weight: weight}

			cb(nil, server.Id)
			break
		}
	}
}
