package cherryFacade

import "net"

// IConnector 网络连接器接口
type IConnector interface {
	IComponent
	Start()                     // 启动连接器
	Stop()                      // 停止连接器
	OnConnect(fn OnConnectFunc) // 建立新连接时触发的函数
}

// OnConnectFunc 建立连接时监听的函数
type OnConnectFunc func(conn net.Conn)
