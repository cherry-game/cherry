package cherryCluster

import (
	"fmt"
	"time"

	ccode "github.com/cherry-game/cherry/code"
	cerror "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

const (
	Name = "cluster_component"
)

type (
	Component struct {
		cfacade.Component
		natsSubjects
		prefix                string // cluster prefix
		localSubject          string // local subject
		remoteSubject         string // remote subject
		replySubject          string // reply subject
		remoteNodeTypeSubject string // remote node type subject
	}

	natsSubjects struct {
		localSubjectFormat      string // cherry.{prefix}.local.{nodeType}.{nodeID}
		remoteSubjectFormat     string // cherry.{prefix}.remote.{nodeType}.{nodeID}
		remoteTypeSubjectFormat string // cherry.{prefix}.remoteType.{nodeType}
		replySubjectFormat      string // cherry.{prefix}.reply.{nodeType}.{nodeID}
	}
)

func New() *Component {
	return &Component{
		natsSubjects: natsSubjects{
			localSubjectFormat:      "cherry-%s.local.%s.%s",
			remoteSubjectFormat:     "cherry-%s.remote.%s.%s",
			remoteTypeSubjectFormat: "cherry-%s.remoteType.%s",
			replySubjectFormat:      "cherry-%s.reply.%s.%s",
		},
	}
}

func (*Component) Name() string {
	return Name
}

func (*Component) Mode() string {
	return "nats"
}

func (p *Component) Init() {
	p.loadNatsConfig()

	p.localProcess()
	p.remoteProcess()
	p.remoteTypeProcess()

	clog.Info("Nats cluster execute OnInit().")
}

func (p *Component) OnStop() {
	cnats.PoolClose()
	clog.Info("Nats cluster execute OnStop().")
}

func (p *Component) loadNatsConfig() {
	natsConfig := cprofile.GetConfig("cluster").GetConfig("nats")
	if natsConfig.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	p.prefix = natsConfig.GetString("prefix", "node")
	p.localSubject = p.GetLocalSubject(p.prefix, p.App().NodeType(), p.App().NodeID())
	p.remoteSubject = p.GetRemoteSubject(p.prefix, p.App().NodeType(), p.App().NodeID())
	p.remoteNodeTypeSubject = p.GetRemoteTypeSubject(p.prefix, p.App().NodeType())
	p.replySubject = p.GetReplySubject(p.prefix, p.App().NodeType(), p.App().NodeID())

	cnats.InitPool(p.replySubject, natsConfig, true)
}

func (p *Component) localProcess() {
	process := func(natsMsg *nats.Msg) {
		packet, err := cproto.UnmarshalPacket(natsMsg.Data)
		defer packet.Recycle()

		if err != nil {
			clog.Warnf("[localProcess] Unmarshal fail. [subject = %s, dataLen = %d, err = %s]",
				natsMsg.Subject,
				len(natsMsg.Data),
				err,
			)
			return
		}

		message := cfacade.BuildClusterMessage(packet)
		p.App().ActorSystem().PostLocal(&message)
	}

	err := cnats.Subscribe(p.localSubject, process)
	if err != nil {
		clog.Errorf("[localProcess] Create subscribe fail. [subject = %s, err = %v]",
			p.localSubject,
			err,
		)
	}
}

func (p *Component) remoteProcess() {
	process := func(natsMsg *nats.Msg) {
		packet, err := cproto.UnmarshalPacket(natsMsg.Data)
		defer packet.Recycle()

		if err != nil {
			clog.Warnf("[remoteProcess] Unmarshal fail. [subject = %s, dataLen = %d, err = %v]",
				natsMsg.Subject,
				len(natsMsg.Data),
				err,
			)
			return
		}

		message := cfacade.BuildClusterMessage(packet)

		if len(natsMsg.Reply) > 0 {
			message.ReqID = natsMsg.Header.Get(cnats.REQ_ID)
			message.Reply = natsMsg.Reply
		}

		p.App().ActorSystem().PostRemote(&message)
	}

	err := cnats.Subscribe(p.remoteSubject, process)
	if err != nil {
		clog.Errorf("[remoteProcess] Create subscribe fail. [subject = %s, err = %v]",
			p.remoteSubject,
			err,
		)
	}
}

func (p *Component) remoteTypeProcess() {
	process := func(natsMsg *nats.Msg) {
		packet, err := cproto.UnmarshalPacket(natsMsg.Data)
		defer packet.Recycle()

		if err != nil {
			clog.Warnf("[remoteTypeProcess] Unmarshal fail. [subject = %s, dataLen = %d, err = %v]",
				natsMsg.Subject,
				len(natsMsg.Data),
				err,
			)
			return
		}

		message := cfacade.BuildClusterMessage(packet)

		p.App().ActorSystem().PostRemote(&message)
	}

	err := cnats.Subscribe(p.remoteNodeTypeSubject, process)
	if err != nil {
		clog.Errorf("[remoteTypeProcess] Create subscribe fail. [subject = %s, err = %v]",
			p.remoteSubject,
			err,
		)
	}
}

func (p *Component) PublishLocal(nodeID string, cpacket *cproto.ClusterPacket) error {
	defer cpacket.Recycle()

	nodeType, err := p.App().Discovery().GetType(nodeID)
	if err != nil {
		clog.Warnf("[PublishLocal] Get node type fail. [nodeID = %s, packet = %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)
		return cerror.DiscoveryNotFoundNode
	}

	bytes, err := proto.Marshal(cpacket)
	if err != nil {
		clog.Warnf("[PublishLocal] Marshal error. [nodeID = %s, packet = %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)
		return cerror.ClusterPacketMarshalFail
	}

	subject := p.GetLocalSubject(p.prefix, nodeType, nodeID)
	err = cnats.Publish(subject, bytes)
	if err != nil {
		clog.Warnf("[PublishLocal] Nats publish fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return cerror.ClusterPublishFail
	}

	return nil
}

func (p *Component) PublishRemote(nodeID string, cpacket *cproto.ClusterPacket) error {
	defer cpacket.Recycle()

	nodeType, err := p.App().Discovery().GetType(nodeID)
	if err != nil {
		clog.Warnf("[PublishRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)
		return cerror.DiscoveryNotFoundNode
	}

	bytes, err := proto.Marshal(cpacket)
	if err != nil {
		clog.Warnf("[PublishRemote] Marshal error. [nodeID = %s, packet = %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)
		return cerror.ClusterPacketMarshalFail
	}

	subject := p.GetRemoteSubject(p.prefix, nodeType, nodeID)
	err = cnats.Publish(subject, bytes)
	if err != nil {
		clog.Warnf("[PublishRemote] Nats publish fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return cerror.ClusterPublishFail
	}

	return nil
}

func (p *Component) PublishRemoteType(nodeType string, cpacket *cproto.ClusterPacket) error {
	defer cpacket.Recycle()

	bytes, err := proto.Marshal(cpacket)
	if err != nil {
		clog.Warnf("[PublishRemoteType] Marshal error. [nodeType = %s, packet = %s, err = %v]",
			nodeType,
			cpacket.PrintLog(),
			err,
		)
		return cerror.ClusterPacketMarshalFail
	}

	if nodeType == "" {
		return cerror.ClusterNodeTypeIsNil
	}

	if members := p.App().Discovery().ListByType(nodeType); len(members) < 1 {
		return cerror.ClusterNodeTypeMemberNotFound
	}

	subject := p.GetRemoteTypeSubject(p.prefix, nodeType)
	err = cnats.Publish(subject, bytes)
	if err != nil {
		clog.Warnf("[PublishRemoteType] Nats publish fail. [nodeType = %s, %s, err = %v]",
			nodeType,
			cpacket.PrintLog(),
			err,
		)

		return cerror.ClusterPublishFail
	}

	return nil
}

func (p *Component) RequestRemote(nodeID string, cpacket *cproto.ClusterPacket, timeout ...time.Duration) ([]byte, int32) {
	defer cpacket.Recycle()

	nodeType, err := p.App().Discovery().GetType(nodeID)
	if err != nil {
		clog.Warnf("[RequestRemote] Get node type fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return nil, ccode.DiscoveryNotFoundNode
	}

	msg, err := proto.Marshal(cpacket)
	if err != nil {
		clog.Warnf("[RequestRemote] Marshal fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return nil, ccode.RPCMarshalError
	}

	reqID := cnats.NewStringReqID()
	subject := p.GetRemoteSubject(p.prefix, nodeType, nodeID)

	natsData, err := cnats.RequestSync(reqID, subject, msg, timeout...)
	if err != nil {
		clog.Warnf("[RequestRemote] Nats request fail. [nodeID = %s, %s, err = %v]",
			nodeID,
			cpacket.PrintLog(),
			err,
		)

		return nil, ccode.RPCRemoteExecuteError
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

func (p *natsSubjects) GetLocalSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(p.localSubjectFormat, prefix, nodeType, nodeID)
}

// GetRemoteSubject remote message nats chan
func (p *natsSubjects) GetRemoteSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(p.remoteSubjectFormat, prefix, nodeType, nodeID)
}

func (p *natsSubjects) GetRemoteTypeSubject(prefix, nodeType string) string {
	return fmt.Sprintf(p.remoteTypeSubjectFormat, prefix, nodeType)
}

func (p *natsSubjects) GetReplySubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(p.replySubjectFormat, prefix, nodeType, nodeID)
}
