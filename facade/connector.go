package cherryFacade

import "net"

// IConnector 网络连接器接口
type IConnector interface {
	IComponent
	OnStart()
	OnStop()
	OnConnectListener(listener ...OnConnectListener)
	IsSetListener() bool
	GetConnChan() chan INetConn
}

// OnConnectListener 建立连接时监听的函数
type OnConnectListener func(conn INetConn)

type INetConn interface {
	net.Conn
	GetNextMessage() (b []byte, err error)
}
