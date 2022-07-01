package cherryCluster

import (
	ccode "github.com/cherry-game/cherry/code"
	cherryConst "github.com/cherry-game/cherry/const"
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
		msgBufferSize    int
	}
)

func NewNatsRPCServer(app cfacade.IApplication, rpcClient cfacade.RPCClient, bufferSize int) *NatsRPCServer {
	return &NatsRPCServer{
		IApplication:  app,
		rpcClient:     rpcClient,
		msgBufferSize: bufferSize,
	}
}

func (n *NatsRPCServer) Init() {
	found := false
	n.handlerComponent, found = n.Find(cherryConst.HandlerComponent).(*chandler.Component)
	if found == false {
		clog.Fatalf("%s not found", cherryConst.HandlerComponent)
	}

	go n.subscribeRemote()
	go n.subscribeLocal()

	if n.IsFrontend() {
		go n.subscribePush()
		go n.subscribeKick()
	}
}

func (n *NatsRPCServer) OnStop() {
	clog.Infof("execute nats rpc server OnStop()")
}

func (n *NatsRPCServer) subscribeRemote() {
	var (
		nodeSubject = getRemoteSubject(n.NodeType(), n.NodeId())
		remoteChan  = make(chan *nats.Msg, n.msgBufferSize)
	)

	chanSubscribe, chanErr := cnats.ChanSubscribe(nodeSubject, remoteChan)
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

		statusCode := n.handlerComponent.ProcessRemote(packet.Route, packet.Data, msg)
		if ccode.IsFail(statusCode) {
			n.replyError(msg, statusCode)
		}
	}

	for msg := range remoteChan {
		process(msg)
	}
}

func (n *NatsRPCServer) subscribeLocal() {
	var (
		nodeSubject = getLocalSubject(n.NodeType(), n.NodeId())
		localChan   = make(chan *nats.Msg, n.msgBufferSize)
	)

	_, chanErr := cnats.ChanSubscribe(nodeSubject, localChan)
	if chanErr != nil {
		clog.Errorf("chan subscribe fail. [subject = %s, err = %s]", nodeSubject, chanErr)
		return
	}

	if n.IsFrontend() {
		for msg := range localChan {
			n.frontendLocalProcess(msg)
		}
	} else {
		for msg := range localChan {
			n.backendLocalProcess(msg)
		}
	}
}

func (n *NatsRPCServer) frontendLocalProcess(msg *nats.Msg) {
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

func (n *NatsRPCServer) backendLocalProcess(msg *nats.Msg) {
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

// subscribePush subscribe message write to client
func (n *NatsRPCServer) subscribePush() {
	var (
		nodeSubject = getPushSubject(n.NodeType(), n.NodeId())
		pushChan    = make(chan *nats.Msg, n.msgBufferSize)
		process     = func(msg *nats.Msg) {
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
	)

	cnats.ChanExecute(nodeSubject, pushChan, process)
}

// subscribeKick subscribe message write to client
func (n *NatsRPCServer) subscribeKick() {
	var (
		nodeSubject = getKickSubject(n.NodeType(), n.NodeId())
		kickChan    = make(chan *nats.Msg, n.msgBufferSize)
		process     = func(msg *nats.Msg) {
			kick := &cproto.Kick{}
			err := n.Unmarshal(msg.Data, kick)
			if err != nil {
				clog.Error("unmarshal kick error. [error = %v]", err)
				return
			}

			if session, found := csession.GetByUID(kick.Uid); found {
				session.Kick(kick.Data, kick.Close)
			}
		}
	)

	cnats.ChanExecute(nodeSubject, kickChan, process)
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
