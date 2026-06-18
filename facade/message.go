// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines the in-process message carrier and Actor path helpers:
//   - Message: pooled, reference-counted envelope for Actor dispatch
//   - ActorPath: parsed {NodeID, ActorID, ChildID} triplet
package cherryFacade

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	cconst "github.com/cherry-game/cherry/const"
	cerr "github.com/cherry-game/cherry/error"
	cstring "github.com/cherry-game/cherry/extend/string"
	cproto "github.com/cherry-game/cherry/net/proto"
	"google.golang.org/protobuf/proto"
)

var (
	messagePool = &sync.Pool{
		New: func() interface{} {
			return new(Message)
		},
	}
)

type (
	// Message is the in-process message carrier for the Actor system.
	// For cross-process transfer, use Marshal/Unmarshal which internally uses ClusterPacket proto.
	//
	// Field groups:
	//   Common: BuildTime, Source, Target, FuncName, Args
	//   Local (client->Actor, set by parser): Session
	//   Remote (Actor->Actor, set by System.Call/CallWait/CallType): ReqID, Reply, ChanResult
	Message struct {
		// --- Common fields ---
		refs      int32       // reference count, see Recycle()
		BuildTime int64       // message build time(ms)
		Source    string      // source actor path
		Target    string      // target actor path (node.actor or node.actor.child)
		FuncName  string      // target function name
		Args      interface{} // payload: same-node=decoded object, cross-node=[]byte (pending decode)

		// --- Local only (client->Actor, set by parser) ---
		Session *cproto.Session // client session

		// --- Remote only (Actor->Actor, set by System.Call/CallWait/CallType) ---
		ReqID      string           // NATS request ID (cross-node request-reply)
		Reply      string           // NATS reply subject (non-empty if from cross-node)
		ChanResult chan interface{} // same-node CallWait sync channel

		targetPath *ActorPath // lazily cached on first TargetPath() call; cleared in Recycle()
	}

	ActorPath struct {
		NodeID  string
		ActorID string
		ChildID string
	}
)

// GetMessage returns a pooled Message with BuildTime set to now.
// Caller must Recycle() when done (or the receiving Actor will do so).
func GetMessage() *Message {
	msg := messagePool.Get().(*Message)
	msg.BuildTime = time.Now().UnixMilli()
	return msg
}

// Marshal serializes Message for cross-process transfer via NATS.
// Internally uses ClusterPacket proto as the wire format.
func (p *Message) Marshal() ([]byte, error) {
	cp := cproto.GetClusterPacket()
	defer cp.Recycle()

	cp.BuildTime = p.BuildTime
	cp.SourcePath = p.Source
	cp.TargetPath = p.Target
	cp.FuncName = p.FuncName
	cp.Session = p.Session

	if argBytes, ok := p.Args.([]byte); ok {
		cp.ArgBytes = argBytes
	}

	return proto.Marshal(cp)
}

// Unmarshal deserializes cross-process transfer data back into Message.
func (p *Message) Unmarshal(data []byte) error {
	cp := cproto.GetClusterPacket()
	defer cp.Recycle()

	if err := proto.Unmarshal(data, cp); err != nil {
		return err
	}

	p.BuildTime = cp.BuildTime
	p.Source = cp.SourcePath
	p.Target = cp.TargetPath
	p.FuncName = cp.FuncName
	p.Args = cp.ArgBytes // keep as []byte, decoded on-demand by Actor
	p.Session = cp.Session

	return nil
}

// Clone creates a shallow copy for child Actor forwarding.
// Session, Args and ChanResult are intentionally shared.
func (p *Message) Clone() *Message {
	clone := GetMessage()
	clone.BuildTime = p.BuildTime
	clone.Source = p.Source
	clone.Target = p.Target
	clone.FuncName = p.FuncName
	clone.Args = p.Args       // shared (same-node=object, cross-node=read-only []byte)
	clone.Session = p.Session // shared (Session must persist across forwarding)
	clone.ReqID = p.ReqID
	clone.Reply = p.Reply
	clone.ChanResult = p.ChanResult // shared (CallWait sync channel)
	return clone
}

// AddRef increments the reference count. Each call to PostLocal/PostRemote
// adds a reference that will be released by the receiving actor's Recycle.
func (p *Message) AddRef() {
	atomic.AddInt32(&p.refs, 1)
}

// Recycle decrements the reference count and returns the Message to the pool
// when it reaches zero. Each PostLocal/PostRemote increments refs; the receiving
// Actor calls Recycle after processing. If delivery fails, the caller must Recycle.
func (p *Message) Recycle() {
	if atomic.AddInt32(&p.refs, -1) > 0 {
		return
	}

	p.refs = 0
	p.BuildTime = 0
	p.Source = ""
	p.Target = ""
	p.FuncName = ""
	p.Session = nil
	p.Args = nil
	p.ReqID = ""
	p.Reply = ""
	p.ChanResult = nil
	p.targetPath = nil
	messagePool.Put(p)
}

// TargetPath lazily parses the Target field into an ActorPath.
// Result is cached across the message lifetime and cleared on Recycle.
func (p *Message) TargetPath() *ActorPath {
	if p.targetPath == nil {
		p.targetPath, _ = ToActorPath(p.Target)
	}
	return p.targetPath
}

// IsChild returns true if this path targets a child Actor.
func (p *ActorPath) IsChild() bool {
	return p.ChildID != ""
}

// IsParent returns true if this path targets a parent (non-child) Actor.
func (p *ActorPath) IsParent() bool {
	return p.ChildID == ""
}

// String reconstructs the dotted path notation: "nodeID.actorID" or "nodeID.actorID.childID".
func (p *ActorPath) String() string {
	return NewChildPath(p.NodeID, p.ActorID, p.ChildID)
}

// NewActorPath creates an ActorPath from individual components.
// Pass empty string for childID when targeting a parent Actor.
func NewActorPath(nodeID, actorID, childID string) *ActorPath {
	return &ActorPath{
		NodeID:  nodeID,
		ActorID: actorID,
		ChildID: childID,
	}
}

// NewChildPath builds a dotted path string. If childID is empty,
// it returns "nodeID.actorID"; otherwise "nodeID.actorID.childID".
func NewChildPath(nodeID, actorID, childID interface{}) string {
	if childID == "" {
		return NewPath(nodeID, actorID)
	}
	return cstring.ToString(nodeID) + cconst.DOT + cstring.ToString(actorID) + cconst.DOT + cstring.ToString(childID)
}

// NewPath builds a two-segment dotted path "nodeID.actorID".
func NewPath(nodeID, actorID interface{}) string {
	return cstring.ToString(nodeID) + cconst.DOT + cstring.ToString(actorID)
}

// ToActorPath parses a dotted path string into an ActorPath.
// Accepts "node.actor" (2-segment) or "node.actor.child" (3-segment) formats.
func ToActorPath(path string) (*ActorPath, error) {
	if path == "" {
		return nil, cerr.ActorPathError
	}

	p := strings.Split(path, cconst.DOT)
	pLen := len(p)

	if pLen == 2 {
		return NewActorPath(p[0], p[1], ""), nil
	}

	if pLen == 3 {
		return NewActorPath(p[0], p[1], p[2]), nil
	}

	return nil, cerr.ActorPathError
}
