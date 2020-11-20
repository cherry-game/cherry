package cherryInterfaces

import "net"

type IConnectListener func(conn net.Conn)

type IConnector interface {
	Start()

	OnConnect(listener IConnectListener)

	Stop()
}

//type IDataEncoder func(reqId int64, route string, msg []byte) (id int64, body []byte)
//
//type IDataDecoder func(msg []byte) (id int64, route string, body []byte)
