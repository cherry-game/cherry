package cherryRPC

type RpcNodeInfo struct {
	Id       string
	Host     string
	Port     int
	NodeType string
	weight   int
}

type RouteNodes []RpcNodeInfo

type RouteContextClass interface {
	GetNodesByType(nodeType string) RouteNodes
}

type RouteContext struct {
}

//type RouterFunction func(session interfaces.ISession,msg RpcMsg,)

type RpcMsg interface {
	Namespace() string
	NodeType() string
	Service() string
	Method() string
	Args() []interface{}
}

type Callback func(error error, nodeId string)

type TargetRouteFunction func(nodeType string, msg RpcMsg, routeParam interface{}, cb Callback)

type RpcClientOpts struct {
}

type RpcClient struct {
	Nodes   map[string]RpcNodeInfo
	NodeMap map[string][]string

	RoundRobinParam  map[string]int
	WeightRoundParam map[string]struct {
		Index  int
		Weight int
	}
}
