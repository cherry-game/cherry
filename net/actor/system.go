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
	// System is the Actor system
	System struct {
		app              cfacade.IApplication // application
		actorMap         *sync.Map            // key:actorID, value:*actor
		actorEventMap    *sync.Map            // map[string]map[string]int64 => key:eventName, value:map[actorPath]uniqueID
		localInvokeFunc  cfacade.InvokeFunc   // default local func
		remoteInvokeFunc cfacade.InvokeFunc   // default remote func
		wg               *sync.WaitGroup      // wait group
		callTimeout      time.Duration        // call timeout
		arrivalTimeOut   int64                // message arrival timeout (ms)
		executionTimeout int64                // message execution timeout (ms)
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
	clog.Info("[OnStop] actor system stopped!")
}

// GetIActor returns IActor by actor ID
func (p *System) GetIActor(id string) (cfacade.IActor, bool) {
	return p.GetActor(id)
}

// GetActor returns *Actor by actor ID
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

// CreateActor creates a new Actor
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

// Call sends a remote message (no reply)
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
		remoteMsg, errCode := p.buildClusterMessage(source, target, funcName, arg)
		if ccode.IsFail(errCode) {
			clog.Warnf("[Call] Marshal arg error. [targetPath = %s, error = %d]", target, errCode)
			return errCode
		}

		// PublishRemote recycles remoteMsg via defer on all paths.
		err := p.app.Cluster().PublishRemote(targetPath.NodeID, remoteMsg)
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

		if !p.PostRemote(remoteMsg) {
			clog.Warnf("[Call] Post remote fail. [source = %s, target = %s, funcName = %s]", source, target, funcName)
			return ccode.ActorInvokeRemoteError
		}
	}

	return ccode.OK
}

// CallWait sends a remote message and waits for reply
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

	// Forward to remote node.
	if targetPath.NodeID != "" && targetPath.NodeID != sourcePath.NodeID {
		remoteMsg, errCode := p.buildClusterMessage(source, target, funcName, arg)
		if ccode.IsFail(errCode) {
			clog.Warnf("[CallWait] Marshal arg error. [targetPath = %s, error = %d]", target, errCode)
			return errCode
		}

		// RequestRemote recycles remoteMsg via defer on all paths.
		rspData, rspCode := p.app.Cluster().RequestRemote(targetPath.NodeID, remoteMsg, p.callTimeout)
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

		if sourcePath.ActorID == targetPath.ActorID {
			childActor, found := p.GetChildActor(targetPath.ActorID, targetPath.ChildID)
			if !found {
				message.Recycle()
				return ccode.ActorChildIDNotFound
			}
			childActor.PostRemote(message)
		} else {
			if !p.PostRemote(message) {
				clog.Warnf("[CallWait] Post remote fail. [source = %s, target = %s, funcName = %s]", source, target, funcName)
				return ccode.ActorInvokeRemoteError
			}
		}

		var result interface{}

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

// CallType publishes message by node type
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

	argsBytes, errCode := p.marshalArg(arg)
	if ccode.IsFail(errCode) {
		clog.Warnf("[CallType] Marshal arg error. [nodeType = %s, actorID = %s, funcName = %s, error = %d]",
			nodeType,
			actorID,
			funcName,
			errCode,
		)
		return errCode
	}

	remoteMsg := cfacade.GetMessage()
	remoteMsg.Target = cfacade.NewPath("", actorID)
	remoteMsg.FuncName = funcName
	remoteMsg.Args = argsBytes

	// PublishRemoteType recycles remoteMsg via defer on all paths.
	err := p.app.Cluster().PublishRemoteType(nodeType, remoteMsg)
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

// PostRemote delivers message to the remote mailbox.
func (p *System) PostRemote(m *cfacade.Message) bool {
	if m == nil {
		clog.Error("Message is nil.")
		return false
	}

	targetActor, found := p.GetActor(m.TargetPath().ActorID)
	if !found {
		clog.Warnf("[PostRemote] actor not found. [source = %s, target = %s -> %s]", m.Source, m.Target, m.FuncName)
		m.Recycle()
		return false
	}

	if targetActor.state != WorkerState {
		m.Recycle()
		return false
	}

	targetActor.PostRemote(m)
	return true
}

// PostLocal delivers message to the local mailbox.
func (p *System) PostLocal(m *cfacade.Message) bool {
	if m == nil {
		clog.Error("Message is nil.")
		return false
	}

	targetActor, found := p.GetActor(m.TargetPath().ActorID)
	if !found {
		clog.Warnf("[PostLocal] actor not found. [source = %s, target = %s -> %s]", m.Source, m.Target, m.FuncName)
		m.Recycle()
		return false
	}

	if targetActor.state != WorkerState {
		m.Recycle()
		return false
	}

	targetActor.PostLocal(m)
	return true
}

// PostEvent delivers an event to subscribed actors
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

		if targetActor.state != WorkerState {
			return true
		}

		// no set unique
		if value == nil {
			targetActor.event.Push(data)
			return true
		}

		uniqueID, ok := value.(int64)
		if !ok {
			clog.Warnf("[PostEvent] UniqueID set error in actorEventMap. value = %v", value)
			return true
		}

		if uniqueID == data.UniqueID() {
			targetActor.event.Push(data)
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

// marshalArg serializes the argument for cross-node transfer.
func (p *System) marshalArg(arg any) ([]byte, int32) {
	if arg == nil {
		return nil, ccode.OK
	}
	bytes, err := p.app.Serializer().Marshal(arg)
	if err != nil {
		return nil, ccode.ActorMarshalError
	}
	return bytes, ccode.OK
}

// buildClusterMessage creates a Message for cross-node transfer.
// Args are serialized to []byte. The caller is responsible for recycling
// (PublishRemote / RequestRemote recycle via defer on all paths).
func (p *System) buildClusterMessage(source, target, funcName string, arg any) (*cfacade.Message, int32) {
	argsBytes, errCode := p.marshalArg(arg)
	if ccode.IsFail(errCode) {
		return nil, errCode
	}

	msg := cfacade.GetMessage()
	msg.Source = source
	msg.Target = target
	msg.FuncName = funcName
	msg.Args = argsBytes
	return msg, ccode.OK
}
