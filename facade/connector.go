package cherryFacade

import "net"

// IConnector 网络连接器接口
type IConnector interface {
	OnStart()
	OnStop()
	OnConnect(listener ...OnConnectListener) // 启动前设置连接器监听函数
}

// 建立连接时的监听函数
type OnConnectListener func(conn INetConn)

type INetConn interface {
	net.Conn
	GetNextMessage() (b []byte, err error)
}
