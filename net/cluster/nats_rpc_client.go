package cherryCluster

import (
	cherryCode "github.com/cherry-game/cherry/code"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryDiscovery "github.com/cherry-game/cherry/net/discovery"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"time"
)

const (
	PushRoute = "sys.sessionHandler.push"
	KickRoute = "sys.sessionHandler.kick"
)

type NatsRPCClient struct {
	cherryFacade.IApplication
	running    bool
	nats       *nats.Conn
	natsConfig *cherryProfile.NatsConfig
}

func NewNatsRPCClient(conn *nats.Conn, natsConfig *cherryProfile.NatsConfig) *NatsRPCClient {
	return &NatsRPCClient{
		nats:       conn,
		natsConfig: natsConfig,
	}
}

func (n *NatsRPCClient) Init(app cherryFacade.IApplication) {
	n.IApplication = app
	n.running = true
}

func (n *NatsRPCClient) OnStop() {
	n.running = false
}

func (n *NatsRPCClient) Publish(subject string, val interface{}) error {
	if !n.running {
		return cherryError.ClusterRPCClientIsStop
	}

	msg, err := n.Marshal(val)
	if err != nil {
		return err
	}

	return n.nats.Publish(subject, msg)
}

func (n *NatsRPCClient) CallLocal(nodeId string, packet *cherryProto.LocalPacket) error {
	nodeType, err := cherryDiscovery.GetType(nodeId)
	if err != nil {
		return err
	}

	subject := GetLocalNodeSubject(nodeType, nodeId)

	if cherryProfile.Debug() {
		msgType := cherryMessage.Type(packet.MsgType)
		cherryLogger.Debugf("[CallLocal] [uid = %d] [subject = %s] [msgType = %s] [msgId = %d] [isError = %v]",
			packet.Session.Uid,
			subject,
			msgType.String(),
			packet.MsgId,
			packet.IsError,
		)
	}

	return n.Publish(subject, packet)
}

func (n *NatsRPCClient) CallRemote(nodeId string, route string, val interface{}, timeout time.Duration) cherryProto.Response {
	rsp := cherryProto.Response{
		Code: cherryCode.OK,
	}

	nodeType, err := cherryDiscovery.GetType(nodeId)
	if err != nil {
		cherryLogger.Warnf("[CallRemote] get nodeType fail. [nodeId = %s, route = %s, val = %v] [error = %v]",
			nodeId,
			route,
			val,
			err,
		)

		rsp.Code = cherryCode.DiscoveryNotFoundNode
		return rsp
	}

	var data []byte
	if val != nil {
		data, err = n.Marshal(val)
		if err != nil {
			cherryLogger.Warnf("[CallRemote] marshal fail. [nodeId = %s, route = %s, val = %v] [error = %v]",
				nodeId,
				route,
				val,
				err,
			)

			rsp.Code = cherryCode.RPCMarshalError
			return rsp
		}
	}

	packet := &cherryProto.RemotePacket{
		Route: route,
		Data:  data,
	}

	msg, err := n.Marshal(packet)
	if err != nil {
		cherryLogger.Warnf("[CallRemote] marshal fail. [nodeId = %s, route = %s, val = %v] [error = %v]",
			nodeId,
			route,
			val,
			err,
		)

		rsp.Code = cherryCode.RPCMarshalError
		return rsp
	}

	if timeout < 1 {
		timeout = n.natsConfig.RequestTimeout
	}

	subject := GetRemoteNodeSubject(nodeType, nodeId)

	if cherryProfile.Debug() {
		cherryLogger.Debugf("[CallRemote] [route = %s] [val = %v]",
			packet.Route,
			val,
		)
	}

	rspData, err := n.nats.Request(subject, msg, timeout)
	if err != nil {
		cherryLogger.Warnf("[CallRemote] nats request fail. [nodeId = %s, route = %s, val = %v] [error = %v]",
			nodeId,
			route,
			val,
			err,
		)

		rsp.Code = cherryCode.RPCNetError
		return rsp
	}

	err = n.Unmarshal(rspData.Data, &rsp)
	if err != nil {
		cherryLogger.Warnf("[CallRemote] unmarshal fail. [nodeId = %s, route = %s, val = %v] [error = %v]",
			nodeId,
			route,
			val,
			err,
		)

		rsp.Code = cherryCode.RPCUnmarshalError
		return rsp
	}

	return rsp
}

func (n *NatsRPCClient) CallRemoteAsync(nodeId string, route string, val interface{}) {
	nodeType, err := cherryDiscovery.GetType(nodeId)
	if err != nil {
		cherryLogger.Warnf("[CallRemoteAsync] get nodeType error. [nodeId = %s] [error = %s]", nodeId, err)
		return
	}

	var data []byte
	if val != nil {
		data, err = n.Marshal(val)
		if err != nil {
			cherryLogger.Warnf("[CallRemoteAsync] marshal fail. [nodeId = %s, route = %d, val = %v] [error = %s].",
				nodeId,
				route,
				val,
				err,
			)
			return
		}
	}

	packet := &cherryProto.RemotePacket{
		Route: route,
		Data:  data,
	}

	subject := GetRemoteNodeSubject(nodeType, nodeId)

	if cherryProfile.Debug() {
		cherryLogger.Debugf("[CallRemoteAsync] [subject = %s] [route = %s]", subject, route)
	}

	err = n.Publish(subject, packet)
	if err != nil {
		cherryLogger.Warnf("[CallRemoteAsync] nats publish error. [subject = %s] [error = %s]", subject, err)
		return
	}
}

func (n *NatsRPCClient) SendKick(nodeId string, uid cherryFacade.UID, reason interface{}) {
	bytes, err := n.Marshal(reason)
	if err != nil {
		cherryLogger.Warnf("[SendKick] marshal fail. [uid = %d] [reason = %v] [error = %s].", uid, reason, err)
	}

	kickRequest := &cherryProto.KickRequest{
		Uid:    uid,
		Reason: bytes,
	}

	n.CallRemoteAsync(nodeId, KickRoute, kickRequest)
}

func (n *NatsRPCClient) SendPush(nodeId string, route string, uid cherryFacade.UID, val interface{}) {
	data, err := n.Marshal(val)
	if err != nil {
		cherryLogger.Warnf("[SendPush] marshal error. [route = %s, uid = %d, val = %v] [err = %v]", route, uid, val, err)
		return
	}

	pushRequest := &cherryProto.PushRequest{
		Route: route,
		Uid:   uid,
		Data:  data,
	}

	n.CallRemoteAsync(nodeId, PushRoute, pushRequest)
}
