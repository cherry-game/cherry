package pomelo

import (
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	ppacket "github.com/cherry-game/cherry/net/parser/pomelo/packet"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap/zapcore"
)

type (
	Command struct {
		writeBacklog    int
		sysData         map[string]interface{}
		heartbeatTime   time.Duration
		handshakeBytes  []byte
		heartbeatBytes  []byte
		onPacketFuncMap map[ppacket.Type]PacketFunc
		onDataRouteFunc DataRouteFunc
	}

	PacketFunc    func(agent *Agent, packet *ppacket.Packet)
	DataRouteFunc func(agent *Agent, route *pmessage.Route, msg *pmessage.Message)
)

const (
	DataHeartbeat  = "heartbeat"
	DataDict       = "dict"
	DataSerializer = "serializer"
)

var (
	cmd = Command{
		writeBacklog:    64,
		sysData:         make(map[string]interface{}),
		heartbeatTime:   60 * time.Second,
		handshakeBytes:  make([]byte, 0),
		heartbeatBytes:  make([]byte, 0),
		onPacketFuncMap: make(map[ppacket.Type]PacketFunc, 4),
		onDataRouteFunc: DefaultDataRoute,
	}
)

func (p *Command) init(app cfacade.IApplication) {
	p.setData(DataHeartbeat, p.heartbeatTime.Seconds())
	p.setData(DataDict, pmessage.GetDictionary())
	p.setData(DataSerializer, app.Serializer().Name())

	p.setHandshakeBytes()
	p.setHeartbeatBytes()

	p.setOnPacketFunc()

}

func (p *Command) setData(name string, value interface{}) {
	if _, found := p.sysData[name]; !found {
		p.sysData[name] = value
	}
}

func (p *Command) setHandshakeBytes() {
	handshakeData := map[string]interface{}{
		"code": 200,
		"sys":  p.sysData,
	}

	handshakeBytes, err := jsoniter.Marshal(handshakeData)
	if err != nil {
		clog.Error(err)
		return
	}

	p.handshakeBytes, err = ppacket.Encode(ppacket.Handshake, handshakeBytes)
	if err != nil {
		clog.Error(err)
		return
	}

	clog.Infof("[initCommand] handshake data = %v", handshakeData)
}

func (p *Command) setHeartbeatBytes() {
	heartbeatBytes, err := ppacket.Encode(ppacket.Heartbeat, nil)
	if err != nil {
		clog.Error(err)
		return
	}

	p.heartbeatBytes = heartbeatBytes
}

func (p *Command) setOnPacketFunc() {
	packetFuncMaps := map[ppacket.Type]PacketFunc{
		ppacket.Handshake:    handshakeCommand,
		ppacket.HandshakeAck: handshakeACKCommand,
		ppacket.Heartbeat:    heartbeatCommand,
		ppacket.Data:         dataCommand,
	}

	for name, packetFunc := range packetFuncMaps {
		_, found := p.onPacketFuncMap[name]
		if !found {
			p.onPacketFuncMap[name] = packetFunc
		}
	}
}

func handshakeCommand(agent *Agent, _ *ppacket.Packet) {
	agent.SetState(AgentWaitAck)
	agent.SendRaw(cmd.handshakeBytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Request handshake. [address = %s]",
			agent.SID(),
			agent.UID(),
			agent.RemoteAddr(),
		)
	}
}

func handshakeACKCommand(agent *Agent, _ *ppacket.Packet) {
	agent.SetState(AgentWorking)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] request handshakeACK. [address = %s]",
			agent.SID(),
			agent.UID(),
			agent.RemoteAddr(),
		)
	}
}

func heartbeatCommand(agent *Agent, _ *ppacket.Packet) {
	agent.SendRaw(cmd.heartbeatBytes)
}

func dataCommand(agent *Agent, pkg *ppacket.Packet) {
	if agent.State() != AgentWorking {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Data State is not working. [state = %d]",
				agent.SID(),
				agent.UID(),
				agent.State(),
			)
		}
		return
	}

	msg, err := pmessage.Decode(pkg.Data())
	if err != nil {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Data message decode error. [data = %s, error = %s]",
				agent.SID(),
				agent.UID(),
				pkg.Data(),
				err,
			)
		}
		return
	}

	route, err := pmessage.DecodeRoute(msg.Route)
	if err != nil {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Data Message decode route error. [data = %s, error = %s]",
				agent.SID(),
				agent.UID(),
				pkg.Data(),
				err,
			)
		}
		return
	}

	cmd.onDataRouteFunc(agent, route, &msg)
}
