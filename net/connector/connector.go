package cherryConnector

import (
	"crypto/tls"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"net"
	"net/http"
)

type Connector struct {
	Address          string
	Listener         net.Listener
	CertFile         string
	KeyFile          string
	connectListeners []cfacade.OnConnectListener
	connChan         chan cfacade.INetConn
}

func (p *Connector) OnConnectListener(listener ...cfacade.OnConnectListener) {
	p.connectListeners = append(p.connectListeners, listener...)
}

func (p *Connector) ConnectListeners() []cfacade.OnConnectListener {
	return p.connectListeners
}

func (p *Connector) IsSetListener() bool {
	return len(p.connectListeners) > 0
}

func (p *Connector) GetConnChan() chan cfacade.INetConn {
	return p.connChan
}

func (p *Connector) InChan(conn cfacade.INetConn) {
	p.connChan <- conn
}

func (p *Connector) ExecuteListener() {
	go func() {
		for conn := range p.GetConnChan() {
			for _, listener := range p.connectListeners {
				listener(conn)
			}
		}
	}()
}

func NewConnector(address, certFile, keyFile string) Connector {
	return Connector{
		Address:          address,
		Listener:         nil,
		CertFile:         certFile,
		KeyFile:          keyFile,
		connectListeners: make([]cfacade.OnConnectListener, 0),
		connChan:         make(chan cfacade.INetConn),
	}
}

// GetNetListener 证书构造 net.Listener
func GetNetListener(address, certFile, keyFile string) (net.Listener, error) {
	if certFile == "" || keyFile == "" {
		return net.Listen("tcp", address)
	}

	crt, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		clog.Fatalf("failed to listen: %s", err.Error())
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{crt}}

	return tls.Listen("tcp", address, tlsCfg)
}

//CheckOrigin 请求检查函数 防止跨站请求伪造 true则不检查
func CheckOrigin(_ *http.Request) bool {
	return true
}
