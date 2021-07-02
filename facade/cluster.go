package cherryFacade

import jsoniter "github.com/json-iterator/go"

type (
	// {key:nodeType,value:{key:nodeId,value:INode}}
	NodeMap map[string]map[string]INode
)

// INode 结点信息
type INode interface {
	NodeId() string         // 结点id(全局唯一)
	NodeType() string       // 结点类型
	Address() string        // 网络ip地址
	RpcAddress() string     // rpc网络ip地址
	Settings() jsoniter.Any // 结点配置参数
	Enabled() bool          // 是否启用
}

type INodeDiscovery interface {
	All() NodeMap
	GetType(nodeId string) (nodeType string, err error)
	Get(nodeId string) INode
	Sync()
	AddListener(listener INodeListener)
}

type INodeListener interface {
	Add(node INode)
	Remove(node INode)
}
