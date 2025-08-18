package cherryNatsCluster

import (
	"time"

	"google.golang.org/protobuf/proto"

	ccode "github.com/cherry-game/cherry/code"
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
		prefix        string
		localSubject  string
		remoteSubject string
		replySubject  string
	}
)

func New(app cfacade.IApplication) cfacade.ICluster {
	cluster := &Cluster{
		app: app,
	}

	return cluster
}

func (p *Cluster) loadNatsConfig() {
	natsConfig := cprofile.GetConfig("cluster").GetConfig("nats")
	if natsConfig.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	p.prefix = natsConfig.GetString("prefix", "node")
	p.localSubject = GetLocalSubject(p.prefix, p.app.NodeType(), p.app.NodeID())
	p.remoteSubject = GetRemoteSubject(p.prefix, p.app.NodeType(), p.app.NodeID())
	p.replySubject = GetReplySubject(p.prefix, p.app.NodeType(), p.app.NodeID())

	cnats.NewPool(p.replySubject, natsConfig, true)
}

func (p *Cluster) Init() {
	p.loadNatsConfig()

	p.localProcess()
	p.remoteProcess()

	clog.Info("Nats cluster execute OnInit().")
}

func (p *Cluster) Stop() {
	cnats.ConnectClose()

	clog.Info("Nats cluster execute OnStop().")
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

	subscribeWithPool(p.localSubject, LocalType, process)
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

	subscribeWithPool(p.remoteSubject, RemoteType, process)
}

func subscribeWithPool(subject, queue string, cb nats.MsgHandler) {
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

func (p *Cluster) PublishLocal(nodeID string, cpacket *cproto.ClusterPacket) error {
	defer cpacket.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishLocal] get node type fail. [nodeID = %s, %s]",
			nodeID,
			cpacket.PrintLog(),
		)
		return err
	}

	bytes, err := proto.Marshal(cpacket)
	if err != nil {
		return err
	}

	subject := GetLocalSubject(p.prefix, nodeType, nodeID)
	err = cnats.GetConnect().Publish(subject, bytes)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[PublishLocal] [nodeID = %s, %s]",
			nodeID,
			cpacket.PrintLog(),
		)
	}

	return err
}

func (p *Cluster) PublishRemote(nodeID string, cpacket *cproto.ClusterPacket) error {
	defer cpacket.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)
		return err
	}

	bytes, err := proto.Marshal(cpacket)
	if err != nil {
		clog.Warn(err)
		return err
	}

	subject := GetRemoteSubject(p.prefix, nodeType, nodeID)
	err = cnats.GetConnect().Publish(subject, bytes)
	return err
}

func (p *Cluster) RequestRemote(nodeID string, cpacket *cproto.ClusterPacket, timeout ...time.Duration) ([]byte, int32) {
	defer cpacket.Recycle()

	nodeType, err := p.app.Discovery().GetType(nodeID)
	if err != nil {
		clog.Debugf("[PublishRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return nil, ccode.DiscoveryNotFoundNode
	}

	msg, err := proto.Marshal(cpacket)
	if err != nil {
		clog.Debugf("[PublishRemote] Marshal fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return nil, ccode.RPCMarshalError
	}

	subject := GetRemoteSubject(p.prefix, nodeType, nodeID)
	natsData, err := cnats.GetConnect().RequestSync(subject, msg, timeout...)
	if err != nil {
		clog.Warnf("[RequestRemote] nats request fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return nil, ccode.RPCNetError
	}

	rsp := &cproto.Response{}
	if err = proto.Unmarshal(natsData, rsp); err != nil {
		clog.Warnf("[RequestRemote] unmarshal fail. [nodeID = %s, %s, rsp = %v, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			rsp,
			err,
		)

		return nil, ccode.RPCUnmarshalError
	}

	return rsp.Data, rsp.Code
}
