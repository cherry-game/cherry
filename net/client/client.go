package cherryClient

import (
	"crypto/tls"
	cherryError "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/extend/compress"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/packet"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"net"
	"net/url"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// Client struct
	Client struct {
		options
		TagName         string                      // 客户标识
		conn            cherryFacade.INetConn       // 连接对象
		connected       bool                        // 是否连接
		incomingMsgChan chan *cherryMessage.Message // 接收消息队列
		responseMaps    sync.Map                    // 响应消息队列 key:ID, value:Message
		pushMsgMaps     sync.Map                    // push消息回调列表 key:route, value: OnMessageFn
		nextID          uint32                      // 消息自增id
		closeChan       chan struct{}               // 关闭chan
	}

	OnMessageFn func(msg *cherryMessage.Message)
)

// New returns a new client
func New(opts ...Option) *Client {
	client := &Client{
		TagName:   "client",
		connected: false,
		options: options{
			serializer:     cherrySerializer.NewProtobuf(),
			codec:          cherryPacket.NewPomeloCodec(),
			heartBeat:      30,
			requestTimeout: 3 * time.Second,
		},
		incomingMsgChan: make(chan *cherryMessage.Message, 256),
		responseMaps:    sync.Map{},
		pushMsgMaps:     sync.Map{},
		nextID:          0,
		closeChan:       make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&client.options)
	}

	return client
}

func (p *Client) Request(route string, val interface{}) (*cherryMessage.Message, error) {
	data, err := p.serializer.Marshal(val)
	if err != nil {
		return nil, cherryError.Error("serializer error.")
	}

	id, err := p.Send(cherryMessage.Request, route, data)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(p.requestTimeout)
	ch := make(chan bool)

	rsp := &cherryMessage.Message{}

	go func() {
		for {
			if m, found := p.responseMaps.LoadAndDelete(id); found {
				rsp = m.(*cherryMessage.Message)
				ch <- true
				break
			}
		}
	}()

	select {
	case <-ch:
		{
			//ok
			return rsp, nil
		}
	case <-ticker.C:
		{
			return nil, cherryError.Error("time out")
		}
	}
}

// Notify sends a notify to the server
func (p *Client) Notify(route string, val interface{}) error {
	data, err := p.serializer.Marshal(val)
	if err != nil {
		return err
	}

	_, err = p.Send(cherryMessage.Notify, route, data)
	if err != nil {
		return err
	}

	return nil
}

// On listener route
func (p *Client) On(route string, fn OnMessageFn) {
	p.pushMsgMaps.Store(route, fn)
}

// IsConnected return the connection status
func (p *Client) IsConnected() bool {
	return p.connected
}

// Disconnect disconnects the client
func (p *Client) Disconnect() {
	for p.connected {
		p.connected = false
		close(p.closeChan)
		err := p.conn.Close()
		if err != nil {
			cherryLogger.Error(err)
		}

		cherryLogger.Debugf("[%s] is disconnect.", p.TagName)
	}
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

	p.conn, err = cherryConnector.NewWSConn(conn)
	if err != nil {
		return err
	}

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

	p.conn = &cherryConnector.TcpConn{
		Conn: conn,
	}

	if err = p.handleHandshake(); err != nil {
		return err
	}

	return nil
}

func (p *Client) handleHandshake() error {
	if err := p.sendHandshakeRequest(); err != nil {
		return err
	}

	if err := p.handleHandshakeResponse(); err != nil {
		return err
	}

	go p.handlePackets()
	go p.handleData()

	return nil
}

func (p *Client) sendHandshakeRequest() error {
	pkg, err := p.codec.PacketEncode(cherryPacket.Handshake, []byte(p.handshake))
	if err != nil {
		return err
	}
	_, err = p.conn.Write(pkg)
	return err
}

func (p *Client) handleHandshakeResponse() error {
	packets, err := p.readPackets()
	if err != nil {
		return err
	}

	handshakePacket := packets[0]
	if handshakePacket.Type() != cherryPacket.Handshake {
		return cherryError.Errorf("[%s] got handshake packet error.", p.TagName)
	}

	handshake := &HandshakeData{}
	if cherryCompress.IsCompressed(handshakePacket.Data()) {
		data, err := cherryCompress.InflateData(handshakePacket.Data())
		if err != nil {
			return err
		}
		handshakePacket.SetData(data)
	}

	err = jsoniter.Unmarshal(handshakePacket.Data(), handshake)
	if err != nil {
		return err
	}

	cherryLogger.Debugf("[%s] got handshake from sv, data: %+v", p.TagName, handshake)

	if handshake.Sys.Dict != nil {
		cherryMessage.SetDictionary(handshake.Sys.Dict)
	}

	pkg, err := p.codec.PacketEncode(cherryPacket.HandshakeAck, []byte{})
	if err != nil {
		return err
	}

	_, err = p.conn.Write(pkg)
	if err != nil {
		return err
	}

	p.connected = true // is connected

	if handshake.Sys.Heartbeat > 1 {
		p.heartBeat = handshake.Sys.Heartbeat
	}

	return nil
}

func (p *Client) handlePackets() {
	defer p.Disconnect()

	for p.connected {
		packets, err := p.readPackets()
		if err != nil {
			cherryLogger.Error(err)
			break
		}

		for _, pkg := range packets {
			switch pkg.Type() {
			case cherryPacket.Data:
				{
					m, err := cherryMessage.Decode(pkg.Data())
					if err != nil {
						cherryLogger.Warnf("[%s] error decoding msg from sv: %s", p.TagName, string(m.Data))
						return
					}

					p.incomingMsgChan <- m
				}
			case cherryPacket.Kick:
				{
					cherryLogger.Warnf("[%s] got kick packet from the server! disconnecting...", p.TagName)
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
	}()

	for {
		select {
		case msg := <-p.incomingMsgChan:
			{
				p.processMessage(msg)
			}
		case <-heartBeatTicker.C:
			{
				pkg, _ := p.codec.PacketEncode(cherryPacket.Heartbeat, []byte{})
				_, err := p.conn.Write(pkg)
				if err != nil {
					cherryLogger.Warnf("[%s] packet encode error. %s", p.TagName, err.Error())
					return
				} else {
					cherryLogger.Debugf("[%s] Send heart beat", p.TagName)
				}
			}
		case <-p.closeChan:
			return
		}
	}
}

func (p *Client) processMessage(msg *cherryMessage.Message) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Warnf("[%s] recover in executor. %s", p.TagName, string(debug.Stack()))
		}
	}()

	if msg.Type == cherryMessage.Response {
		p.responseMaps.Store(msg.ID, msg)
		return
	}

	if msg.Type == cherryMessage.Push {
		value, found := p.pushMsgMaps.LoadAndDelete(msg.Route)
		if found {
			fn, ok := value.(OnMessageFn)
			if ok {
				fn(msg)
			}
			return
		}
	}
}

func (p *Client) readPackets() ([]cherryFacade.IPacket, error) {
	data, err := p.conn.GetNextMessage()
	if err != nil {
		return nil, err
	}

	packets, err := p.codec.PacketDecode(data)
	if err != nil {
		cherryLogger.Errorf("[%s] error decoding packet from server: %s", p.TagName, err.Error())
	}

	return packets, nil
}

func (p *Client) buildPacket(msg *cherryMessage.Message) ([]byte, error) {
	encMsg, err := cherryMessage.Encode(msg)
	if err != nil {
		return nil, err
	}

	bytes, err := p.codec.PacketEncode(cherryPacket.Data, encMsg)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Send send the message to the server
func (p *Client) Send(msgType cherryMessage.Type, route string, data []byte) (uint, error) {
	m := &cherryMessage.Message{
		ID:    uint(atomic.AddUint32(&p.nextID, 1)),
		Type:  msgType,
		Route: route,
		Data:  data,
	}

	pkg, err := p.buildPacket(m)
	if err != nil {
		return m.ID, err
	}

	_, err = p.conn.Write(pkg)
	return m.ID, err
}
