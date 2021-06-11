package cherryCluster

import cherryFacade "github.com/cherry-game/cherry/facade"

type RpcClientComponent struct {
	cherryFacade.Component
	nodeTypeConfig []string // 需要建立连接的结点类型列表
	// 启动后，注册到master node，获取所有结点信息

	// 同类型结点提供路由策略配置函数
}
