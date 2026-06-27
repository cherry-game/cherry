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

// Command holds the parser-level configuration shared by all agents, including
// heartbeat timing, handshake payload, and the registered packet/data-route handlers.
type (
	Command struct {
		writeBacklog    int                          // write/pending channel buffer size
		sysData         map[string]interface{}       // system data sent during handshake
		heartbeatTime   time.Duration                // heartbeat interval
		handshakeBytes  []byte                       // pre-encoded handshake packet
		heartbeatBytes  []byte                       // pre-encoded heartbeat packet
		onPacketFuncMap map[ppacket.Type]PacketFunc  // packet type → handler
		onDataRouteFunc DataRouteFunc                // data message routing handler
	}

// PacketFunc is called when a packet of a registered type arrives.
	PacketFunc    func(agent *Agent, packet *ppacket.Packet)

// DataRouteFunc is called to route a decoded data message to the target actor.
	DataRouteFunc func(agent *Agent, route *pmessage.Route, msg *pmessage.Message)
)

// System data keys sent in the handshake response.
const (
	DataHeartbeat  = "heartbeat"  // heartbeat interval in seconds
	DataDict       = "dict"       // route compression dictionary
	DataSerializer = "serializer" // serializer name
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

// handshakeCommand is the packet handler for the initial client handshake.
// It sets the agent to AgentWaitAck and sends the pre-encoded handshake bytes.
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

// handshakeACKCommand is the packet handler for the handshake acknowledgement.
// It transitions the agent to AgentWorking state.
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

// heartbeatCommand is the packet handler for heartbeat packets.
// It replies with the pre-encoded heartbeat bytes.
func heartbeatCommand(agent *Agent, _ *ppacket.Packet) {
	agent.SendRaw(cmd.heartbeatBytes)
}

// dataCommand is the packet handler for data packets. It decodes the message,
// resolves the route, and dispatches via the configured onDataRouteFunc.
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
