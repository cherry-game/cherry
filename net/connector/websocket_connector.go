package cherryConnector

import (
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"time"
)

type WebSocketConnector struct {
	address           string
	listener          net.Listener
	up                *websocket.Upgrader
	certFile          string
	keyFile           string
	onConnectListener cherryInterfaces.IConnectListener
}

func NewWebSocketConnector(address string) *WebSocketConnector {
	if address == "" {
		cherryLogger.Warn("create websocket fail. address is null.")
		return nil
	}

	ws := &WebSocketConnector{
		address: address,
	}
	return ws
}

func NewWSConnectorLTS(address, certFile, keyFile string) *WebSocketConnector {
	if address == "" {
		cherryLogger.Warn("create websocket fail. address is null.")
		return nil
	}

	if certFile == "" || keyFile == "" {
		cherryLogger.Warn("create websocket fail. certFile or keyFile is null.")
		return nil
	}

	w := &WebSocketConnector{
		address:  address,
		certFile: certFile,
		keyFile:  keyFile,
	}

	return w
}

// ListenAndServe listens and serve in the specified addr
func (w *WebSocketConnector) Start() {
	if w.onConnectListener == nil {
		panic("onConnectionListener() not set.")
	}

	var err error

	w.listener, err = GetNetListener(w.address, w.certFile, w.keyFile)
	if err != nil {
		cherryLogger.Fatalf("Failed to listen: %s", err.Error())
	}

	cherryLogger.Debugf("websocket connector listening at address %s", w.address)

	w.up = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     CheckOrigin,
	}

	defer w.Stop()

	http.Serve(w.listener, w)
}

func (w *WebSocketConnector) OnConnect(listener cherryInterfaces.IConnectListener) {
	w.onConnectListener = listener
}

func (w *WebSocketConnector) Stop() {
	err := w.listener.Close()
	if err != nil {
		cherryLogger.Errorf("Failed to stop: %s", err.Error())
	}
}

//ServerHTTP server.Handler
func (w *WebSocketConnector) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	wsConn, err := w.up.Upgrade(rw, r, nil)
	if err != nil {
		cherryLogger.Infof("Upgrade failure, URI=%s, Error=%s", r.RequestURI, err.Error())
		return
	}

	conn, err := newWSConn(wsConn)
	if err != nil {
		cherryLogger.Errorf("Failed to create new ws connection: %s", err.Error())
		return
	}

	//new goroutine process socket connection
	go w.onConnectListener(conn)
}

// wsConn is an adapter to t.Conn, which implements all t.Conn
// interface base on *websocket.Conn
type wsConn struct {
	conn   *websocket.Conn
	typ    int // message type
	reader io.Reader
}

// newWSConn return an initialized *wsConn
func newWSConn(conn *websocket.Conn) (*wsConn, error) {
	c := &wsConn{conn: conn}

	t, r, err := conn.NextReader()
	if err != nil {
		return nil, err
	}

	c.typ = t
	c.reader = r

	return c, nil
}

func (c *wsConn) Read(b []byte) (int, error) {
	n, err := c.reader.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	} else if err == io.EOF {
		_, r, err := c.conn.NextReader()
		if err != nil {
			return 0, err
		}
		c.reader = r
	}

	return n, nil
}

func (c *wsConn) Write(b []byte) (int, error) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *wsConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}

	return c.SetWriteDeadline(t)
}

func (c *wsConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *wsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
