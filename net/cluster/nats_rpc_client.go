package cherryCluster

import (
	ccode "github.com/cherry-game/cherry/code"
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cdiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cnats "github.com/cherry-game/cherry/net/cluster/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap/zapcore"
	"time"
)

type NatsRPCClient struct {
	cfacade.IApplication
}

func (n *NatsRPCClient) OnStop() {
	clog.Info("nats rpc client execute OnStop().")
}

func NewRPCClient(app cfacade.IApplication) *NatsRPCClient {
	return &NatsRPCClient{
		IApplication: app,
	}
}

func (n *NatsRPCClient) Publish(subject string, data []byte) error {
	if n.Running() == false {
		return cerr.ClusterRPCClientIsStop
	}

	return cnats.Publish(subject, data)
}

func (n *NatsRPCClient) PublishPush(frontendId cfacade.FrontendId, push *cproto.Push) error {
	nodeType, err := cdiscovery.GetType(frontendId)
	if err != nil {
		clog.Warnf("[PublishPush] get nodeType fail. [frontendId = %s, route = %s, uid = %d, dataLen = %d, err = %v]",
			frontendId,
			push.Route,
			push.Uid,
			len(push.Data),
			err,
		)
		return err
	}

	subject := getPushSubject(nodeType, frontendId)
	bytes, err := proto.Marshal(push)
	if err != nil {
		return err
	}

	err = n.Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishPush] [frontendId = %s, route = %s, uid = %d, dataLen = %d, err= %v]",
			frontendId,
			push.Route,
			push.Uid,
			len(push.Data),
			err,
		)
	}

	return err
}

func (n *NatsRPCClient) PublishKick(nodeId string, kick *cproto.Kick) error {
	nodeType, err := cdiscovery.GetType(nodeId)
	if err != nil {
		clog.Warnf("[PublishKick] get nodeType fail. [nodeId = %s, kick = {%+v}]",
			nodeId,
			kick,
		)
		return err
	}

	subject := getKickSubject(nodeType, nodeId)
	bytes, err := proto.Marshal(kick)
	if err != nil {
		return err
	}

	err = n.Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishKick] [nodeId = %s, kick = {%+v}, err = %v]",
			nodeId,
			kick,
			err,
		)
	}

	return err
}

func (n *NatsRPCClient) PublishLocal(nodeId string, request *cproto.Request) error {
	nodeType, err := cdiscovery.GetType(nodeId)
	if err != nil {
		clog.Warnf("[PublishLocal] get nodeType fail. [nodeId = %s, req = {%+v}, err = %v]",
			nodeId,
			request,
			err,
		)
		return err
	}

	subject := getLocalSubject(nodeType, nodeId)
	bytes, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	err = n.Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishLocal] [nodeId = %s, req = {%+v}, err = %v]",
			nodeId,
			request,
			err,
		)
	}

	return err
}

func (n *NatsRPCClient) PublishRemote(nodeId string, request *cproto.Request) error {
	nodeType, err := cdiscovery.GetType(nodeId)
	if err != nil {
		clog.Warnf("[PublishRemote] get nodeType fail. [nodeId = %s, req = {%+v}, err = %v]",
			nodeId,
			request,
			err,
		)
		return err
	}

	subject := getRemoteSubject(nodeType, nodeId)
	bytes, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	err = n.Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishRemote] [nodeId = %s, req = {%+v}, err = %v]",
			nodeId,
			request,
			err,
		)
	}

	return err
}

func (n *NatsRPCClient) RequestRemote(nodeId string, request *cproto.Request, timeout ...time.Duration) (*cproto.Response, error) {
	rsp := &cproto.Response{}

	nodeType, err := cdiscovery.GetType(nodeId)
	if err != nil {
		clog.Warnf("[RequestRemote] get nodeType fail. [nodeId = %s, req = {%+v}, err = %v]",
			nodeId,
			request,
			err,
		)

		rsp.Code = ccode.DiscoveryNotFoundNode
		return rsp, err
	}

	msg, err := proto.Marshal(request)
	if err != nil {
		clog.Warnf("[RequestRemote] marshal fail. [nodeId = %s, req = {%+v}, err = %v]",
			nodeId,
			request,
			err,
		)

		rsp.Code = ccode.RPCMarshalError
		return rsp, err
	}

	tt := cnats.App().RequestTimeout
	if len(timeout) > 0 && timeout[0] > 0 {
		tt = timeout[0]
	}

	subject := getRemoteSubject(nodeType, nodeId)
	natsMsg, err := cnats.Request(subject, msg, tt)
	if err != nil {
		clog.Warnf("[RequestRemote] nats request fail. [nodeId = %s, req = {%+v}, timeout = %d, err = %v]",
			nodeId,
			request,
			int(tt.Seconds()),
			err,
		)

		rsp.Code = ccode.RPCNetError
		return rsp, err
	}

	if err = proto.Unmarshal(natsMsg.Data, rsp); err != nil {
		clog.Warnf("[RequestRemote] unmarshal fail. [nodeId = %s, req = {%+v}, rsp = %v, err = %v]",
			nodeId,
			request,
			rsp,
			err,
		)

		rsp.Code = ccode.RPCUnmarshalError
		return rsp, err
	}

	return rsp, nil
}
