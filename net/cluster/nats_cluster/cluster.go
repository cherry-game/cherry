package cherryNatsCluster

import (
	"time"

	"google.golang.org/protobuf/proto"

	ccode "github.com/cherry-game/cherry/code"
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap/zapcore"
)

type (
	Cluster struct {
		app           cfacade.IApplication
		bufferSize    int
		prefix        string
		localSubject  string
		remoteSubject string
		replySubject  string
	}

	OptionFunc func(o *Cluster)
)

func New(app cfacade.IApplication, options ...OptionFunc) cfacade.ICluster {
	cluster := &Cluster{
		app:        app,
		bufferSize: 1024,
	}

	for _, option := range options {
		option(cluster)
	}

	return cluster
}

func (p *Cluster) loadNatsConfig() {
	natsConfig := cprofile.GetConfig("cluster").GetConfig("nats")
	if natsConfig.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	p.prefix = natsConfig.GetString("prefix", "node")
	p.localSubject = getLocalSubject(p.prefix, p.app.NodeType(), p.app.NodeID())
	p.remoteSubject = getRemoteSubject(p.prefix, p.app.NodeType(), p.app.NodeID())
	p.replySubject = getReplySubject(p.prefix, p.app.NodeType(), p.app.NodeID())

	cnats.NewPool(p.replySubject, natsConfig, true)
}

func (p *Cluster) Init() {
	p.loadNatsConfig()

	p.localProcess()
	p.remoteProcess()

	clog.Info("nats cluster execute OnInit().")
}

func (p *Cluster) Stop() {
	cnats.ConnectClose()

	clog.Info("nats cluster execute OnStop().")
}

func (p *Cluster) localProcess() {
	process := func(natsMsg *nats.Msg) {
		packet, err := cproto.UnmarshalPacket(natsMsg.Data)
		defer packet.Recycle()

		if err != nil {
			clog.Warnf("[localProcess] Unmarshal fail. [subject = %s, %s, err = %s]",
				natsMsg.Subject,
				packet.PrintLog(),
				err,
			)
			return
		}

		message := cfacade.BuildClusterMessage(packet)
		p.app.ActorSystem().PostLocal(&message)
	}

	createConnPool(p.localSubject, "local", process)
}

func (p *Cluster) remoteProcess() {
	process := func(natsMsg *nats.Msg) {
		packet, err := cproto.UnmarshalPacket(natsMsg.Data)
		defer packet.Recycle()

		if err != nil {
			clog.Warnf("[remoteProcess] Unmarshal fail. [subject = %s, %s, err = %v]",
				natsMsg.Subject,
				packet.PrintLog(),
				err,
			)
			return
		}

		message := cfacade.BuildClusterMessage(packet)

		if len(natsMsg.Reply) > 0 {
			message.Header = natsMsg.Header
			message.Reply = natsMsg.Reply
		}

		p.app.ActorSystem().PostRemote(&message)
	}

	createConnPool(p.remoteSubject, "remote", process)
}

func createConnPool(subject, queue string, cb nats.MsgHandler) {
	for _, conn := range cnats.GetPool() {
		if err := conn.QueueSubscribe(subject, queue, cb); err != nil {
			clog.Errorf("[%s] Create queue subscribe fail. [subject = %s, err = %v]",
				queue,
				subject,
				err,
			)
			break
		}
	}
}

func (p *Cluster) PublishLocal(nodeID string, clusterPacket *cproto.ClusterPacket) error {
	defer clusterPacket.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishLocal] get node type fail. [nodeID = %s, %s]",
			nodeID,
			clusterPacket.PrintLog(),
		)
		return err
	}

	subject := getLocalSubject(p.prefix, nodeType, nodeID)
	bytes, err := proto.Marshal(clusterPacket)
	if err != nil {
		return err
	}

	err = p.Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishLocal] [nodeID = %s, %s]",
			nodeID,
			clusterPacket.PrintLog(),
		)
	}

	return err
}

func (p *Cluster) PublishRemote(nodeID string, clusterPacket *cproto.ClusterPacket) error {
	defer clusterPacket.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			clusterPacket.PrintLog(),
			err,
		)
		return err
	}

	subject := getRemoteSubject(p.prefix, nodeType, nodeID)
	bytes, err := proto.Marshal(clusterPacket)
	if err != nil {
		clog.Warn(err)
		return err
	}

	err = p.Publish(subject, bytes)
	return err
}

func (p *Cluster) RequestRemote(nodeID string, request *cproto.ClusterPacket, timeout ...time.Duration) ([]byte, int32) {
	defer request.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			request.PrintLog(),
			err,
		)

		return nil, ccode.DiscoveryNotFoundNode
	}

	msg, err := proto.Marshal(request)
	if err != nil {
		clog.Debugf("[PublishRemote] Marshal fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			request.PrintLog(),
			err,
		)

		return nil, ccode.RPCMarshalError
	}

	subject := getRemoteSubject(p.prefix, nodeType, nodeID)
	natsData, err := cnats.GetConnect().RequestSync(subject, msg, timeout...)
	if err != nil {
		clog.Warnf("[RequestRemote] nats request fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			request.PrintLog(),
			err,
		)

		return nil, ccode.RPCNetError
	}

	rsp := &cproto.Response{}
	if err = proto.Unmarshal(natsData, rsp); err != nil {
		clog.Warnf("[RequestRemote] unmarshal fail. [nodeID = %s, %s, rsp = %v, err = %v]",
			nodeID,
			request.PrintLog(),
			rsp,
			err,
		)

		return nil, ccode.RPCUnmarshalError
	}

	return rsp.Data, rsp.Code
}

func (p *Cluster) Publish(subject string, data []byte) error {
	if !p.app.Running() {
		return cerr.ClusterClientIsStop
	}

	return cnats.GetConnect().Publish(subject, data)
}

func WithBufferSize(size int) OptionFunc {
	return func(o *Cluster) {
		o.bufferSize = size
	}
}
