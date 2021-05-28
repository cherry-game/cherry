package cherryFacade

import "net"

// IConnector 网络连接器接口
type IConnector interface {
	OnConnect(listener IConnectListener) // 启动前设置连接器监听函数
	OnStart()                            // 启动
	OnStop()                             // 停止
}

// IConnectListener 网络连接器监听函数
type IConnectListener func(conn net.Conn)
