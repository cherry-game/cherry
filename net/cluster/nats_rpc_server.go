package cherryCluster

import (
	cherryCode "github.com/cherry-game/cherry/code"
	cherryConst "github.com/cherry-game/cherry/const"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	cherryNats "github.com/cherry-game/cherry/net/cluster/nats"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"github.com/nats-io/nats.go"
)

type (
	NatsRPCServer struct {
		cherryFacade.IApplication
		localChan        chan *nats.Msg
		remoteChan       chan *nats.Msg
		handlerComponent *cherryHandler.Component
		rpcClient        cherryFacade.RPCClient
	}
)

func NewRPCServer(rpcClient cherryFacade.RPCClient) *NatsRPCServer {
	return &NatsRPCServer{
		localChan:  make(chan *nats.Msg, 2048),
		remoteChan: make(chan *nats.Msg, 2048),
		rpcClient:  rpcClient,
	}
}

func (n *NatsRPCServer) Init(app cherryFacade.IApplication) {
	n.IApplication = app
	n.loadHandler()

	go n.processLocal()
	go n.processRemote()
}

func (n *NatsRPCServer) OnStop() {
}

func (n *NatsRPCServer) loadHandler() {
	handlerComponent, found := n.Find(cherryConst.HandlerComponent).(*cherryHandler.Component)
	if found == false {
		cherryLogger.Fatalf("%s not found", cherryConst.HandlerComponent)
	}
	n.handlerComponent = handlerComponent
}

func (n *NatsRPCServer) processLocal() {
	nodeSubject := GetLocalNodeSubject(n.NodeType(), n.NodeId())

	_, err := cherryNats.Conn().ChanSubscribe(nodeSubject, n.localChan)
	if err != nil {
		cherryLogger.Errorf("chan subscribe fail. [error = %s]", err)
		return
	}

	if n.IsFrontend() {
		// to client
		for local := range n.localChan {
			packet := &cherryProto.LocalPacket{}
			err := n.Unmarshal(local.Data, packet)
			if err != nil {
				cherryLogger.Warnf("unmarshal fail. [packet = %s] [error = %s]", packet, err)
				continue
			}

			if cherryMessage.Type(packet.MsgType) != cherryMessage.Response {
				cherryLogger.Warnf("message type not Request. [packet = %s]", packet)
				continue
			}

			session, found := cherrySession.GetByUID(packet.Session.Uid)
			if found == false {
				cherryLogger.Warnf("uid not found. [packet = %v]", packet)
				continue
			}

			session.ResponseMID(uint(packet.MsgId), packet.Data, packet.IsError)
		}
	} else {
		for local := range n.localChan {
			packet := &cherryProto.LocalPacket{}
			err := n.Unmarshal(local.Data, packet)
			if err != nil {
				cherryLogger.Warnf("unmarshal fail. [packet = %s] [error = %s]", packet, err)
				continue
			}

			if packet.Data == nil {
				packet.Data = []byte{}
			}

			// new fake session for backend node
			agent := cherryAgent.NewAgentBackend(n.IApplication, n.rpcClient, packet.Session.Ip)
			session := cherrySession.FakeSession(packet.Session, &agent)
			agent.SetSession(session)

			// build message
			message := &cherryMessage.Message{
				Type:  cherryMessage.Type(packet.MsgType),
				ID:    uint(packet.MsgId),
				Route: packet.Route,
				Data:  packet.Data,
				Error: packet.IsError,
			}

			if n.handlerComponent == nil {
				cherryLogger.Warnf("handler component not found. [packet = %v]", packet)
				return
			}

			n.handlerComponent.ProcessLocal(session, message)
		}
	}
}

func (n *NatsRPCServer) processRemote() {
	nodeSubject := GetRemoteNodeSubject(n.NodeType(), n.NodeId())
	_, err := cherryNats.Conn().ChanSubscribe(nodeSubject, n.remoteChan)
	if err != nil {
		cherryLogger.Errorf("chan subscribe fail. [error = %s]", err)
		return
	}

	for remote := range n.remoteChan {
		packet := &cherryProto.RemotePacket{}

		if err := n.Unmarshal(remote.Data, packet); err != nil {
			cherryLogger.Warnf("unmarshal fail. [packet = %s] [error = %s]", packet, err)
			n.replyError(remote, cherryCode.RPCUnmarshalError)
			continue
		}

		if packet.Data == nil {
			packet.Data = []byte{}
		}

		if n.handlerComponent == nil {
			cherryLogger.Warnf("handler component not found. [packet = %v]", packet)
			n.replyError(remote, cherryCode.RPCHandlerError)
			continue
		}

		rt, group, handler, found := n.handlerComponent.GetHandler(packet.Route)
		if found == false {
			cherryLogger.Warnf("handler not found. [packet = %v]", packet)
			n.replyError(remote, cherryCode.RPCHandlerError)
			continue
		}

		fn, found := handler.RemoteHandler(rt.Method())
		if found == false {
			cherryLogger.Debugf("could not find [method = %s] for [Route = %v].", rt.Method(), packet.Route)
			n.replyError(remote, cherryCode.RPCHandlerError)
			continue
		}

		executor := &cherryHandler.ExecutorRemote{
			IApplication: n.IApplication,
			HandlerFn:    fn,
			RemotePacket: packet,
			NatsMsg:      remote,
		}

		n.handlerComponent.ProcessRemote(group, executor)
	}
}

func (n *NatsRPCServer) replyError(msg *nats.Msg, code int32) {
	if msg.Reply == "" {
		return
	}

	rsp := &cherryProto.Response{
		Code: code,
	}

	rspData, _ := n.Marshal(rsp)
	err := msg.Respond(rspData)
	if err != nil {
		cherryLogger.Warn(err)
	}
}
