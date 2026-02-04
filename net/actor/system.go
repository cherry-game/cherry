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
		actorEventMap    *sync.Map          // map[string]map[string]int64 => key:eventName, value:map[actorPath]uniqueID
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
		actorEventMap:    &sync.Map{},
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

func (p *System) NodeID() string {
	if p.app == nil {
		return ""
	}

	return p.app.NodeID()
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

	clog.Info("[OnStop] actor system stopping!")
	p.wg.Wait()
	clog.Info("[OnStop]actor system stopped!")
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

func (p *System) GetActorWithPath(path string) (*Actor, bool) {
	actorPath, err := cfacade.ToActorPath(path)
	if err != nil {
		clog.Warnf("[GetActorWithPath] Actor path is error. path = %s, err = %v", path, err)
		return nil, false
	}

	if actorPath.IsChild() {
		return p.GetChildActor(actorPath.ActorID, actorPath.ChildID)
	}

	return p.GetActor(actorPath.ActorID)
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
func (p *System) Call(source, target, funcName string, arg any) int32 {
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

	if targetPath.NodeID != "" && targetPath.NodeID != p.NodeID() {
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
			return ccode.ActorInvokeRemoteError
		}
	}

	return ccode.OK
}

// CallWait 发送远程消息(等待回复)
func (p *System) CallWait(source, target, funcName string, arg, reply any) int32 {
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

		rspData, rspCode := p.app.Cluster().RequestRemote(targetPath.NodeID, clusterPacket, p.callTimeout)
		if ccode.IsFail(rspCode) {
			return rspCode
		}

		if reply != nil {
			if err = p.app.Serializer().Unmarshal(rspData, reply); err != nil {
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
		} else {
			if !p.PostRemote(&message) {
				clog.Warnf("[CallWait] Post remote fail. [source = %s, target = %s, funcName = %s]",
					source,
					target,
					funcName,
				)
				return ccode.ActorInvokeRemoteError
			}
		}

		select {
		case result = <-message.ChanResult:
			{
				if result == nil {
					clog.Warnf("[CallWait] Response is nil. [source = %s, target = %s, funcName = %s]",
						source,
						target,
						funcName,
					)
					return ccode.ActorInvokeResultIsNil
				}

				rsp := result.(*cproto.Response)
				if rsp == nil {
					clog.Warnf("[CallWait] Response is nil. [source = %s, target = %s, funcName = %s]",
						source,
						target,
						funcName,
					)
					return ccode.ActorResponseIsError
				}

				if ccode.IsFail(rsp.Code) {
					return rsp.Code
				}

				if reply != nil {
					if rsp.Data == nil {
						clog.Warnf("[CallWait] rsp.Data is nil.[source = %s, target = %s, funcName = %s, error = %s]",
							source,
							target,
							funcName,
							err,
						)
					}

					err = p.app.Serializer().Unmarshal(rsp.Data, reply)
					if err != nil {
						clog.Warnf("[CallWait] Unmarshal reply error.[source = %s, target = %s, funcName = %s, error = %s]",
							source,
							target,
							funcName,
							err,
						)
						return ccode.ActorUnmarshalError
					}
				}
			}
		case <-time.After(p.callTimeout):
			return ccode.ActorCallTimeout
		}
	}

	return ccode.OK
}

// Broadcast 根据节点类型发布消息
func (p *System) CallType(nodeType, actorID, funcName string, arg any) int32 {
	if actorID == "" {
		return ccode.ActorIDIsNil
	}

	if len(funcName) < 1 {
		clog.Warnf("[CallType] FuncName error. [nodeType = %s, actorID = %s, funcName = %s]",
			nodeType,
			actorID,
			funcName,
		)
		return ccode.ActorFuncNameError
	}

	clusterPacket := cproto.GetClusterPacket()
	clusterPacket.TargetPath = cfacade.NewPath("", actorID)
	clusterPacket.FuncName = funcName

	if arg != nil {
		argsBytes, err := p.app.Serializer().Marshal(arg)
		if err != nil {
			clog.Warnf("[CallType] Marshal arg error. [nodeType = %s, actorID = %s, funcName = %s, error = %s]",
				nodeType,
				actorID,
				funcName,
				err,
			)
			return ccode.ActorMarshalError
		}
		clusterPacket.ArgBytes = argsBytes
	}

	err := p.app.Cluster().PublishRemoteType(nodeType, clusterPacket)
	if err != nil {
		clog.Warnf("[CallType] Publish remote fail. [nodeType = %s, actorID = %s, funcName = %s, error = %v]",
			nodeType,
			actorID,
			funcName,
			err,
		)
		return ccode.ActorPublishRemoteError
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

	clog.Warnf("[PostRemote] actor not found. [source = %s, target = %s -> %s]", m.Source, m.Target, m.FuncName)
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

	clog.Warnf("[PostLocal] actor not found. [source = %s, target = %s -> %s]", m.Source, m.Target, m.FuncName)

	return false
}

// PostEvent 提交事件
func (p *System) PostEvent(data cfacade.IEventData) {
	if data == nil {
		clog.Error("[PostEvent] Event is nil.")
		return
	}

	if len(data.Name()) < 1 {
		clog.Warnf("[PostEvent] Event name is empty. value = %v", data)
		return
	}

	valueMap, found := p.actorEventMap.Load(data.Name())
	if !found {
		return
	}

	// map[string]int64
	actorIDSMap, ok := valueMap.(*sync.Map)
	if !ok {
		return
	}

	actorIDSMap.Range(func(key, value any) bool {
		path := key.(string)
		targetActor, found := p.GetActorWithPath(path)
		if !found {
			return true
		}

		// no set unique
		if value == nil {
			if targetActor.state == WorkerState {
				targetActor.event.Push(data)
			}

			return true
		}

		uniqueID, ok := value.(int64)
		if !ok {
			clog.Warnf("[PostEvent] UniqueID set error in actorEventMap. value = %v", value)
			return true
		}

		if uniqueID == data.UniqueID() {
			if targetActor.state == WorkerState {
				targetActor.event.Push(data)
			}

			return true
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

func (p *System) addActorEvent(actorPath string, eventName string, uniqueID ...int64) {
	// map[string]map[string]int64 => key:eventName, value:map[actorPath]uniqueID
	value, _ := p.actorEventMap.LoadOrStore(eventName, &sync.Map{})
	eventMap := value.(*sync.Map)

	if len(uniqueID) > 0 {
		eventMap.Store(actorPath, uniqueID[0])
	} else {
		eventMap.Store(actorPath, nil) // no set unique
	}
}

func (p *System) removeActorEvent(actorPath string, eventNames ...string) {
	for _, eventName := range eventNames {
		value, found := p.actorEventMap.Load(eventName)
		if !found {
			continue
		}

		if actorIDMap, found := value.(*sync.Map); found {
			actorIDMap.Delete(actorPath)
		}
	}
}
