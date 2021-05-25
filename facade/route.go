package cherryFacade

// IRoute 路由接口
type IRoute interface {
	NodeType() string    // 结点类型				(0x01 - 首先找到具体类型的服务器)
	HandlerName() string // 消息处理句柄名称		(0x02 - 再找到具体的handler对象)
	Method() string      // 消息处理执行函数名称	(0x03 - 再找到具体的函数执行)
}

// RouteFunction 路由规则处理函数
// @session 当前会话对象
// @packet 处理的网络数据包
// @app 当前应用实例
type RouteFunction func(session ISession, packet interface{}, app IApplication) error
