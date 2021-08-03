package cherryClient

import (
	"crypto/tls"
	"net"

	"github.com/gorilla/websocket"
)

type INetConn interface {
	Write(data []byte) (int, error)
	ConnectTO() error
	ConnectToTLS(skipVerify bool) error
	Read() ([]byte, error)
	Close() error
}

type webSocketConn struct {
	INetConn
	conn *websocket.Conn
	addr string
}

func NewWebSocketConn(addr string) *webSocketConn {
	return &webSocketConn{
		addr: addr,
	}
}

func (w *webSocketConn) ConnectTO() error {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(w.addr, nil)
	w.conn = conn
	return err
}

func (w *webSocketConn) ConnectToTLS(skipVerify bool) error {
	return w.ConnectTO()
}

func (w *webSocketConn) Write(data []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.BinaryMessage, data)

	if nil != err {
		return 0, err
	}

	return len(data), err
}

func (w *webSocketConn) Read() ([]byte, error) {
	_, data, err := w.conn.ReadMessage()
	return data, err
}

func (w *webSocketConn) Close() error {
	return w.conn.Close()
}

type tcpSocketConn struct {
	INetConn
	conn net.Conn
	addr string
}

func NewTcpConn(addr string) *tcpSocketConn {
	return &tcpSocketConn{
		addr: addr,
	}
}

func (t *tcpSocketConn) ConnectTO() error {
	conn, err := net.Dial("tcp", t.addr)
	t.conn = conn
	return err

}
func (t *tcpSocketConn) ConnectToTLS(skipVerify bool) error {
	conn, err := tls.Dial("tcp", t.addr, &tls.Config{
		InsecureSkipVerify: skipVerify,
	})

	t.conn = conn
	return err
}

func (t *tcpSocketConn) Write(data []byte) (int, error) {
	return t.conn.Write(data)
}

func (t *tcpSocketConn) Read() ([]byte, error) {
	data := make([]byte, 1024)
	n, err := t.conn.Read(data)
	if err != nil {
		return nil, err
	}

	readData := make([]byte, n)
	copy(readData, data)
	return readData, err
}

func (t *tcpSocketConn) Close() error {
	return t.conn.Close()
}
