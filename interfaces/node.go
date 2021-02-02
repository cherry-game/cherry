package cherryInterfaces

import jsoniter "github.com/json-iterator/go"

type INode interface {
	NodeId() string
	NodeType() string
	Address() string
	RpcAddress() string
	Settings() jsoniter.Any
	Enabled() bool
}
