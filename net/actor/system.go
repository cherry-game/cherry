package cherryActor

import (
	ccode "github.com/cherry-game/cherry/code"
	cerror "github.com/cherry-game/cherry/error"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	"strings"
	"sync"
	"time"
)

type (
	// System Actor系统
	System struct {
		mutex            sync.RWMutex
		app              cfacade.IApplication
		actorMap         map[string]*Actor  // key:actorID, value:*actor
		actorOrder       []string           // key:actorID
		localInvokeFunc  cfacade.InvokeFunc // default local func
		remoteInvokeFunc cfacade.InvokeFunc // default remote func
		tellTimeout      time.Duration
	}
)

func NewSystem(app cfacade.IApplication) *System {
	system := &System{
		app:              app,
		mutex:            sync.RWMutex{},
		actorMap:         make(map[string]*Actor, 0),
		actorOrder:       []string{},
		localInvokeFunc:  InvokeLocalFunc,
		remoteInvokeFunc: InvokeRemoteFunc,
		tellTimeout:      3 * time.Second,
	}

	return system
}

func (p *System) NodeId() string {
	if p.app == nil {
		return ""
	}
	return p.app.NodeId()
}

func (p *System) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// reverse order
	for i := len(p.actorOrder) - 1; i >= 0; i-- {
		actorID := p.actorOrder[i]

		cutils.Try(func() {
			thisActor := p.actorMap[actorID]
			thisActor.Exit()
		}, func(err string) {
			clog.Warnf("[OnStop] - [actorID = %s, err = %s]", actorID, err)
		})
	}
}

func (p *System) SetTellTimeout(d time.Duration) {
	p.tellTimeout = d
}

// GetIActor 根据ActorID获取IActor
func (p *System) GetIActor(id string) (cfacade.IActor, bool) {
	return p.GetActor(id)
}

// GetActor 根据ActorID获取*actor
func (p *System) GetActor(id string) (*Actor, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	actorInstance, found := p.actorMap[id]
	return actorInstance, found
}

func (p *System) remoteActor(actorID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.actorMap, actorID)
}

// CreateActor 创建Actor
func (p *System) CreateActor(id string, handler cfacade.IActorHandler) (cfacade.IActor, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrActorIDIsNil
	}

	if actor, found := p.GetIActor(id); found {
		return actor, nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	thisActor, err := newActor(id, "", handler, p)
	if err != nil {
		return nil, err
	}

	p.actorMap[id] = &thisActor             // add to map
	p.actorOrder = append(p.actorOrder, id) // record actor create order
	go thisActor.run()                      // new actor is running!

	return &thisActor, nil
}

// Call 发送远程消息(不回复)
func (p *System) Call(source, target, funcName string, arg interface{}) error {
	if target == "" {
		return cerror.Error("TargetPath is nil.")
	}

	if len(funcName) < 1 {
		return cerror.Errorf("FuncName = %s error.", funcName)
	}

	targetPath, err := cfacade.ToActorPath(target)
	if err != nil {
		return err
	}

	if targetPath.NodeID != "" && targetPath.NodeID != p.NodeId() {
		clusterPacket := cproto.GetClusterPacket()
		clusterPacket.SourcePath = source
		clusterPacket.TargetPath = target
		clusterPacket.FuncName = funcName

		if arg != nil {
			argsBytes, err := p.app.Serializer().Marshal(arg)
			if err != nil {
				clog.Warnf("Marshal error. [targetPath = %s, error = %s]",
					target,
					err,
				)
				return err
			}
			clusterPacket.ArgBytes = argsBytes
		}

		return p.app.Cluster().PublishRemote(targetPath.NodeID, clusterPacket)
	}

	remoteMsg := cfacade.GetMessage()
	remoteMsg.Source = source
	remoteMsg.Target = target
	remoteMsg.FuncName = funcName
	remoteMsg.Args = arg

	if p.PostRemote(remoteMsg) {
		return nil
	}

	clog.Warnf("Call error. [message = %+v]", remoteMsg)

	return ErrActorPathTell
}

// CallWait 发送远程消息(等待回复)
func (p *System) CallWait(source, target, funcName string, arg interface{}, reply interface{}) error {
	if target == "" {
		return cerror.Error("TargetPath is nil.")
	}

	if len(funcName) < 1 {
		return cerror.Errorf("FuncName = %s error.", funcName)
	}

	targetPath, err := cfacade.ToActorPath(target)
	if err != nil {
		return err
	}

	if source == target {
		return cerror.Error("SourcePath can't be the same as targetPath.")
	}

	// forward to remote actor
	if targetPath.NodeID != "" && targetPath.NodeID != p.NodeId() {
		clusterPacket := cproto.GetClusterPacket()
		clusterPacket.SourcePath = source
		clusterPacket.TargetPath = target
		clusterPacket.FuncName = funcName

		if arg != nil {
			argsBytes, err := p.app.Serializer().Marshal(arg)
			if err != nil {
				clog.Warnf("Marshal error. [targetPath = %s, error = %s]",
					target,
					err,
				)
				return err
			}
			clusterPacket.ArgBytes = argsBytes
		}

		rsp := p.app.Cluster().RequestRemote(targetPath.NodeID, clusterPacket, p.tellTimeout)
		if ccode.IsOK(rsp.Code) {
			return p.app.Serializer().Unmarshal(rsp.Data, reply)
		}

		return cerror.Errorf("return code = %d", rsp.Code)

	} else {
		message := cfacade.GetMessage()
		message.Source = source
		message.Target = target
		message.FuncName = funcName
		message.Args = arg

		message.ChanResult = make(chan interface{})
		if !p.PostRemote(message) {
			return cerror.Error("error")
		}

		result := <-message.ChanResult
		if result != nil {
			rsp := result.(*cproto.Response)
			if rsp == nil {
				return cerror.Error("error.")
			}

			if ccode.IsOK(rsp.Code) {
				return p.app.Serializer().Unmarshal(rsp.Data, reply)
			}

			return cerror.Errorf("return code = %d", rsp.Code)
		}
	}

	return nil
}

// PostRemote 提交远程消息
func (p *System) PostRemote(m *cfacade.Message) bool {
	if m == nil {
		clog.Error("Message is nil.")
		return false
	}

	if targetActor, found := p.GetActor(m.TargetPath().ActorID); found {
		targetActor.remoteMail.Push(m)
		return true
	}

	clog.Warnf("[PostRemote] actor not found. [source = %s, target = %s -> %s]",
		m.Source,
		m.Target,
		m.FuncName,
	)
	return false
}

// PostLocal 提交本地消息
func (p *System) PostLocal(m *cfacade.Message) bool {
	if m == nil {
		clog.Error("Message is nil.")
		return false
	}

	if targetActor, found := p.GetActor(m.TargetPath().ActorID); found {
		targetActor.localMail.Push(m)
		return true
	}

	clog.Warnf("[PostLocal] actor not found. [source = %s, target = %s -> %s]",
		m.Source,
		m.Target,
		m.FuncName,
	)

	return false
}

// PostEvent 提交事件
func (p *System) PostEvent(data cfacade.IEventData) {
	if data == nil {
		clog.Error("[PostEvent] Event is nil.")
		return
	}

	for _, thisActor := range p.actorMap {
		thisActor.event.Push(data)
	}
}

func (p *System) SetLocalInvoke(fn cfacade.InvokeFunc) {
	if fn != nil {
		p.localInvokeFunc = fn
	}
}

func (p *System) SetRemoteInvoke(fn cfacade.InvokeFunc) {
	if fn != nil {
		p.remoteInvokeFunc = fn
	}
}
