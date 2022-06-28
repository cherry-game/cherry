package cherryCluster

import (
	ccode "github.com/cherry-game/cherry/code"
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cdiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cnats "github.com/cherry-game/cherry/net/cluster/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
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

func (n *NatsRPCClient) Publish(subject string, val interface{}) error {
	if n.Running() == false {
		return cerr.ClusterRPCClientIsStop
	}

	bytes, err := n.Marshal(val)
	if err != nil {
		return err
	}

	return cnats.Publish(subject, bytes)
}

func (n *NatsRPCClient) PublishPush(frontendId cfacade.FrontendId, push *cproto.Push) error {
	nodeType, err := cdiscovery.GetType(frontendId)
	if err != nil {
		clog.Warnf("[PublishPush] get nodeType fail. [frontendId = %s, push = %+v, err = %v]",
			frontendId,
			push,
			err,
		)
		return err
	}

	subject := getPushSubject(nodeType, frontendId)
	err = n.Publish(subject, push)

	if cprofile.Debug() {
		clog.Debugf("[PublishPush] [frontendId = %s, push = %+v, err= %v]",
			frontendId,
			push,
			err,
		)
	}

	return err
}

func (n *NatsRPCClient) PublishKick(nodeType string, kick *cproto.Kick) error {
	if nodeType == "" {
		clog.Warnf("[PublishKick] nodeType is nil. [nodeType = %s, kick = %+v]",
			nodeType,
			kick,
		)
		return cerr.NodeTypeIsNil
	}

	subject := getKickSubject(nodeType)
	err := n.Publish(subject, kick)

	if cprofile.Debug() {
		clog.Debugf("[PublishKick] [nodeType = %s, kick = %+v, err = %v]",
			nodeType,
			kick,
			err,
		)
	}

	return err
}

func (n *NatsRPCClient) PublishLocal(nodeId string, request *cproto.Request) error {
	nodeType, err := cdiscovery.GetType(nodeId)
	if err != nil {
		clog.Warnf("[PublishLocal] get nodeType fail. [nodeId = %s, req = %+v, err = %v]",
			nodeId,
			request,
			err,
		)
		return err
	}

	subject := getLocalSubject(nodeType, nodeId)
	err = n.Publish(subject, request)

	if cprofile.Debug() {
		clog.Debugf("[PublishLocal] [nodeId = %s, request = %+v, err = %v]",
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
		clog.Warnf("[PublishRemote] get nodeType fail. [nodeId = %s, req = %+v, err = %v]",
			nodeId,
			request,
			err,
		)
		return err
	}

	subject := getRemoteSubject(nodeType, nodeId)
	err = n.Publish(subject, request)

	if cprofile.Debug() {
		clog.Debugf("[PublishRemote] [nodeId = %s, request = %+v, err = %v]",
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
		clog.Warnf("[RequestRemote] get nodeType fail. [nodeId = %s, req = %+v, err = %v]",
			nodeId,
			request,
			err,
		)

		rsp.Code = ccode.DiscoveryNotFoundNode
		return rsp, err
	}

	msg, err := n.Marshal(request)
	if err != nil {
		clog.Warnf("[RequestRemote] marshal fail. [nodeId = %s, req = %+v, err = %v]",
			nodeId,
			request,
			err,
		)

		rsp.Code = ccode.RPCMarshalError
		return rsp, err
	}

	tt := cnats.App().RequestTimeout
	if len(timeout) > 0 {
		tt = timeout[0]
	}

	subject := getRemoteSubject(nodeType, nodeId)
	rspData, err := cnats.Request(subject, msg, tt)
	if err != nil {
		clog.Warnf("[RequestRemote] nats request fail. [nodeId = %s, req = %+v, err = %v]",
			nodeId,
			request,
			err,
		)

		rsp.Code = ccode.RPCNetError
		return rsp, err
	}

	if err = n.Unmarshal(rspData.Data, rsp); err != nil {
		clog.Warnf("[RequestRemote] unmarshal fail. [nodeId = %s, req = %+v, rsp = %v, err = %v]",
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

func (n *NatsRPCClient) BroadcastSessionBind(uid string) error {
	return nil
}
