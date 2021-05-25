package cherryConnector

import (
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"net"
)

type TCPConnector struct {
	address           string
	listener          net.Listener
	running           bool
	certFile          string
	keyFile           string
	onConnectListener cherryFacade.IConnectListener
}

func NewTCPConnector(address string) *TCPConnector {
	if address == "" {
		cherryLogger.Warn("create tcp socket fail. address is null.")
		return nil
	}

	return &TCPConnector{
		address: address,
	}
}

func NewTCPConnectorLTS(address, certFile, keyFile string) *TCPConnector {
	if address == "" {
		cherryLogger.Warn("create tcp socket fail. address is null.")
		return nil
	}

	if certFile == "" || keyFile == "" {
		cherryLogger.Warn("create tcp socket fail. certFile or keyFile is null.")
		return nil
	}

	return &TCPConnector{
		address:  address,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

// OnStartup
func (t *TCPConnector) OnStart() {
	if t.onConnectListener == nil {
		panic("OnConnect() not set.")
	}

	var err error

	t.listener, err = GetNetListener(t.address, t.certFile, t.keyFile)
	if err != nil {
		cherryLogger.Fatalf("Failed to listen: %s", err.Error())
	}

	cherryLogger.Debugf("tcp connector listening at address %s", t.address)

	t.running = true
	for t.running {
		conn, err := t.listener.Accept()
		if err != nil {
			cherryLogger.Errorf("Failed to accept TCP connection: %s", err.Error())
			continue
		}

		// open goroutine for new connection
		go t.onConnectListener(conn)
	}
}

func (t *TCPConnector) OnConnect(listener cherryFacade.IConnectListener) {
	t.onConnectListener = listener
}

// OnShutdown stops the acceptor
func (t *TCPConnector) OnStop() {
	t.running = false
	err := t.listener.Close()
	if err != nil {
		cherryLogger.Errorf("Failed to stop: %s", err.Error())
	}
}
