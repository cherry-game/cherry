# net/actor

## 角色

`net/actor` 是框架执行内核。每个 Actor 独占 goroutine，通过串行消费消息来承接状态和会话逻辑。

## 真实入口

- [component.go](./component.go:9)
- [system.go](./system.go:17)
- [actor.go](./actor.go)

## 启动链

1. `Application.Startup()` 自动注册 `actor_component`
2. `actor_component.Init()` 注入 `app` 到 `System`
3. `actor_component.OnAfterInit()` 遍历 `AddActors(...)` 收集到的 handler
4. 每个 handler 通过 `CreateActor()` 创建真实 Actor 并启动 goroutine

## 关键行为

- `CreateActor()`：
  - `actorID` 为空会报错
  - 同名 Actor 已存在时直接返回已有实例
  - 创建成功后立即 `go thisActor.run()`
- `Call()`：
  - 本机节点走本地 `Remote` 队列
  - 跨节点调用会序列化参数并转成 `ClusterPacket`
- `CallWait()`：
  - 本地调用走 `ChanResult`
  - 跨节点调用走 `Cluster.RequestRemote`
  - `source == target` 会直接返回错误码
- `PostLocal()` / `PostRemote()`：
  - 按路径查找 Actor 或子 Actor
  - 再投递到对应邮箱

## 运行模型

Actor 主循环会串行处理这些输入：

- `Local` 邮箱
- `Remote` 邮箱
- `Event` 队列
- `Timer`
- `close` 信号

默认超时见 [system.go:30](./system.go:30)：

- `callTimeout = 3s`
- `arrivalTimeOut = 100ms`
- `executionTimeout = 100ms`

这几个值是 Actor 处理链路的超时保护，不是网络超时。

## 局部约束

- 子 Actor 只能由父 Actor 创建；子 Actor 不能继续创建子 Actor
- 子 Actor 路径格式为 `node.actor.child`
- `System.Stop()` 会等待全部 Actor 退出；单个 Actor 卡住会拖住整个应用停机

## 常见坑

- 上层如果假设 handler 可并发执行，会和 Actor 串行模型冲突
- `CreateActor()` 是幂等返回已有实例，不会替换旧 handler
- 子 Actor 生命周期依附父 Actor，排查问题别只看顶层 Actor

## 联动检查

- 改 `system.go`：同步检查 `facade/actor.go`、`net/cluster/`、`code/code.go`、`error/error.go`
- 改 `actor.go`：同步检查 `facade/message.go`、`invoke.go`、`actor_mailbox.go`
- 改 `invoke.go`：同步检查 `extend/reflect/`、`facade/actor.go`
- 改 `actor_child.go`：同步检查 `facade/message.go` 和路径规则
