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
		app        cfacade.IApplication
		bufferSize int
		prefix     string
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

	cluster.loadConfig()

	return cluster
}

func (p *Cluster) loadConfig() {
	natsConfig := cprofile.GetConfig("cluster").GetConfig("nats")
	if natsConfig.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	natsConn := cnats.NewFromConfig(natsConfig)
	cnats.SetInstance(natsConn)

	p.prefix = natsConfig.GetString("prefix", "node")

	localSubject := getLocalSubject(p.prefix, p.app.NodeType(), p.app.NodeID())
	p.local = newNatsSubject(localSubject, p.bufferSize)

	remoteSubject := getRemoteSubject(p.prefix, p.app.NodeType(), p.app.NodeID())
	p.remote = newNatsSubject(remoteSubject, p.bufferSize)
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

		p.app.ActorSystem().PostLocal(&message)
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

		p.app.ActorSystem().PostRemote(&message)
	}

	for msg := range p.remote.ch {
		process(msg)
	}
}

func (p *Cluster) PublishLocal(nodeID string, request *cproto.ClusterPacket) error {
	defer request.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishLocal] get node type fail. [nodeID = %s, %s]",
			nodeID,
			request.PrintLog(),
		)
		return err
	}

	subject := getLocalSubject(p.prefix, nodeType, nodeID)
	bytes, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	err = p.Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishLocal] [nodeID = %s, %s]",
			nodeID,
			request.PrintLog(),
		)
	}

	return err
}

func (p *Cluster) PublishRemote(nodeID string, request *cproto.ClusterPacket) error {
	defer request.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			request.PrintLog(),
			err,
		)
		return err
	}

	subject := getRemoteSubject(p.prefix, nodeType, nodeID)
	bytes, err := proto.Marshal(request)
	if err != nil {
		clog.Warn(err)
		return err
	}

	err = p.Publish(subject, bytes)
	return err
}

func (p *Cluster) RequestRemote(nodeID string, request *cproto.ClusterPacket, timeout ...time.Duration) cproto.Response {
	defer request.Recycle()

	rsp := cproto.Response{}
	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			request.PrintLog(),
			err,
		)

		rsp.Code = ccode.DiscoveryNotFoundNode
		return rsp
	}

	msg, err := proto.Marshal(request)
	if err != nil {
		clog.Debugf("[PublishRemote] Marshal fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			request.PrintLog(),
			err,
		)

		rsp.Code = ccode.RPCMarshalError
		return rsp
	}

	subject := getRemoteSubject(p.prefix, nodeType, nodeID)
	natsMsg, err := cnats.Get().Request(subject, msg, timeout...)
	if err != nil {
		clog.Warnf("[RequestRemote] nats request fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			request.PrintLog(),
			err,
		)

		rsp.Code = ccode.RPCNetError
		return rsp
	}

	if err = proto.Unmarshal(natsMsg.Data, &rsp); err != nil {
		clog.Warnf("[RequestRemote] unmarshal fail. [nodeID = %s, %s, rsp = %v, err = %v]",
			nodeID,
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
