# net/actor - Actor Model 实现

## 概述

Actor 系统核心实现，每个 Actor 独立运行在单个 goroutine，逻辑串行处理。

## 目录结构

```
net/actor/
├── actor.go          # Actor 实体定义、消息循环
├── actor_base.go     # Actor 基础功能扩展
├── actor_child.go    # 子 Actor 管理
├── actor_mailbox.go  # 消息邮箱（Local/Remote 队列）
├── actor_timer.go    # 定时器管理
├── actor_event.go    # 事件订阅处理
├── system.go         # Actor 系统管理器
├── component.go      # Actor 组件（注册到 Application）
├── invoke.go         # 函数调用实现
├── queue.go          # 消息队列
├── timer.go          # 定时器接口
├── const.go          # Actor 常量、错误定义
└── facade.go         # 接口定义
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|-------|
| 创建 Actor | `system.go:119` | `CreateActor(id, handler)` |
| Actor 生命周期 | `actor.go:57-66` | `run()` → `onInit()` → `loop()` → `onStop()` |
| 发送消息 | `actor.go:367-372` | `PostRemote()`, `PostLocal()` |
| 函数调用 | `actor.go:186-237` | `invokeFunc()` - 反射调用 |
| 子 Actor | `actor_child.go` | `Create()`, `Get()`, `Remove()` |
| 定时器 | `actor_timer.go` | `AddTimer()`, `RemoveTimer()` |
| 事件订阅 | `actor_event.go` | `Subscribe()`, `Unsubscribe()` |

## 关键接口

```go
// IActorHandler - Actor 处理器合约
type IActorHandler interface {
    AliasID() string                          // Actor ID
    OnInit()                                  // 启动前触发
    OnStop()                                  // 停止前触发
    OnLocalReceived(m *Message) (bool, bool)  // 本地消息处理
    OnRemoteReceived(m *Message) (bool, bool) // 远程消息处理
    OnFindChild(m *Message) (IActor, bool)    // 查找子 Actor
}

// IActor - Actor 实例接口
type IActor interface {
    App() IApplication
    ActorID() string
    Path() *ActorPath
    Call(targetPath, funcName string, arg any) int32
    CallWait(targetPath, funcName string, arg, reply any) int32
    PostRemote(m *Message)
    PostLocal(m *Message)
    Exit()
}
```

## 消息流转

**三种消息队列**:
1. **Local Mailbox** - 客户端发来的本地消息
2. **Remote Mailbox** - Actor 间远程调用消息
3. **Event Queue** - 订阅发布的事件消息

**处理顺序**: FIFO，每类消息独立队列

**消息循环** (`actor.go:68-101`):
```go
select {
case <-p.localMail.C:    processLocal()
case <-p.remoteMail.C:   processRemote()
case <-p.event.C:        processEvent()
case <-p.timer.C:        processTimer()
case <-p.close:          state = StopState
}
```

## 项目约定

- Actor 路径格式: `{nodeID}.{actorID}.{childID}`
- 状态机: `InitState(0)` → `WorkerState(1)` → `FreeState(2)` → `StopState(3)`
- 停止时等待所有队列清空后才退出
- 子 Actor 不能再创建子 Actor（层级限制）

## 反模式

- 子 Actor 尝试创建子 Actor → 日志警告并返回 nil
- ActorID 为空 → 返回 `ErrActorIDIsNil`
- 消息超时: `arrivalTimeOut` (100ms), `executionTimeout` (100ms)

## 示例代码

```go
// 创建 Actor 处理器
type MyActor struct {
    cherryActor.Base
}

func (p *MyActor) AliasID() string { return "my-actor" }
func (p *MyActor) OnInit() { /* 初始化 */ }
func (p *MyActor) OnLocalReceived(m *cfacade.Message) (bool, bool) {
    // 处理本地消息
    return true, true // next=true, invoke=true
}

// 注册
appBuilder.AddActors(&MyActor{})
```