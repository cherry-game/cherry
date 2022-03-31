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
	cherryProto "github.com/cherry-game/cherry/net/proto"
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
		TagName      string                // 客户标识
		conn         cherryFacade.INetConn // 连接对象
		connected    bool                  // 是否连接
		responseMaps sync.Map              // 响应消息队列 key:ID, value:Message
		pushMsgMaps  sync.Map              // push消息回调列表 key:route, value: OnMessageFn
		nextID       uint32                // 消息自增id
		closeChan    chan struct{}         // 关闭chan
		actionChan   chan ActionFn         // 动作执行队列
	}

	ActionFn    func() error
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
			isErrorBreak:   true,
		},
		responseMaps: sync.Map{},
		pushMsgMaps:  sync.Map{},
		nextID:       0,
		closeChan:    make(chan struct{}),
		actionChan:   make(chan ActionFn, 128),
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

func (p *Client) AddAction(actionFn ActionFn) {
	p.actionChan <- actionFn
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
				ticker.Stop()

				rsp = m.(*cherryMessage.Message)
				ch <- true
				break
			}
		}
	}()

	select {
	case <-ch:
		{
			if rsp.Error {
				errRsp := &cherryProto.CodeResult{}
				if e := p.serializer.Unmarshal(rsp.Data, errRsp); e != nil {
					return nil, e
				}

				return nil, cherryError.Errorf("[route = %s, val = %+v] statusCode = %s", route, val, errRsp.String())
			} else {
				//ok
				return rsp, nil
			}
		}
	case <-ticker.C:
		{
			ticker.Stop()
			return nil, cherryError.Errorf("[route = %s, val = %+v] time out", route, val)
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

func (p *Client) handleHandshake() error {
	// send handshake message
	if err := p.SendRaw(cherryPacket.Handshake, []byte(p.handshake)); err != nil {
		return err
	}

	packets, err := p.getPackets()
	if err != nil {
		return err
	}

	handshakePacket := packets[0]
	if handshakePacket.Type() != cherryPacket.Handshake {
		return cherryError.Errorf("[%s] got handshake packet error.", p.TagName)
	}

	handshakeData := &HandshakeData{}
	if cherryCompress.IsCompressed(handshakePacket.Data()) {
		data, err := cherryCompress.InflateData(handshakePacket.Data())
		if err != nil {
			return err
		}
		handshakePacket.SetData(data)
	}

	err = jsoniter.Unmarshal(handshakePacket.Data(), handshakeData)
	if err != nil {
		return err
	}

	cherryLogger.Debugf("[%s] [Handshake] response data: %+v", p.TagName, handshakeData)

	if handshakeData.Sys.Dict != nil {
		cherryMessage.SetDictionary(handshakeData.Sys.Dict)
	}

	if handshakeData.Sys.Heartbeat > 1 {
		p.heartBeat = handshakeData.Sys.Heartbeat
	}

	err = p.SendRaw(cherryPacket.HandshakeAck, []byte{})
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
			cherryLogger.Warn(err)
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

					p.processMessage(m)
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
		defer p.Disconnect()
	}()

	for {
		select {
		case actionFn := <-p.actionChan:
			{
				if err := actionFn(); err != nil {
					cherryLogger.Warn(err)
					if p.isErrorBreak {
						return
					}
				}
			}
		case <-heartBeatTicker.C:
			{
				if err := p.SendRaw(cherryPacket.Heartbeat, []byte{}); err != nil {
					cherryLogger.Warnf("[%s] packet encode error. %s", p.TagName, err.Error())
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

func (p *Client) getPackets() ([]cherryFacade.IPacket, error) {
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

// Send send the message to the server
func (p *Client) Send(msgType cherryMessage.Type, route string, data []byte) (uint, error) {
	m := &cherryMessage.Message{
		ID:    uint(atomic.AddUint32(&p.nextID, 1)),
		Type:  msgType,
		Route: route,
		Data:  data,
	}

	encMsg, err := cherryMessage.Encode(m)
	if err != nil {
		return 0, err
	}

	bytes, err := p.codec.PacketEncode(cherryPacket.Data, encMsg)
	if err != nil {
		return 0, err
	}

	_, err = p.conn.Write(bytes)
	return m.ID, err
}

func (p *Client) SendRaw(typ cherryPacket.Type, data []byte) error {
	pkg, err := p.codec.PacketEncode(typ, data)
	if err != nil {
		return err
	}
	_, err = p.conn.Write(pkg)
	return err
}
