package cherryFacade

import (
	"time"

	creflect "github.com/cherry-game/cherry/extend/reflect"
)

type (
	IActorSystem interface {
		GetIActor(id string) (IActor, bool)
		CreateActor(id string, handler IActorHandler) (IActor, error)
		PostRemote(m *Message) bool
		PostLocal(m *Message) bool
		PostEvent(data IEventData)
		Call(source, target, funcName string, arg interface{}) int32
		CallWait(source, target, funcName string, arg interface{}, reply interface{}) int32
		SetLocalInvoke(invoke InvokeFunc)
		SetRemoteInvoke(invoke InvokeFunc)
		SetCallTimeout(d time.Duration)
		SetArrivalTimeout(t int64)
		SetExecutionTimeout(t int64)
	}

	InvokeFunc func(app IApplication, fi *creflect.FuncInfo, m *Message)

	IActor interface {
		App() IApplication
		ActorID() string
		Path() *ActorPath
		Call(targetPath, funcName string, arg interface{}) int32
		CallWait(targetPath, funcName string, arg interface{}, reply interface{}) int32
		PostRemote(m *Message)
		PostLocal(m *Message)
		LastAt() int64
		Exit()
	}

	IActorHandler interface {
		AliasID() string                          // actorID
		OnInit()                                  // 当Actor启动前触发该函数
		OnStop()                                  // 当Actor停止前触发该函数
		OnLocalReceived(m *Message) (bool, bool)  // 当Actor接收local消息时触发该函数
		OnRemoteReceived(m *Message) (bool, bool) // 当Actor接收remote消息时执行的函数
		OnFindChild(m *Message) (IActor, bool)    // 当actor查找子Actor时触发该函数
	}

	IActorChild interface {
		Create(id string, handler IActorHandler) (IActor, error)                        // 创建子Actor
		Get(id string) (IActor, bool)                                                   // 获取子Actor
		Remove(id string)                                                               // 称除子Actor
		Each(fn func(i IActor))                                                         // 遍历所有子Actor
		Call(childID, funcName string, arg interface{})                                 // 调用当前子actor的函数
		CallWait(targetPath, funcName string, arg interface{}, reply interface{}) int32 // 调用当前子actor的函数并等待返回
	}
)

type (
	IEventData interface {
		Name() string    // 事件名
		UniqueID() int64 // 唯一id
	}
)
