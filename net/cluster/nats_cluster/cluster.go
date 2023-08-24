package cherryNatsCluster

import (
	"time"

	ccode "github.com/cherry-game/cherry/code"
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap/zapcore"
)

type (
	Cluster struct {
		app        cfacade.IApplication
		bufferSize int
		local      *natsSubject
		remote     *natsSubject
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

	cluster.loadNats()

	localSubject := getLocalSubject(app.NodeType(), app.NodeId())
	cluster.local = newNatsSubject(localSubject, cluster.bufferSize)

	remoteSubject := getRemoteSubject(app.NodeType(), app.NodeId())
	cluster.remote = newNatsSubject(remoteSubject, cluster.bufferSize)

	return cluster
}

func (p *Cluster) loadNats() {
	natsConfig := cprofile.GetConfig("cluster").GetConfig("nats")
	if natsConfig.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	natsConn := cnats.NewFromConfig(natsConfig)
	cnats.SetInstance(natsConn)
}

func (p *Cluster) Init() {
	cnats.Get().Connect()

	go p.localProcess()
	go p.remoteProcess()

	clog.Info("nats cluster execute OnInit().")
}

func (p *Cluster) Stop() {
	p.local.stop()
	p.remote.stop()

	cnats.Get().Close()

	clog.Info("nats cluster execute OnStop().")
}

func (p *Cluster) localProcess() {
	var err error
	p.local.subscription, err = cnats.Get().ChanSubscribe(p.local.subject, p.local.ch)
	if err != nil {
		clog.Errorf("[localProcess] Subscribe fail. [subject = %s, err = %s]", p.local.subject, err)
		return
	}

	process := func(natsMsg *nats.Msg) {
		if dropped, err := p.local.subscription.Dropped(); err != nil {
			clog.Errorf("[localProcess] Dropped messages. [subject = %s, dropped = %d, err = %v]",
				p.local.subject,
				dropped,
				err,
			)
		}

		packet := cproto.GetClusterPacket()
		defer packet.Recycle()

		err = proto.Unmarshal(natsMsg.Data, packet)
		if err != nil {
			clog.Warnf("[localProcess] Unmarshal fail. [subject = %s, %s, err = %s]",
				natsMsg.Subject,
				packet.PrintLog(),
				err,
			)
			return
		}

		message := cfacade.GetMessage()
		message.BuildTime = packet.BuildTime
		message.Source = packet.SourcePath
		message.Target = packet.TargetPath
		message.FuncName = packet.FuncName
		message.IsCluster = true
		message.Session = packet.Session
		message.Args = packet.ArgBytes

		p.app.ActorSystem().PostLocal(message)
	}

	for msg := range p.local.ch {
		process(msg)
	}
}

func (p *Cluster) remoteProcess() {
	var err error
	p.remote.subscription, err = cnats.Get().ChanSubscribe(p.remote.subject, p.remote.ch)
	if err != nil {
		clog.Errorf("[remoteProcess] Subscribe fail. [subject = %s, err = %s]", p.remote.subject, err)
		return
	}

	process := func(natsMsg *nats.Msg) {
		if dropped, err := p.remote.subscription.Dropped(); err != nil {
			clog.Errorf("[remoteProcess] Dropped messages. [subject = %s, dropped = %d, err = %v]",
				p.remote.subject,
				dropped,
				err,
			)
		}

		packet := cproto.GetClusterPacket()
		defer packet.Recycle()

		err = proto.Unmarshal(natsMsg.Data, packet)
		if err != nil {
			clog.Warnf("[remoteProcess] Unmarshal fail. [subject = %s, %s, err = %v]",
				natsMsg.Subject,
				packet.PrintLog(),
				err,
			)
			return
		}

		message := cfacade.GetMessage()
		message.BuildTime = packet.BuildTime
		message.Source = packet.SourcePath
		message.Target = packet.TargetPath
		message.FuncName = packet.FuncName
		if packet.ArgBytes != nil {
			message.Args = packet.ArgBytes
		}

		message.IsCluster = true
		if len(natsMsg.Reply) > 0 {
			message.ClusterReply = natsMsg
		}

		p.app.ActorSystem().PostRemote(message)
	}

	for msg := range p.remote.ch {
		process(msg)
	}
}

func (p *Cluster) PublishLocal(nodeId string, request *cproto.ClusterPacket) error {
	defer request.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeId)
	if err != nil {
		clog.Debugf("[PublishLocal] get node type fail. [nodeId = %s, %s]",
			nodeId,
			request.PrintLog(),
		)
		return err
	}

	subject := getLocalSubject(nodeType, nodeId)
	bytes, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	err = p.Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishLocal] [nodeId = %s, %s]",
			nodeId,
			request.PrintLog(),
		)
	}

	return err
}

func (p *Cluster) PublishRemote(nodeId string, request *cproto.ClusterPacket) error {
	defer request.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeId)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeId = %s, %s, err = %v]",
			nodeId,
			request.PrintLog(),
			err,
		)
		return err
	}

	subject := getRemoteSubject(nodeType, nodeId)
	bytes, err := proto.Marshal(request)
	if err != nil {
		clog.Warn(err)
		return err
	}

	err = p.Publish(subject, bytes)
	return err
}

func (p *Cluster) RequestRemote(nodeId string, request *cproto.ClusterPacket, timeout ...time.Duration) cproto.Response {
	defer request.Recycle()

	rsp := cproto.Response{}
	nodeType, err := p.app.Discovery().GetType(nodeId)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeId = %s, %s, err = %v]",
			nodeId,
			request.PrintLog(),
			err,
		)

		rsp.Code = ccode.DiscoveryNotFoundNode
		return rsp
	}

	msg, err := proto.Marshal(request)
	if err != nil {
		clog.Debugf("[PublishRemote] Marshal fail. [nodeId = %s, %s, err = %v]",
			nodeId,
			request.PrintLog(),
			err,
		)

		rsp.Code = ccode.RPCMarshalError
		return rsp
	}

	subject := getRemoteSubject(nodeType, nodeId)
	natsMsg, err := cnats.Get().Request(subject, msg, timeout...)
	if err != nil {
		clog.Warnf("[RequestRemote] nats request fail. [nodeId = %s, %s, err = %v]",
			nodeId,
			request.PrintLog(),
			err,
		)

		rsp.Code = ccode.RPCNetError
		return rsp
	}

	if err = proto.Unmarshal(natsMsg.Data, &rsp); err != nil {
		clog.Warnf("[RequestRemote] unmarshal fail. [nodeId = %s, %s, rsp = %v, err = %v]",
			nodeId,
			request.PrintLog(),
			rsp,
			err,
		)

		rsp.Code = ccode.RPCUnmarshalError
		return rsp
	}

	return rsp
}

func (p *Cluster) Publish(subject string, data []byte) error {
	if !p.app.Running() {
		return cerr.ClusterRPCClientIsStop
	}

	return cnats.Get().Publish(subject, data)
}

func WithBufferSize(size int) OptionFunc {
	return func(o *Cluster) {
		o.bufferSize = size
	}
}
