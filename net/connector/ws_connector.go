package cherryConnector

import (
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cpacket "github.com/cherry-game/cherry/net/packet"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"time"
)

type WSConnector struct {
	cfacade.Component
	connector
	up *websocket.Upgrader
}

func NewWS(address string) *WSConnector {
	if address == "" {
		clog.Warn("create websocket fail. address is null.")
		return nil
	}

	ws := &WSConnector{
		connector: newConnector(address, "", ""),
	}
	return ws
}

func NewWSLTS(address, certFile, keyFile string) *WSConnector {
	if address == "" {
		clog.Warn("create websocket fail. address is null.")
		return nil
	}

	if certFile == "" || keyFile == "" {
		clog.Warn("create websocket fail. certFile or keyFile is null.")
		return nil
	}

	w := &WSConnector{
		connector: newConnector(address, certFile, keyFile),
	}

	return w
}

func (w *WSConnector) Name() string {
	return "websocket_connector"
}

func (w *WSConnector) OnAfterInit() {
	w.executeListener()
	go w.OnStart()
}

func (w *WSConnector) OnStart() {
	if len(w.onConnectListener) < 1 {
		panic("onConnectListener() not set.")
	}

	var err error
	w.listener, err = GetNetListener(w.address, w.certFile, w.keyFile)
	if err != nil {
		clog.Fatalf("failed to listen: %s", err.Error())
	}

	if w.certFile == "" || w.keyFile == "" {
		clog.Infof("websocket connector listening at address ws://%s", w.address)
	} else {
		clog.Infof("websocket connector listening at address wss://%s", w.address)
		clog.Infof("certFile = %s, keyFile = %s", w.certFile, w.keyFile)
	}

	if w.up == nil {
		w.up = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     CheckOrigin,
		}
	}

	http.Serve(w.listener, w)
}

func (w *WSConnector) SetUpgrade(upgrade *websocket.Upgrader) {
	w.up = upgrade
}

func (w *WSConnector) OnStop() {
	err := w.listener.Close()
	if err != nil {
		clog.Errorf("failed to stop: %s", err.Error())
	}
}

//ServerHTTP server.Handler
func (w *WSConnector) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	wsConn, err := w.up.Upgrade(rw, r, nil)
	if err != nil {
		clog.Infof("Upgrade failure, URI=%s, Error=%s", r.RequestURI, err.Error())
		return
	}

	conn, err := NewWSConn(wsConn)
	if err != nil {
		clog.Errorf("Failed to create new ws connection: %s", err.Error())
		return
	}

	w.inChan(conn)
}

// WSConn is an adapter to t.INetConn, which implements all t.INetConn
// interface base on *websocket.INetConn
type WSConn struct {
	conn   *websocket.Conn
	typ    int // message type
	reader io.Reader
}

// NewWSConn return an initialized *WSConn
func NewWSConn(conn *websocket.Conn) (*WSConn, error) {
	c := &WSConn{conn: conn}
	return c, nil
}

// GetNextMessage reads the next message available in the stream
func (c *WSConn) GetNextMessage() (b []byte, err error) {
	_, msgBytes, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	if len(msgBytes) < cpacket.HeadLength {
		return nil, cerr.PacketInvalidHeader
	}

	header := msgBytes[:cpacket.HeadLength]

	msgSize, _, err := cpacket.ParseHeader(header)
	if err != nil {
		return nil, err
	}

	dataLen := len(msgBytes[cpacket.HeadLength:])

	if dataLen < msgSize || dataLen > msgSize {
		return nil, cerr.PacketMsgSmallerThanExpected
	}

	return msgBytes, err
}

func (c *WSConn) Read(b []byte) (int, error) {
	if c.reader == nil {
		t, r, err := c.conn.NextReader()
		if err != nil {
			return 0, err
		}
		c.typ = t
		c.reader = r
	}
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

func (c *WSConn) Write(b []byte) (int, error) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (c *WSConn) Close() error {
	return c.conn.Close()
}

func (c *WSConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *WSConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *WSConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}

	return c.SetWriteDeadline(t)
}

func (c *WSConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *WSConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
