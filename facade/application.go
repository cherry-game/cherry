package cherryFacade

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type (
	// INode 节点信息
	INode interface {
		NodeID() string        // 节点id(全局唯一)
		NodeType() string      // 节点类型
		Address() string       // 对外网络监听地址(前端节点用)
		RpcAddress() string    // rpc监听地址(未用)
		Settings() ProfileJSON // 节点配置参数
		Enabled() bool         // 是否启用
	}

	IApplication interface {
		INode
		Running() bool                     // 是否运行中
		DieChan() chan bool                // die chan
		IsFrontend() bool                  // 是否为前端节点
		Register(components ...IComponent) // 注册组件
		Find(name string) IComponent       // 根据name获取组件对象
		Remove(name string) IComponent     // 根据name移除组件对象
		All() []IComponent                 // 获取所有组件列表
		OnShutdown(fn ...func())           // 关闭前执行的函数
		Startup()                          // 启动应用实例
		Shutdown()                         // 关闭应用实例
		Serializer() ISerializer           // 序列化
		Discovery() IDiscovery             // 发现服务
		Cluster() ICluster                 // 集群服务
		ActorSystem() IActorSystem         // actor系统
	}

	// ProfileJSON profile配置文件读取接口
	ProfileJSON interface {
		jsoniter.Any
		GetConfig(path ...interface{}) ProfileJSON
		GetString(path interface{}, defaultVal ...string) string
		GetBool(path interface{}, defaultVal ...bool) bool
		GetInt(path interface{}, defaultVal ...int) int
		GetInt32(path interface{}, defaultVal ...int32) int32
		GetInt64(path interface{}, defaultVal ...int64) int64
		GetDuration(path interface{}, defaultVal ...time.Duration) time.Duration
		Unmarshal(ptrVal interface{}) error
	}
)
