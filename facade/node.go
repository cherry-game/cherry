package cherryFacade

import jsoniter "github.com/json-iterator/go"

// INode 结点信息，每一个结点代表一个独立的进程实例
type INode interface {
	NodeId() string         // 结点id(全局唯一)
	NodeType() string       // 结点类型
	Address() string        // 网络ip地址
	RpcAddress() string     // rpc网络ip地址
	Settings() jsoniter.Any // 结点配置参数
	Enabled() bool          // 是否启用
}
