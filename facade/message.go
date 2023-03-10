package cherryFacade

import (
	cconst "github.com/cherry-game/cherry/const"
	cerr "github.com/cherry-game/cherry/error"
	cproto "github.com/cherry-game/cherry/net/proto"
	"strings"
	"sync"
)

type (
	Message struct {
		Source       string           // 来源actor path
		Target       string           // 目标actor path
		targetPath   *ActorPath       // 目标actor path对象
		FuncName     string           // 请求调用的函数名
		Session      *cproto.Session  // session of gateway
		Args         interface{}      // 请求的参数
		EncodeArgs   bool             // 是否已解码args
		Err          error            // 返回的错误
		ClusterReply IRespond         // 返回消息的接口
		IsCluster    bool             // 是否为集群消息
		ChanResult   chan interface{} //
	}

	IRespond interface {
		Respond(data []byte) error
	}

	// ActorPath = NodeID . ActorID
	// ActorPath = NodeID . ActorID . ChildID
	ActorPath struct {
		NodeID  string
		ActorID string
		ChildID string
	}
)

var (
	messagePool = &sync.Pool{
		New: func() interface{} {
			return new(Message)
		},
	}
)

func GetMessage() *Message {
	return messagePool.Get().(*Message)
}

func (p *Message) Recycle() {
	p.Source = ""
	p.Target = ""
	p.targetPath = nil
	p.FuncName = ""
	p.Args = nil
	p.Err = nil
	p.ClusterReply = nil
	p.IsCluster = false
	p.ChanResult = nil
	messagePool.Put(p)
}

func (p *Message) TargetPath() *ActorPath {
	if p.targetPath == nil {
		p.targetPath, _ = ToActorPath(p.Target)
	}
	return p.targetPath
}

func (p *Message) IsReply() bool {
	return p.ClusterReply != nil
}

func (p *ActorPath) IsChild() bool {
	return p.ChildID != ""
}

func (p *ActorPath) IsParent() bool {
	return p.ChildID == ""
}

// String
func (p *ActorPath) String() string {
	return NewChildPath(p.NodeID, p.ActorID, p.ChildID)
}

func NewActorPath(nodeID, actorID, childID string) *ActorPath {
	return &ActorPath{
		NodeID:  nodeID,
		ActorID: actorID,
		ChildID: childID,
	}
}

func NewChildPath(nodeID, actorID, childID string) string {
	if childID == "" {
		return NewPath(nodeID, actorID)
	}
	return nodeID + cconst.DOT + actorID + cconst.DOT + childID
}

func NewPath(nodeID, actorID string) string {
	return nodeID + cconst.DOT + actorID
}

func ToActorPath(path string) (*ActorPath, error) {
	if path == "" {
		return nil, cerr.ActorPathError
	}

	p := strings.Split(path, cconst.DOT)
	pLen := len(p)

	if pLen == 2 {
		return NewActorPath(p[0], p[1], ""), nil
	}

	if len(p) == 3 {
		return NewActorPath(p[0], p[1], p[2]), nil
	}

	return nil, cerr.ActorPathError
}
