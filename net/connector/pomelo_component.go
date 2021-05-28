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
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/net/session"
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
	UseHostFilter       bool //好像没什么用，或者以后单独做一个组件
	CheckClient         CheckClientFn
	DataCompression     bool
	IgnoreHandshakeStep bool //忽略握手的步骤
}

type PomeloComponent struct {
	cherryFacade.Component
	PomeloComponentOptions
	connectStat      *ConnectStat
	sessionComponent *cherrySession.SessionComponent
	handlerComponent *cherryHandler.HandlerComponent
}

// default options
func NewPomelo() *PomeloComponent {
	opts := PomeloComponentOptions{
		PacketEncode:        cherryPacket.NewPomeloEncoder(),
		PacketDecode:        cherryPacket.NewPomeloDecoder(),
		Serializer:          cherrySerializer.NewJSON(),
		BlackListFunc:       nil,
		BlackList:           nil,
		ForwardMessage:      false, //转发消息
		Heartbeat:           30,
		DisconnectOnTimeout: false,
		UseDict:             false,
		UseProtobuf:         false,
		UseCrypto:           false,
		UseHostFilter:       false,
		CheckClient:         nil,
		IgnoreHandshakeStep: false,
	}

	return NewPomeloWithOpts(opts)
}

func NewPomeloWithOpts(opts PomeloComponentOptions) *PomeloComponent {
	return &PomeloComponent{
		PomeloComponentOptions: opts,
		connectStat:            &ConnectStat{},
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
		p.Connector.OnConnect(p.defaultInitSession)
	}

	// new goroutine
	go p.Connector.OnStart()
}

func (p *PomeloComponent) OnStop() {
	if p.Connector != nil {
		p.Connector.OnStop()
	}
}

func (p *PomeloComponent) ConnectStat() *ConnectStat {
	return p.connectStat
}

func (p *PomeloComponent) defaultInitSession(conn net.Conn) {
	session := p.sessionComponent.Create(conn, nil) //TODO INetworkEntity

	// conn stat
	p.connectStat.IncreaseConn()

	if p.IgnoreHandshakeStep {
		// skip handshake
		session.SetStatus(cherrySession.Working)
	}

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

	if p.SessionOnClosed != nil {
		session.OnClose(p.SessionOnClosed)
	}

	session.OnClose(func(_ cherryFacade.ISession) {
		p.connectStat.DecreaseConn()
	})

	if p.SessionOnError != nil {
		session.OnError(p.SessionOnError)
	}

	//create a new goroutine to process read data for current socket
	session.Start()
}

func (p *PomeloComponent) processPacket(session cherryFacade.ISession, pkg *cherryPacket.Packet) error {
	switch pkg.Type {
	case cherryPacket.Handshake:
		if err := session.Send(hrd); err != nil {
			return err
		}
		session.SetStatus(cherrySession.WaitAck)
		cherryLogger.Debugf("[Handshake] session=[%s]", session)

	case cherryPacket.HandshakeAck:
		if session.Status() != cherrySession.WaitAck {
			cherryLogger.Warnf("[HandshakeAck] session=[%s]", session)
			session.Closed()
			return nil
		}
		session.SetStatus(cherrySession.Working)
		cherryLogger.Debugf("[HandshakeAck] session=[%s]", session)

	case cherryPacket.Data:
		if session.Status() != cherrySession.Working {
			return cherryUtils.Errorf("[Data] status error. session=[%s]", session)
		}

		msg, err := cherryMessage.Decode(pkg.Data)
		if err != nil {
			p.handleMessage(session, msg)
		}

	case cherryPacket.Heartbeat:
		d, err := p.PacketEncode.Encode(cherryPacket.Heartbeat, nil)
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

	unHandleMessage := &cherryHandler.UnhandledMessage{
		Session: session,
		Route:   route,
		Msg:     msg,
	}

	p.handlerComponent.DoHandle(unHandleMessage)
}

var (
	// hbd contains the heartbeat packet data
	hbd []byte
	// hrd contains the handshake response data
	hrd []byte
)

func (p *PomeloComponent) initHandshakeData() {
	data := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat":  p.Heartbeat,
			"dict":       cherryMessage.GetDictionary(),
			"serializer": p.Serializer.Name(),
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		cherryLogger.Error(err)
		return
	}

	if p.DataCompression {
		compressedData, err := cherryCompress.DeflateData(jsonData)
		if err != nil {
			cherryLogger.Error(err)
			return
		}

		if len(compressedData) < len(jsonData) {
			jsonData = compressedData
		}
	}

	hrd, err = p.PacketEncode.Encode(cherryPacket.Handshake, jsonData)
	if err != nil {
		cherryLogger.Error(err)
		return
	}
}

func (p *PomeloComponent) initHeartbeatData() {
	var err error
	hbd, err = p.PacketEncode.Encode(cherryPacket.Heartbeat, nil)
	if err != nil {
		cherryLogger.Error(err)
		return
	}
}
