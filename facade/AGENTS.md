# facade - 核心接口定义

## 概述

定义框架核心接口合约，所有组件实现必须遵循这些接口。框架仅提供基础实现，开发者可基于接口进行扩展。

## 目录结构

```
facade/
├── application.go  # IApplication, INode, ProfileJSON
├── actor.go        # IActorSystem, IActor, IActorHandler, IActorChild, IEventData
├── cluster.go      # ICluster, IDiscovery, IMember 接口
├── component.go    # IComponent, IComponentLifecycle
├── session.go      # SID, UID 类型定义
├── serializer.go   # ISerializer 接口（可扩展）
├── net_parser.go   # INetParser 接口（可扩展）
├── connector.go    # IConnector 接口（可扩展）
└── message.go      # Message, ActorPath 结构体
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 节点信息 | `application.go` | `INode` 接口 |
| 应用接口 | `application.go` | `IApplication` 接口 |
| Actor 系统 | `actor.go` | `IActorSystem` 接口 |
| Actor 实例 | `actor.go:28-39` | `IActor` 接口 |
| Actor 处理器 | `actor.go:41-48` | `IActorHandler` 接口 |
| 子 Actor | `actor.go:50-57` | `IActorChild` 接口 |
| 事件数据 | `actor.go:60-64` | `IEventData` 接口 |
| 集群 RPC | `cluster.go` | `ICluster` 接口 |
| 节点发现 | `cluster.go` | `IDiscovery` 接口 |
| 成员信息 | `cluster.go` | `IMember` 接口 |
| 组件合约 | `component.go` | `IComponent`, `IComponentLifecycle` |
| 序列化器 | `serializer.go` | `ISerializer` 接口（可扩展） |
| 协议解析器 | `net_parser.go` | `INetParser` 接口（可扩展） |
| 连接器 | `connector.go` | `IConnector` 接口（可扩展） |
| 消息结构 | `message.go` | `Message`, `ActorPath` |

## 核心接口

### 应用层接口

**INode** - 节点信息:
```go
type INode interface {
    NodeID() string        // 节点 ID（全局唯一）
    NodeType() string      // 节点类型
    Address() string       // 对外监听地址
    RpcAddress() string    // RPC 地址
    Settings() ProfileJSON // 节点配置
    Enabled() bool         // 是否启用
}
```

**IApplication** - 应用主接口:
```go
type IApplication interface {
    INode
    Running() bool
    Register(components ...IComponent)
    Find(name string) IComponent
    Startup()
    Shutdown()
    Serializer() ISerializer
    Discovery() IDiscovery
    Cluster() ICluster
    ActorSystem() IActorSystem
}
```

### Actor 系统接口

**IActorSystem** - Actor 系统管理:
```go
type IActorSystem interface {
    GetIActor(id string) (IActor, bool)
    CreateActor(id string, handler IActorHandler) (IActor, error)
    PostRemote(m *Message) bool
    PostLocal(m *Message) bool
    PostEvent(data IEventData)
    Call(source, target, funcName string, arg any) int32
    CallWait(source, target, funcName string, arg, reply any) int32
    CallType(nodeType, actorID, funcName string, arg any) int32
    SetLocalInvoke(invoke InvokeFunc)
    SetRemoteInvoke(invoke InvokeFunc)
    SetCallTimeout(d time.Duration)
    SetArrivalTimeout(t int64)
    SetExecutionTimeout(t int64)
}
```

**IActor** - Actor 实例接口（可扩展）:
```go
type IActor interface {
    App() IApplication
    ActorID() string
    Path() *ActorPath
    Call(targetPath, funcName string, arg any) int32
    CallWait(targetPath, funcName string, arg, reply any) int32
    CallType(nodeType, actorID, funcName string, arg any) int32
    PostRemote(m *Message)
    PostLocal(m *Message)
    LastAt() int64
    Exit()
}
```

**IActorHandler** - Actor 处理器合约（开发者实现）:
```go
type IActorHandler interface {
    AliasID() string                          // Actor ID
    OnInit()                                  // 启动前触发
    OnStop()                                  // 停止前触发
    OnLocalReceived(m *Message) (bool, bool)  // 本地消息处理
    OnRemoteReceived(m *Message) (bool, bool) // 远程消息处理
    OnFindChild(m *Message) (IActor, bool)    // 查找子 Actor
}
```

**IActorChild** - 子 Actor 管理接口:
```go
type IActorChild interface {
    Create(id string, handler IActorHandler) (IActor, error)
    Get(id string) (IActor, bool)
    Remove(id string)
    Each(fn func(i IActor))
    Call(childID, funcName string, arg any)
    CallWait(targetPath, funcName string, arg, reply any) int32
}
```

**IEventData** - 事件数据接口（可扩展）:
```go
type IEventData interface {
    Name() string    // 事件名
    UniqueID() int64 // 唯一 ID
}
```

**InvokeFunc** - 函数调用回调（可扩展）:
```go
type InvokeFunc func(app IApplication, fi *creflect.FuncInfo, m *Message)
```

### 集群接口

**ICluster** - 集群 RPC 接口:
```go
type ICluster interface {
    Init()
    PublishLocal(nodeID string, packet *cproto.ClusterPacket) error
    PublishRemote(nodeID string, packet *cproto.ClusterPacket) error
    PublishRemoteType(nodeType string, cpacket *cproto.ClusterPacket) error
    RequestRemote(nodeID string, packet *cproto.ClusterPacket, timeout ...time.Duration) ([]byte, int32)
    Stop()
}
```

**IDiscovery** - 节点发现接口（可扩展）:
```go
type IDiscovery interface {
    Load(app IApplication)
    Name() string
    Map() map[string]IMember
    ListByType(nodeType string, filterNodeID ...string) []IMember
    Random(nodeType string) (IMember, bool)
    GetType(nodeID string) (nodeType string, err error)
    GetMember(nodeID string) (member IMember, found bool)
    AddMember(member IMember)
    RemoveMember(nodeID string)
    OnAddMember(listener MemberListener)
    OnRemoveMember(listener MemberListener)
    Stop()
}
```

**IMember** - 成员信息接口:
```go
type IMember interface {
    GetNodeID() string
    GetNodeType() string
    GetAddress() string
    GetSettings() map[string]string
}
```

### 可扩展接口

**ISerializer** - 序列化器接口（开发者可自定义实现）:
```go
type ISerializer interface {
    Marshal(interface{}) ([]byte, error)
    Unmarshal([]byte, interface{}) error
    Name() string
}
```

**INetParser** - 协议解析器接口（开发者可自定义实现）:
```go
type INetParser interface {
    Load(application IApplication)
    AddConnector(connector IConnector)
    Connectors() []IConnector
}
```

**IConnector** - 连接器接口（开发者可自定义实现）:
```go
type IConnector interface {
    IComponent
    Start()
    Stop()
    OnConnect(fn OnConnectFunc)
}
```

### 数据结构

**Message** - 消息结构:
```go
type Message struct {
    BuildTime  int64            // 构建时间（毫秒）
    PostTime   int64            // 投递时间（毫秒）
    Source     string           // 来源 Actor 路径
    Target     string           // 目标 Actor 路径
    targetPath *ActorPath       // 目标路径对象
    FuncName   string           // 调用函数名
    Session    *cproto.Session  // 会话信息
    Args       interface{}      // 参数
    Header     nats.Header      // NATS Header
    Reply      string           // NATS Reply Subject
    IsCluster  bool             // 是否集群消息
    ChanResult chan interface{} // 同步调用结果通道
}
```

**ActorPath** - Actor 路径:
```go
type ActorPath struct {
    NodeID  string  // 节点 ID
    ActorID string  // Actor ID
    ChildID string  // 子 Actor ID
}
```

路径格式: `NodeID.ActorID` 或 `NodeID.ActorID.ChildID`

## 组件生命周期

```go
type IComponentLifecycle interface {
    Set(app IApplication)    // 设置应用引用
    Init()                   // 初始化
    OnAfterInit()            // 初始化后
    OnBeforeStop()           // 停止前
    OnStop()                 // 停止
}
```

流程: `Set → Init → OnAfterInit → [运行中] → OnBeforeStop → OnStop`

## 项目约定

- 所有组件必须实现 `Name()` 返回唯一名称
- 接口使用 `I` 前缀命名（如 `IApplication`）
- ProfileJSON 基于 `jsoniter.Any` 扩展
- Actor 路径使用 `.` 分隔（`const.DOT`）

## 可扩展点

框架提供以下接口供开发者自定义实现：

| 接口 | 位置 | 用途 |
|------|------|------|
| `ISerializer` | `serializer.go` | 自定义序列化方式 |
| `INetParser` | `net_parser.go` | 自定义网络协议 |
| `IConnector` | `connector.go` | 自定义连接器（如 KCP） |
| `IDiscovery` | `cluster.go` | 自定义节点发现服务 |
| `IActorHandler` | `actor.go` | 自定义 Actor 处理逻辑 |
| `IEventData` | `actor.go` | 自定义事件数据结构 |
| `InvokeFunc` | `actor.go` | 自定义函数调用拦截 |

## 注意事项

- `Component` 结构体是基类，提供默认空实现
- 组件注册时检查 Name 重复
- 停止顺序与启动顺序相反
- 开发者可基于接口扩展，框架仅提供基础实现