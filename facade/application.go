package cherryFacade

import (
	jsoniter "github.com/json-iterator/go"
	"time"
)

type (

	// INode 节点信息
	INode interface {
		NodeId() string       // 节点id(全局唯一)
		NodeType() string     // 节点类型
		Address() string      // 对外网络监听地址
		RpcAddress() string   // rpc监听地址
		Settings() JsonConfig // 节点配置参数
		Enabled() bool        // 是否启用
	}

	IApplication interface {
		INode
		ISerializer
		IPacketCodec
		Running() bool                    // 是否运行中
		DieChan() chan bool               // die chan
		IsFrontend() bool                 // 是否为前端节点
		Find(name string) IComponent      // 根据name获取组件对象
		Remove(name string) IComponent    // 根据name移除组件对象
		All() []IComponent                // 获取所有组件列表
		OnShutdown(fn ...func())          // 关闭前执行的函数
		Startup(components ...IComponent) // 启动应用实例
		Shutdown()                        // 关闭应用实例
	}

	// IAppContext App上下文
	IAppContext interface {
		Set(app IApplication)
		App() IApplication
	}

	// AppContext 继承自IApplication实现默认的方法
	AppContext struct {
		IApplication
	}

	JsonConfig interface {
		jsoniter.Any
		GetConfig(path ...interface{}) JsonConfig
		GetString(path interface{}, defaultVal ...string) string
		GetBool(path interface{}, defaultVal ...bool) bool
		GetInt(path interface{}, defaultVal ...int) int
		GetInt32(path interface{}, defaultVal ...int32) int32
		GetInt64(path interface{}, defaultVal ...int64) int64
		GetDuration(path interface{}, defaultVal ...int64) time.Duration
		MarshalWithPath(path interface{}, ptrVal interface{}) error
		Marshal(ptrVal interface{}) error
	}
)

func (b *AppContext) Set(app IApplication) {
	b.IApplication = app
}

func (b *AppContext) App() IApplication {
	return b.IApplication
}
