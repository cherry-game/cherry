package cherryInterfaces

import jsoniter "github.com/json-iterator/go"

type INode interface {
	NodeId() string
	Address() string
	RpcAddress() string
	Settings() jsoniter.Any
	IsDisabled() bool
}
