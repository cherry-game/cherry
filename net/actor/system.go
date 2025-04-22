package cherryActor

import (
	"strings"
	"sync"
	"time"

	ccode "github.com/cherry-game/cherry/code"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

type (
	// System Actor系统
	System struct {
		app              cfacade.IApplication
		actorMap         *sync.Map          // key:actorID, value:*actor
		localInvokeFunc  cfacade.InvokeFunc // default local func
		remoteInvokeFunc cfacade.InvokeFunc // default remote func
		wg               *sync.WaitGroup    // wait group
		callTimeout      time.Duration      // call调用超时
		arrivalTimeOut   int64              // message到达超时(毫秒)
		executionTimeout int64              // 消息执行超时(毫秒)
	}
)

func NewSystem() *System {
	system := &System{
		actorMap:         &sync.Map{},
		localInvokeFunc:  InvokeLocalFunc,
		remoteInvokeFunc: InvokeRemoteFunc,
		wg:               &sync.WaitGroup{},
		callTimeout:      3 * time.Second,
		arrivalTimeOut:   100,
		executionTimeout: 100,
	}

	return system
}

func (p *System) SetApp(app cfacade.IApplication) {
	p.app = app
}

func (p *System) NodeId() string {
	if p.app == nil {
		return ""
	}

	return p.app.NodeId()
}

func (p *System) Stop() {
	p.actorMap.Range(func(key, value any) bool {
		actor, ok := value.(*Actor)
		if ok {
			cutils.Try(func() {
				actor.Exit()
			}, func(err string) {
				clog.Warnf("[OnStop] - [actorID = %s, err = %s]", actor.path, err)
			})
		}
		return true
	})

	clog.Info("actor system stopping!")
	p.wg.Wait()
	clog.Info("actor system stopped!")
}

// GetIActor 根据ActorID获取IActor
func (p *System) GetIActor(id string) (cfacade.IActor, bool) {
	return p.GetActor(id)
}

// GetActor 根据ActorID获取*actor
func (p *System) GetActor(id string) (*Actor, bool) {
	actorValue, found := p.actorMap.Load(id)
	if !found {
		return nil, false
	}

	actor, found := actorValue.(*Actor)
	return actor, found
}

func (p *System) GetChildActor(actorID, childID string) (*Actor, bool) {
	parentActor, found := p.GetActor(actorID)
	if !found {
		return nil, found
	}

	return parentActor.child.GetActor(childID)
}

func (p *System) removeActor(actorID string) {
	p.actorMap.Delete(actorID)
}

// CreateActor 创建Actor
func (p *System) CreateActor(id string, handler cfacade.IActorHandler) (cfacade.IActor, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrActorIDIsNil
	}

	if actor, found := p.GetIActor(id); found {
		return actor, nil
	}

	thisActor, err := newActor(id, "", handler, p)
	if err != nil {
		return nil, err
	}

	p.actorMap.Store(id, thisActor) // add to map
	go thisActor.run()              // new actor is running!

	return thisActor, nil
}

// Call 发送远程消息(不回复)
func (p *System) Call(source, target, funcName string, arg interface{}) int32 {
	if target == "" {
		clog.Warnf("[Call] Target path is nil. [source = %s, target = %s, funcName = %s]",
			source,
			target,
			funcName,
		)
		return ccode.ActorPathIsNil
	}

	if len(funcName) < 1 {
		clog.Warnf("[Call] FuncName error. [source = %s, target = %s, funcName = %s]",
			source,
			target,
			funcName,
		)
		return ccode.ActorFuncNameError
	}

	targetPath, err := cfacade.ToActorPath(target)
	if err != nil {
		clog.Warnf("[Call] Target path error. [source = %s, target = %s, funcName = %s, err = %v]",
			source,
			target,
			funcName,
			err,
		)
		return ccode.ActorConvertPathError
	}

	if targetPath.NodeID != "" && targetPath.NodeID != p.NodeId() {
		clusterPacket := cproto.GetClusterPacket()
		clusterPacket.SourcePath = source
		clusterPacket.TargetPath = target
		clusterPacket.FuncName = funcName

		if arg != nil {
			argsBytes, err := p.app.Serializer().Marshal(arg)
			if err != nil {
				clog.Warnf("[Call] Marshal arg error. [targetPath = %s, error = %s]",
					target,
					err,
				)
				return ccode.ActorMarshalError
			}
			clusterPacket.ArgBytes = argsBytes
		}

		err = p.app.Cluster().PublishRemote(targetPath.NodeID, clusterPacket)
		if err != nil {
			clog.Warnf("[Call] Publish remote fail. [source = %s, target = %s, funcName = %s, err = %v]",
				source,
				target,
				funcName,
				err,
			)
			return ccode.ActorPublishRemoteError
		}
	} else {
		remoteMsg := cfacade.GetMessage()
		remoteMsg.Source = source
		remoteMsg.Target = target
		remoteMsg.FuncName = funcName
		remoteMsg.Args = arg

		if !p.PostRemote(&remoteMsg) {
			clog.Warnf("[Call] Post remote fail. [source = %s, target = %s, funcName = %s]", source, target, funcName)
			return ccode.ActorCallFail
		}
	}

	return ccode.OK
}

// CallWait 发送远程消息(等待回复)
func (p *System) CallWait(source, target, funcName string, arg interface{}, reply interface{}) int32 {
	sourcePath, err := cfacade.ToActorPath(source)
	if err != nil {
		clog.Warnf("[CallWait] Source path error. [source = %s, target = %s, funcName = %s, err = %v]",
			source,
			target,
			funcName,
			err,
		)
		return ccode.ActorConvertPathError
	}

	targetPath, err := cfacade.ToActorPath(target)
	if err != nil {
		clog.Warnf("[CallWait] Target path error. [source = %s, target = %s, funcName = %s, err = %v]",
			source,
			target,
			funcName,
			err,
		)
		return ccode.ActorConvertPathError
	}

	if source == target {
		clog.Warnf("[CallWait] Source path is equal target. [source = %s, target = %s, funcName = %s]",
			source,
			target,
			funcName,
		)
		return ccode.ActorSourceEqualTarget
	}

	if len(funcName) < 1 {
		clog.Warnf("[CallWait] FuncName error. [source = %s, target = %s, funcName = %s]",
			source,
			target,
			funcName,
		)
		return ccode.ActorFuncNameError
	}

	// forward to remote actor
	if targetPath.NodeID != "" && targetPath.NodeID != sourcePath.NodeID {
		clusterPacket := cproto.BuildClusterPacket(source, target, funcName)

		if arg != nil {
			argsBytes, err := p.app.Serializer().Marshal(arg)
			if err != nil {
				clog.Warnf("[CallWait] Marshal arg error. [targetPath = %s, error = %s]", target, err)
				return ccode.ActorMarshalError
			}
			clusterPacket.ArgBytes = argsBytes
		}

		rsp := p.app.Cluster().RequestRemote(targetPath.NodeID, clusterPacket, p.callTimeout)
		if ccode.IsFail(rsp.Code) {
			return rsp.Code
		}

		if reply != nil {
			if err = p.app.Serializer().Unmarshal(rsp.Data, reply); err != nil {
				clog.Warnf("[CallWait] Marshal reply error. [targetPath = %s, error = %s]", target, err)
				return ccode.ActorMarshalError
			}
		}

	} else {
		message := cfacade.GetMessage()
		message.Source = source
		message.Target = target
		message.FuncName = funcName
		message.Args = arg
		message.ChanResult = make(chan interface{})

		var result interface{}

		if sourcePath.ActorID == targetPath.ActorID {
			if sourcePath.ChildID == targetPath.ChildID {
				return ccode.ActorSourceEqualTarget
			}

			childActor, found := p.GetChildActor(targetPath.ActorID, targetPath.ChildID)
			if !found {
				return ccode.ActorChildIDNotFound
			}

			childActor.PostRemote(&message)
			result = <-message.ChanResult
		} else {
			if !p.PostRemote(&message) {
				clog.Warnf("[CallWait] Post remote fail. [source = %s, target = %s, funcName = %s]", source, target, funcName)
				return ccode.ActorCallFail
			}
			result = <-message.ChanResult
		}

		if result != nil {
			rsp := result.(*cproto.Response)
			if rsp == nil {
				clog.Warnf("[CallWait] Response is nil. [targetPath = %s]",
					target,
				)
				return ccode.ActorCallFail
			}

			if ccode.IsFail(rsp.Code) {
				return rsp.Code
			}

			if reply != nil {
				if rsp.Data == nil {
					clog.Warnf("[CallWait] rsp.Data is nil. [targetPath = %s, error = %s]",
						target,
						err,
					)
				}

				err = p.app.Serializer().Unmarshal(rsp.Data, reply)
				if err != nil {
					clog.Warnf("[CallWait] Unmarshal reply error. [targetPath = %s, error = %s]",
						target,
						err,
					)
					return ccode.ActorUnmarshalError
				}
			}
		}
	}

	return ccode.OK
}

// PostRemote 提交远程消息
func (p *System) PostRemote(m *cfacade.Message) bool {
	if m == nil {
		clog.Error("Message is nil.")
		return false
	}

	if targetActor, found := p.GetActor(m.TargetPath().ActorID); found {
		if targetActor.state == WorkerState {
			targetActor.PostRemote(m)
		}
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
		if targetActor.state == WorkerState {
			targetActor.PostLocal(m)
		}
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

	p.actorMap.Range(func(key, value any) bool {
		if thisActor, found := value.(*Actor); found {
			if thisActor.state == WorkerState {
				thisActor.event.Push(data)
			}
		}
		return true
	})
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

func (p *System) SetCallTimeout(d time.Duration) {
	p.callTimeout = d
}

func (p *System) SetArrivalTimeout(t int64) {
	if t > 1 {
		p.arrivalTimeOut = t
	}
}

func (p *System) SetExecutionTimeout(t int64) {
	if t > 1 {
		p.executionTimeout = t
	}
}
