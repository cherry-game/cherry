package cherryConnector

import (
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cpacket "github.com/cherry-game/cherry/net/packet"
	"io"
	"io/ioutil"
	"net"
)

type (
	TCPConnector struct {
		cfacade.Component
		Connector
		running bool
	}

	TcpConn struct {
		net.Conn
	}
)

func NewTCP(address string) *TCPConnector {
	if address == "" {
		clog.Warn("create tcp socket fail. Address is null.")
		return nil
	}

	return &TCPConnector{
		Connector: NewConnector(address, "", ""),
	}
}

func NewTCPLTS(address, certFile, keyFile string) *TCPConnector {
	if address == "" {
		clog.Warn("create tcp socket fail. Address is null.")
		return nil
	}

	if certFile == "" || keyFile == "" {
		clog.Warn("create tcp socket fail. CertFile or KeyFile is null.")
		return nil
	}

	return &TCPConnector{
		Connector: NewConnector(address, certFile, keyFile),
	}
}

func (t *TCPConnector) Name() string {
	return "tcp_connector"
}

func (t *TCPConnector) OnAfterInit() {
	go t.OnStart()
}

func (t *TCPConnector) OnStart() {
	if len(t.connectListeners) < 1 {
		panic("ConnectListeners() not set.")
	}

	var err error
	t.Listener, err = GetNetListener(t.Address, t.CertFile, t.KeyFile)
	if err != nil {
		clog.Fatalf("failed to listen: %s", err.Error())
	}

	clog.Infof("tcp Connector listening at Address %s", t.Address)

	if t.CertFile != "" || t.KeyFile != "" {
		clog.Infof("CertFile = %s, KeyFile = %s", t.CertFile, t.KeyFile)
	}

	t.running = true

	t.ExecuteListener()

	defer func() {
		if err := t.Listener.Close(); err != nil {
			clog.Errorf("failed to stop: %s", err.Error())
		}
	}()

	for t.running {
		conn, err := t.Listener.Accept()
		if err != nil {
			clog.Errorf("failed to accept TCP connection: %s", err.Error())
			continue
		}

		t.InChan(&TcpConn{Conn: conn})
	}
}

func (t *TCPConnector) OnStop() {
	t.running = false
}

func (t *TcpConn) GetNextMessage() (b []byte, err error) {
	header, err := ioutil.ReadAll(io.LimitReader(t.Conn, int64(cpacket.HeadLength)))
	if err != nil {
		return nil, err
	}

	// if the header has no data, we can consider it as a closed connection
	if len(header) == 0 {
		return nil, cerr.PacketConnectClosed
	}

	msgSize, _, err := cpacket.ParseHeader(header)
	if err != nil {
		return nil, err
	}

	msgData, err := ioutil.ReadAll(io.LimitReader(t.Conn, int64(msgSize)))
	if err != nil {
		return nil, err
	}

	if len(msgData) < msgSize {
		return nil, cerr.PacketMsgSmallerThanExpected
	}

	return append(header, msgData...), nil
}
