package cherryCluster

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
	HandleName() string
	Method() string
	Args() []interface{}
}

type Callback func(error error, nodeId string)

type TargetRouteFunction func(nodeType string, msg RpcMsg, routeParam interface{}, cb Callback)

type RPCClientOpts struct {
}

type RPCClient struct {
	nodeTypeConfig []string // 需要建立连接的结点类型列表
	// 启动后，注册到master node，获取所有结点信息

	// 同类型结点提供路由策略配置函数

	Nodes   map[string]RpcNodeInfo
	NodeMap map[string][]string

	RoundRobinParam  map[string]int
	WeightRoundParam map[string]struct {
		Index  int
		Weight int
	}
}
