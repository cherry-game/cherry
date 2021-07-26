package cherryFacade

import "reflect"

// IHandler 消息处理句柄接口，用于包装消息处理逻辑
type IHandler interface {
	IAppContext                                       // 应用实例上线文
	Name() string                                     // handler名称(用于消息路由)
	SetName(name string)                              // 设置handler名称
	OnPreInit()                                       // 预初始化方法(对象实例化前)
	OnInit()                                          // 初始方法(PreInit之后)
	OnAfterInit()                                     // 最后的初始化方法(Init之后)
	OnStop()                                          // 停止handler运行
	Events() map[string][]EventFunc                   // 已注册的事件列表
	Event(name string) ([]EventFunc, bool)            // 根据事件名获取事件列表
	LocalHandlers() map[string]*HandlerFn             // 已注册的本地handler列表(网络消息的逻辑处理函数)
	LocalHandler(funcName string) (*HandlerFn, bool)  // 根据handler名称获取本地handler
	RemoteHandlers() map[string]*HandlerFn            // 已注册的远程handler列表(内部rpc调用的逻辑处理函数)
	RemoteHandler(funcName string) (*HandlerFn, bool) // 根据handler名称获取远程handler
}

// HandlerFn 函数反射信息
type HandlerFn struct {
	Type    reflect.Type
	Value   reflect.Value
	InArgs  []reflect.Type
	OutArgs []reflect.Type
}
