package cherryConnector

import (
	"github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/packet"
	"io"
	"io/ioutil"
	"net"
)

type (
	TCPConnector struct {
		facade.Component
		connector
		running bool
	}

	TcpConn struct {
		net.Conn
	}
)

func NewTCP(address string) *TCPConnector {
	if address == "" {
		cherryLogger.Warn("create tcp socket fail. address is null.")
		return nil
	}

	return &TCPConnector{
		connector: newConnector(address, "", ""),
	}
}

func NewTCPLTS(address, certFile, keyFile string) *TCPConnector {
	if address == "" {
		cherryLogger.Warn("create tcp socket fail. address is null.")
		return nil
	}

	if certFile == "" || keyFile == "" {
		cherryLogger.Warn("create tcp socket fail. certFile or keyFile is null.")
		return nil
	}

	return &TCPConnector{
		connector: newConnector(address, certFile, keyFile),
	}
}

func (t *TCPConnector) Name() string {
	return "tcp_connector"
}

func (t *TCPConnector) OnAfterInit() {
	t.executeListener()
	go t.OnStart()
}

func (t *TCPConnector) OnStart() {
	if len(t.onConnectListener) < 1 {
		panic("onConnectListener() not set.")
	}

	var err error
	t.listener, err = GetNetListener(t.address, t.certFile, t.keyFile)
	if err != nil {
		cherryLogger.Fatalf("failed to listen: %s", err.Error())
	}

	cherryLogger.Infof("tcp connector listening at address %s", t.address)

	if t.certFile != "" || t.keyFile != "" {
		cherryLogger.Infof("certFile = %s, keyFile = %s", t.certFile, t.keyFile)
	}

	t.running = true

	defer func() {
		if err := t.listener.Close(); err != nil {
			cherryLogger.Errorf("failed to stop: %s", err.Error())
		}
	}()

	for t.running {
		conn, err := t.listener.Accept()
		if err != nil {
			cherryLogger.Errorf("failed to accept TCP connection: %s", err.Error())
			continue
		}

		t.inChan(&TcpConn{Conn: conn})
	}
}

func (t *TCPConnector) OnStop() {
	t.running = false
}

func (t *TcpConn) GetNextMessage() (b []byte, err error) {
	header, err := ioutil.ReadAll(io.LimitReader(t.Conn, int64(cherryPacket.HeadLength)))
	if err != nil {
		return nil, err
	}

	// if the header has no data, we can consider it as a closed connection
	if len(header) == 0 {
		return nil, cherryError.PacketConnectClosed
	}

	msgSize, _, err := cherryPacket.ParseHeader(header)
	if err != nil {
		return nil, err
	}

	msgData, err := ioutil.ReadAll(io.LimitReader(t.Conn, int64(msgSize)))
	if err != nil {
		return nil, err
	}

	if len(msgData) < msgSize {
		return nil, cherryError.PacketMsgSmallerThanExpected
	}

	return append(header, msgData...), nil
}
