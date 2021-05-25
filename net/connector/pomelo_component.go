package cherryConnector

import (
	"encoding/json"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/compress"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/packet"
	"github.com/cherry-game/cherry/net/packet/pomelo"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/serialize"
	"github.com/cherry-game/cherry/net/session"
	"github.com/cherry-game/cherry/profile"
	"net"
)

type BlackListFn func(list []string)

type CheckClientFn func(typ string, version string) bool

type PomeloComponentOptions struct {
	Connector       cherryFacade.IConnector
	ConnectListener cherryFacade.IConnectListener
	PacketEncode    cherryPacket.Encoder
	PacketDecode    cherryPacket.Decoder
	Serializer      cherryFacade.ISerializer
	SessionOnClosed cherryFacade.SessionListener
	SessionOnError  cherryFacade.SessionListener

	BlackListFunc       BlackListFn
	BlackList           []string
	ForwardMessage      bool
	Heartbeat           int
	DisconnectOnTimeout bool
	UseDict             bool
	UseProtobuf         bool
	UseCrypto           bool
	UseHostFilter       bool
	CheckClient         CheckClientFn
	DataCompression     bool
}

type PomeloComponent struct {
	cherryFacade.Component
	PomeloComponentOptions
	connCount        *ConnectStat
	sessionComponent *cherrySession.SessionComponent
	handlerComponent *cherryHandler.HandlerComponent
}

func NewPomelo() *PomeloComponent {
	s := &PomeloComponent{
		PomeloComponentOptions: PomeloComponentOptions{
			PacketEncode:        cherryPacketPomelo.NewEncoder(),
			PacketDecode:        cherryPacketPomelo.NewDecoder(),
			Serializer:          cherrySerialize.NewJSON(),
			BlackListFunc:       nil,
			BlackList:           nil,
			ForwardMessage:      false,
			Heartbeat:           30,
			DisconnectOnTimeout: false,
			UseDict:             false,
			UseProtobuf:         false,
			UseCrypto:           false,
			UseHostFilter:       false,
			CheckClient:         nil,
		},
		connCount: &ConnectStat{},
	}
	return s
}

func NewPomeloWithOpts(opts PomeloComponentOptions) *PomeloComponent {
	return &PomeloComponent{
		PomeloComponentOptions: opts,
		connCount:              &ConnectStat{},
	}
}

func (p *PomeloComponent) Name() string {
	return cherryConst.ConnectorPomeloComponent
}

func (p *PomeloComponent) Init() {
}

func (p *PomeloComponent) OnAfterInit() {
	p.sessionComponent = p.App().Find(cherryConst.SessionComponent).(*cherrySession.SessionComponent)
	if p.sessionComponent == nil {
		panic("please preload session component.")
	}

	p.handlerComponent = p.App().Find(cherryConst.HandlerComponent).(*cherryHandler.HandlerComponent)
	if p.handlerComponent == nil {
		panic("please preload handler component.")
	}

	p.initHandshakeData()
	p.initHeartbeatData()

	// when new connect bind the session
	if p.ConnectListener != nil {
		p.Connector.OnConnect(p.ConnectListener)
	} else {
		p.Connector.OnConnect(p.initSession)
	}

	// new goroutine
	go p.Connector.OnStart()
}

func (p *PomeloComponent) initSession(conn net.Conn) {
	session := p.sessionComponent.Create(conn, nil) //TODO INetworkEntity
	p.connCount.IncreaseConn()

	//receive msg
	session.OnMessage(func(bytes []byte) error {
		packets, err := p.PacketDecode.Decode(bytes)
		if err != nil {
			cherryLogger.Warnf("bytes parse to packets error. session=[%s]", session)
			session.Closed()
			return nil
		}

		if len(packets) < 1 {
			cherryLogger.Warnf("bytes parse to Packets length < 1. session=[%s]", session)
			return nil
		}

		for _, pkg := range packets {
			if err := p.processPacket(session, pkg); err != nil {
				cherryLogger.Warn(err)
				return nil
			}
		}
		return nil
	})

	if p.SessionOnClosed == nil {
		session.OnClose(p.SessionOnClosed)
	}

	session.OnClose(func(_ cherryFacade.ISession) {
		p.connCount.DecreaseConn()
	})

	if p.SessionOnError == nil {
		session.OnError(p.SessionOnError)
	}

	//create a new goroutine to process read data for current socket
	session.Start()
}

func (p *PomeloComponent) processPacket(session cherryFacade.ISession, pkg *cherryPacket.Packet) error {
	switch pkg.Type {
	case cherryPacketPomelo.Handshake:
		if err := session.Send(hrd); err != nil {
			return err
		}
		session.SetStatus(cherryPacketPomelo.WaitAck)

		cherryLogger.Debugf("[Handshake] session=[%session]", session)

	case cherryPacketPomelo.HandshakeAck:
		if session.Status() != cherryPacketPomelo.WaitAck {
			cherryLogger.Warnf("[HandshakeAck] session=[%session]", session)
			session.Closed()
			return nil
		}

		session.SetStatus(cherryPacketPomelo.Working)
		if cherryProfile.Debug() {
			cherryLogger.Debugf("[HandshakeAck] session=[%session]", session)
		}

	case cherryPacketPomelo.Data:
		if session.Status() != cherryPacketPomelo.Working {
			return cherryUtils.Errorf("[Msg] status error. session=[%session]", session)
		}

		msg, err := cherryMessage.Decode(pkg.Data)
		if err != nil {
			p.handleMessage(session, msg)
		}

	case cherryPacketPomelo.Heartbeat:
		d, err := p.PacketEncode.Encode(cherryPacketPomelo.Heartbeat, nil)
		if err != nil {
			return err
		}

		err = session.Send(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PomeloComponent) handleMessage(session cherryFacade.ISession, msg *cherryMessage.Message) {
	route, err := cherryRoute.Decode(msg.Route)
	if err != nil {
		cherryLogger.Warnf("failed to decode route:%s", err.Error())
		return
	}

	if route.NodeType() == "" {
		//TODO ... remove this
		//r.NodeType = p.IAppContext().NodeType()
		return
	}

	if route.NodeType() == p.App().NodeType() {
		unHandleMessage := &cherryHandler.UnhandledMessage{
			Session: session,
			Route:   route,
			Msg:     msg,
		}

		p.handlerComponent.DoHandle(unHandleMessage)
	} else {
		// TODO forward to target node
	}
}

func (p *PomeloComponent) OnStop() {
	if p.Connector != nil {
		p.Connector.OnStop()
	}
}

var (
	// hbd contains the heartbeat packet data
	hbd []byte
	// hrd contains the handshake response data
	hrd []byte
)

func (p *PomeloComponent) initHandshakeData() {
	hData := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat":   p.Heartbeat,
			"dict":        cherryMessage.GetDictionary(),
			"ISerializer": "protobuf",
		},
	}
	data, err := json.Marshal(hData)
	if err != nil {
		panic(err)
	}

	if p.DataCompression {
		compressedData, err := cherryCompress.DeflateData(data)
		if err != nil {
			panic(err)
		}

		if len(compressedData) < len(data) {
			data = compressedData
		}
	}

	hrd, err = p.PacketEncode.Encode(cherryPacketPomelo.Handshake, data)
	if err != nil {
		panic(err)
	}
}

func (p *PomeloComponent) initHeartbeatData() {
	var err error
	hbd, err = p.PacketEncode.Encode(cherryPacketPomelo.Heartbeat, nil)
	if err != nil {
		panic(err)
	}
}
