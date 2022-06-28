package cherryCluster

import (
	ccode "github.com/cherry-game/cherry/code"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cagent "github.com/cherry-game/cherry/net/agent"
	cnats "github.com/cherry-game/cherry/net/cluster/nats"
	chandler "github.com/cherry-game/cherry/net/handler"
	cmsg "github.com/cherry-game/cherry/net/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	csession "github.com/cherry-game/cherry/net/session"
	"github.com/nats-io/nats.go"
)

type (
	NatsRPCServer struct {
		cfacade.IApplication
		handlerComponent *chandler.Component
		rpcClient        cfacade.RPCClient
		remoteChan       chan *nats.Msg
		localChan        chan *nats.Msg
		pushChan         chan *nats.Msg
		kickChan         chan *nats.Msg
		msgBufferSize    int
	}
)

func NewNatsRPCServer(app cfacade.IApplication, rpcClient cfacade.RPCClient) *NatsRPCServer {
	return &NatsRPCServer{
		IApplication:  app,
		msgBufferSize: 2048,
		rpcClient:     rpcClient,
	}
}

func (n *NatsRPCServer) SetMsgBufferSize(bufferSize int) {
	n.msgBufferSize = bufferSize
}

func (n *NatsRPCServer) Init(app cfacade.IApplication) {
	n.IApplication = app
	n.localChan = make(chan *nats.Msg, n.msgBufferSize)
	n.remoteChan = make(chan *nats.Msg, n.msgBufferSize)

	go n.subscribeRemote()

	if n.IsFrontend() {
		go n.subscribeFrontendLocal()
		go n.subscribeFrontendPush()
		go n.subscribeFrontendKick()
	} else {
		go n.subscribeLocal()
	}
}

func (n *NatsRPCServer) OnStop() {
}

func (n *NatsRPCServer) subscribeRemote() {
	nodeSubject := getRemoteSubject(n.NodeType(), n.NodeId())
	chanSubscribe, chanErr := cnats.ChanSubscribe(nodeSubject, n.remoteChan)
	if chanErr != nil {
		clog.Errorf("chan subscribe fail. [subject = %s, err = %s]", nodeSubject, chanErr)
		return
	}

	process := func(msg *nats.Msg) {
		if dropped, err := chanSubscribe.Dropped(); err != nil {
			clog.Errorf("remote chan dropped messages. [dropped = %d, err = %v]", dropped, err)
		}

		packet := cproto.GetRequest()
		defer cproto.PutRequest(packet)

		if err := n.Unmarshal(msg.Data, packet); err != nil {
			clog.Warnf("unmarshal fail. [packet = %s, err = %s]", packet, err)
			n.replyError(msg, ccode.RPCUnmarshalError)
			return
		}

		if packet.Data == nil {
			packet.Data = []byte{}
		}

		statusCode := n.handlerComponent.ProcessRemote(packet.Route, packet.Data, msg)
		if ccode.IsFail(statusCode) {
			n.replyError(msg, statusCode)
		}
	}

	for msg := range n.remoteChan {
		process(msg)
	}
}

func (n *NatsRPCServer) subscribeLocal() {
	nodeSubject := getLocalSubject(n.NodeType(), n.NodeId())
	_, chanErr := cnats.ChanSubscribe(nodeSubject, n.localChan)
	if chanErr != nil {
		clog.Errorf("chan subscribe fail. [subject = %s, err = %s]", nodeSubject, chanErr)
		return
	}

	process := func(msg *nats.Msg) {
		request := cproto.GetRequest()
		defer cproto.PutRequest(request)

		err := n.Unmarshal(msg.Data, request)
		if err != nil {
			clog.Warnf("unmarshal fail. [packet = %s, error = %s]", request, err)
			return
		}

		if request.Data == nil {
			request.Data = []byte{}
		}

		// new fake session for backend node
		agent := cagent.NewAgentBackend(n.IApplication, n.rpcClient, request.Ip)
		session := csession.FakeSession(request, &agent)
		agent.SetSession(session)

		// build message
		message := &cmsg.Message{
			Type:  cmsg.Type(request.MsgType),
			ID:    uint(request.MsgId),
			Route: request.Route,
			Data:  request.Data,
			Error: request.IsError,
		}

		n.handlerComponent.ProcessLocal(session, message)
	}

	for msg := range n.localChan {
		process(msg)
	}
}

// subscribeFrontendLocal subscribe message write to client
func (n *NatsRPCServer) subscribeFrontendLocal() {
	nodeSubject := getLocalSubject(n.NodeType(), n.NodeId())
	_, chanErr := cnats.ChanSubscribe(nodeSubject, n.localChan)
	if chanErr != nil {
		clog.Errorf("chan subscribe fail. [subject = %s, err = %s]", nodeSubject, chanErr)
		return
	}

	process := func(msg *nats.Msg) {
		request := cproto.GetRequest()
		defer cproto.PutRequest(request)

		err := n.Unmarshal(msg.Data, request)
		if err != nil {
			clog.Warnf("unmarshal fail. [packet = %s, err = %s]", request, err)
			return
		}

		if cmsg.Type(request.MsgType) != cmsg.Response {
			clog.Warnf("message type not Request. [packet = %s]", request)
			return
		}

		if session, found := csession.GetByUID(request.Uid); found {
			session.ResponseMID(uint(request.MsgId), request.Data, request.IsError)
		} else {
			clog.Warnf("uid not found. [response = %v]", request)
		}
	}

	for msg := range n.localChan {
		process(msg)
	}

}

// subscribeFrontendPush subscribe message write to client
func (n *NatsRPCServer) subscribeFrontendPush() {
	nodeSubject := getPushSubject(n.NodeType(), n.NodeId())
	_, chanErr := cnats.ChanSubscribe(nodeSubject, n.pushChan)
	if chanErr != nil {
		clog.Errorf("chan subscribe fail. [subject = %s, err = %s]", nodeSubject, chanErr)
		return
	}

	// write to client
	process := func(msg *nats.Msg) {
		push := &cproto.Push{}
		err := n.Unmarshal(msg.Data, push)
		if err != nil {
			clog.Error("unmarshal kick error. [error = %v]", err)
			return
		}

		if session, found := csession.GetByUID(push.Uid); found {
			session.Push(push.Route, push.Data)
		} else {
			clog.Warnf("uid not found. [push = %v]", push)
		}
	}

	for msg := range n.pushChan {
		process(msg)
	}
}

// subscribeFrontendKick subscribe message write to client
func (n *NatsRPCServer) subscribeFrontendKick() {
	nodeSubject := getKickSubject(n.NodeType())
	_, chanErr := cnats.Subscribe(nodeSubject, func(msg *nats.Msg) {
		kick := &cproto.Kick{}
		err := n.Unmarshal(msg.Data, kick)
		if err != nil {
			clog.Error("unmarshal kick error. [error = %v]", err)
			return
		}

		if session, found := csession.GetByUID(kick.Uid); found {
			session.Kick(kick.Data, kick.Close)
		}
	})

	if chanErr != nil {
		clog.Error("subscribe fail. [subject = %s, err = %s]", nodeSubject, chanErr)
	}
}

func (n *NatsRPCServer) replyError(msg *nats.Msg, code int32) {
	if msg.Reply == "" {
		return
	}

	rsp := &cproto.Response{
		Code: code,
	}

	rspData, _ := n.Marshal(rsp)
	err := msg.Respond(rspData)
	if err != nil {
		clog.Warn(err)
	}
}
