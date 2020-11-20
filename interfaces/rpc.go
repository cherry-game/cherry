package cherryInterfaces

type RpcFilterFunc func(nodeId string, msg interface{})

type RpcFilter struct {
	Name   string
	Before RpcFilterFunc
	After  RpcFilterFunc
}

type RpcMessage struct {
	namespace   string
	nodeType    string
	serviceName string
	methodName  string
	args        []interface{}
}

type RpcNodeInfo interface {
	Id() string
	Address() string
	NodeType() string
	Weight() int
}

type RouteFunc func(session ISession, msg RpcMessage, nodeInfo []RpcNodeInfo)
