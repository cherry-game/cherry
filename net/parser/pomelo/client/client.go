package pomeloClient

import (
	"crypto/tls"
	"net"
	"net/url"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	cerr "github.com/cherry-game/cherry/error"
	ccompress "github.com/cherry-game/cherry/extend/compress"
	clog "github.com/cherry-game/cherry/logger"
	cconnector "github.com/cherry-game/cherry/net/connector"
	pomeloMessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	pomeloPacket "github.com/cherry-game/cherry/net/parser/pomelo/packet"
	cproto "github.com/cherry-game/cherry/net/proto"
	cserializer "github.com/cherry-game/cherry/net/serializer"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type (
	// Client struct
	Client struct {
		options
		TagName       string         // 客户标识
		conn          net.Conn       // 连接对象
		connected     bool           // 是否连接
		responseMaps  sync.Map       // 响应消息队列 key:ID, value: chan *Message
		pushBindMaps  sync.Map       // push消息绑定列表 key:route, value:OnMessageFn
		nextID        uint32         // 消息自增id
		closeChan     chan struct{}  // 关闭chan
		actionChan    chan ActionFn  // 动作执行队列
		handshakeData *HandshakeData // handshake data
		chWrite       chan []byte
	}

	ActionFn    func() error
	OnMessageFn func(msg *pomeloMessage.Message)
)

// New returns a new client
func New(opts ...Option) *Client {
	client := &Client{
		TagName:   "client",
		connected: false,
		options: options{
			serializer:     cserializer.NewProtobuf(),
			heartBeat:      30,
			requestTimeout: 10 * time.Second,
			isErrorBreak:   true,
		},
		responseMaps:  sync.Map{},
		pushBindMaps:  sync.Map{},
		nextID:        0,
		closeChan:     make(chan struct{}),
		actionChan:    make(chan ActionFn, 128),
		handshakeData: &HandshakeData{},
		chWrite:       make(chan []byte, 64),
	}

	for _, opt := range opts {
		opt(&client.options)
	}

	return client
}

func (p *Client) ConnectToWS(addr string, path string, tlsConfig ...*tls.Config) error {
	u := url.URL{
		Scheme: "ws",
		Host:   addr,
		Path:   path,
	}

	dialer := websocket.DefaultDialer
	if len(tlsConfig) > 0 {
		dialer.TLSClientConfig = tlsConfig[0]
		u.Scheme = "wss"
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	wsConn := cconnector.NewWSConn(conn)
	p.conn = &wsConn

	if err = p.handleHandshake(); err != nil {
		return err
	}

	return nil
}

func (p *Client) ConnectToTCP(addr string, tlsConfig ...*tls.Config) error {
	var conn net.Conn
	var err error

	if len(tlsConfig) > 0 {
		conn, err = tls.Dial("tcp", addr, tlsConfig[0])
	} else {
		conn, err = net.Dial("tcp", addr)
	}

	if err != nil {
		return err
	}

	p.conn = conn

	if err = p.handleHandshake(); err != nil {
		return err
	}

	return nil
}

func (p *Client) Disconnect() {
	for p.connected {
		p.connected = false
		close(p.closeChan)
		err := p.conn.Close()
		if err != nil {
			clog.Error(err)
		}

		clog.Debugf("[%s] is disconnect.", p.TagName)
	}
}

func (p *Client) AddAction(actionFn ActionFn) {
	p.actionChan <- actionFn
}

func (p *Client) Request(route string, val interface{}) (*pomeloMessage.Message, error) {
	id, err := p.Send(pomeloMessage.Request, route, val)
	if err != nil {
		return nil, err
	}

	reqCtx := NewRequestContext(p.requestTimeout)
	p.responseMaps.Store(id, &reqCtx)

	defer func() {
		reqCtx.Close()
	}()

	select {
	case rsp := <-reqCtx.Chan:
		{
			if rsp.Error {
				errRsp := &cproto.Response{}
				if e := p.serializer.Unmarshal(rsp.Data, errRsp); e != nil {
					return nil, e
				}

				return nil, cerr.Errorf("[route = %s, statusCode = %d, req = %+v]", route, errRsp.Code, val)
			} else {
				return rsp, nil
			}
		}
	case <-reqCtx.C:
		{
			p.responseMaps.Delete(id)
			return nil, cerr.Errorf("[route = %s, req = %+v] time out", route, val)
		}
	}
}

// Notify sends a notify to the server
func (p *Client) Notify(route string, val interface{}) error {
	_, err := p.Send(pomeloMessage.Notify, route, val)
	if err != nil {
		return err
	}

	return nil
}

// On listener route
func (p *Client) On(route string, fn OnMessageFn) {
	p.pushBindMaps.Store(route, fn)
}

// IsConnected return the connection status
func (p *Client) IsConnected() bool {
	return p.connected
}

func (p *Client) HandshakeData() *HandshakeData {
	return p.handshakeData
}

func (p *Client) handleHandshake() error {
	// send handshake message
	if err := p.SendRaw(pomeloPacket.Handshake, []byte(p.handshake)); err != nil {
		return err
	}

	packets, err := p.getPackets()
	if err != nil {
		return err
	}

	handshakePacket := packets[0]
	if handshakePacket.Type() != pomeloPacket.Handshake {
		return cerr.Errorf("[%s] got handshake packet error.", p.TagName)
	}

	if ccompress.IsCompressed(handshakePacket.Data()) {
		data, err := ccompress.InflateData(handshakePacket.Data())
		if err != nil {
			return err
		}
		handshakePacket.SetData(data)
	}

	err = jsoniter.Unmarshal(handshakePacket.Data(), p.handshakeData)
	if err != nil {
		return err
	}

	if p.handshakeData.Sys.Dict != nil {
		pomeloMessage.SetDictionary(p.handshakeData.Sys.Dict)
	}

	if p.handshakeData.Sys.Heartbeat > 1 {
		p.heartBeat = p.handshakeData.Sys.Heartbeat / 2
	}

	err = p.SendRaw(pomeloPacket.HandshakeAck, []byte{})
	if err != nil {
		return err
	}

	p.connected = true // is connected

	go p.handlePackets()
	go p.handleData()

	return nil
}

func (p *Client) handlePackets() {
	for p.connected {
		packets, err := p.getPackets()
		if err != nil {
			clog.Warn(err)
			break
		}

		for _, pkg := range packets {
			switch pkg.Type() {
			case pomeloPacket.Data:
				{
					m, err := pomeloMessage.Decode(pkg.Data())
					if err != nil {
						clog.Warnf("[%s] error decoding msg from sv: %s", p.TagName, string(m.Data))
						return
					}

					p.processMessage(&m)
				}
			case pomeloPacket.Kick:
				{
					clog.Warnf("[%s] got kick packet from the server! disconnecting...", p.TagName)
					p.Disconnect()
				}
			}
		}
	}
}

func (p *Client) handleData() {
	heartBeatTicker := time.NewTicker(time.Duration(p.heartBeat) * time.Second)

	defer func() {
		heartBeatTicker.Stop()
		defer p.Disconnect()
	}()

	for {
		select {
		case actionFn := <-p.actionChan:
			{
				if err := actionFn(); err != nil {
					clog.Warn(err)
					if p.isErrorBreak {
						return
					}
				}
			}
		case <-heartBeatTicker.C:
			{
				if err := p.SendRaw(pomeloPacket.Heartbeat, []byte{}); err != nil {
					clog.Warnf("[%s] packet encode error. %s", p.TagName, err.Error())
					return
				}
			}
		case bytes := <-p.chWrite:
			{
				if _, err := p.conn.Write(bytes); err != nil {
					clog.Warnf("[%s] write packet fail. %s", p.TagName, err.Error())
					return
				}
			}
		case <-p.closeChan:
			return
		}
	}
}

func (p *Client) processMessage(msg *pomeloMessage.Message) {
	defer func() {
		if r := recover(); r != nil {
			clog.Errorf("[%s] recover in executor. %s", p.TagName, string(debug.Stack()))
		}
	}()

	if msg.Type == pomeloMessage.Response {
		value, found := p.responseMaps.LoadAndDelete(msg.ID)
		if !found {
			clog.Warnf("callback not found. [msg = %v]", msg)
			return
		}

		reqCtx, ok := value.(*RequestContext)
		if ok {
			reqCtx.Chan <- msg
		}

		return
	}

	if msg.Type == pomeloMessage.Push {
		value, found := p.pushBindMaps.Load(msg.Route)
		if found {
			fn, ok := value.(OnMessageFn)
			if ok {
				fn(msg)
			}
			return
		}
	}
}

func (p *Client) getPackets() ([]*pomeloPacket.Packet, error) {
	packets, isBreak, err := pomeloPacket.Read(p.conn)
	if err != nil {
		clog.Errorf("[%s] error decoding packet from server: %s", p.TagName, err.Error())
	}

	if isBreak {
		return nil, err
	}

	return packets, nil
}

// Send the message to the server
func (p *Client) Send(msgType pomeloMessage.Type, route string, val interface{}) (uint, error) {
	data, err := p.serializer.Marshal(val)
	if err != nil {
		return 0, cerr.Errorf("serializer error.[route = %s, val =%v]", route, val)
	}

	m := &pomeloMessage.Message{
		ID:    uint(atomic.AddUint32(&p.nextID, 1)),
		Type:  msgType,
		Route: route,
		Data:  data,
	}

	encMsg, err := pomeloMessage.Encode(m)
	if err != nil {
		return 0, err
	}

	bytes, err := pomeloPacket.Encode(pomeloPacket.Data, encMsg)
	if err != nil {
		return 0, err
	}

	p.chWrite <- bytes
	return m.ID, err
}

func (p *Client) SendRaw(typ pomeloPacket.Type, data []byte) error {
	pkg, err := pomeloPacket.Encode(typ, data)
	if err != nil {
		return err
	}
	_, err = p.conn.Write(pkg)
	return err
}
