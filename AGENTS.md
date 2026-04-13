# 项目知识库

## 维护策略

本仓库采用“一个根文档 + 少量核心模块文档”的结构。

- 根 `AGENTS.md` 负责全局事实、启动顺序、模块索引、跨模块约束
- 子目录 `AGENTS.md` 只保留在高复杂度、高风险、强局部约束的模块
- 轻量目录不再单独维护 `AGENTS.md`，避免重复、漂移和维护噪音

当前保留的子目录文档只有：

- `facade/`
- `net/actor/`
- `net/cluster/`
- `net/discovery/`
- `net/nats/`
- `net/parser/`
- `profile/`

其余目录知识统一回收到根文档或由代码本身表达。

## 文档书写规范

`AGENTS.md` 统一使用相对路径，不使用仓库绝对路径。

- 根目录文档引用仓库文件时使用 `./...`
- 子目录文档优先使用相对当前目录的 `./...`
- 跨目录引用时使用最短可读的 `../...` 或 `../../...`
- 行号写在链接目标末尾，例如 `./application.go:151`
- 如果某条规则已经能在根文档表达，就不要在多个子目录重复抄写
- 子目录文档只写局部知识：入口、局部约束、常见坑、联动检查

## 项目概览

`cherry` 是一个基于 Actor Model 的 Golang 游戏服务器框架，核心目标是：

- 节点内执行模型串行化
- 跨节点通信统一化
- 应用生命周期组件化

项目的主链路不是按目录平铺理解，而是按这 4 条链理解：

- 应用装配链：`cherry.go` -> `application.go`
- Actor 执行链：`net/actor/`
- 集群通信链：`net/cluster/` + `net/nats/` + `net/discovery/`
- 前端接入链：`net/parser/` + `net/connector/`

## 启动模型

最重要的入口：

- [Configure](./cherry.go:16)
- [AppBuilder.Startup](./cherry.go:34)
- [Application.Startup](./application.go:151)

真实启动顺序：

1. `Configure(...)` 或 `ConfigureNode(...)` 创建 `Application`
2. `AppBuilder.Startup()` 在 `Cluster` 模式下自动注入 `cluster_component` 和 `discovery_component`
3. 注册自定义组件
4. `Application.Startup()` 自动注册 `actor_component`
5. 如果设置了 `netParser`，把 parser 挂载的 connector 一并注册成组件
6. 所有组件按注册顺序执行 `Set -> Init -> OnAfterInit`
7. 如果 `isFrontend == true`，此时必须已经设置 `netParser`，否则直接 `panic`
8. 应用进入阻塞运行，直到收到系统信号或调用 `Shutdown()`
9. 停机时按组件注册逆序执行 `OnBeforeStop -> OnStop`

## 全局约定

- Actor 路径格式：`{nodeID}.{actorID}` 或 `{nodeID}.{actorID}.{childID}`
- 组件 `Name()` 必须全局唯一；重复注册会被忽略并记录日志
- `Application.Register()` 只能在启动前调用，运行中注册无效
- `actor_component` 由框架自动注册，不需要上层重复注册
- `Cluster` 模式自动注册 `cluster_component` 和 `discovery_component`
- 前端节点通常通过 `netParser` 驱动 connector，而不是手动调用 connector `Start()`
- `Application.Startup()` 是阻塞调用，不是“初始化后立刻返回”

## 重点模块索引

- 看应用装配与生命周期：`cherry.go`、`application.go`
- 看公共契约：`facade/`
- 看 Actor 调用与邮箱模型：`net/actor/`
- 看集群组件封装：`net/cluster/`
- 看 NATS 连接池、请求与 waiter 清理：`net/nats/`
- 看节点发现与心跳：`net/discovery/`
- 看前端 parser 与 agent 入口：`net/parser/`
- 看配置读取与节点匹配：`profile/`

轻量目录的关键约束统一放这里：

- `net/connector/`：连接器一般由 parser 接管；直接 `Start()` 前必须先设置 `OnConnect`
- `net/proto/`：`proto.pb.go` 为生成文件；`ClusterPacket` 使用对象池，发送后不可继续持有
- `net/serializer/`：默认序列化器是 protobuf；切换前要确认 cluster / actor reply 兼容
- `logger/`：大部分调用走全局 `DefaultLogger`；节点日志配置来自 `__settings__.ref_logger`
- `extend/`：`extend/reflect/` 影响 Actor 反射调用；`extend/file/json/string/` 影响 `profile`
- `code/` 与 `error/`：分别承载返回码和错误文本；涉及 RPC / Actor 语义时通常要一起看
- `const/`：`DOT` 影响 ActorPath 拼接与解析，不能随意改

## 容易踩坑的地方

- `isFrontend == true` 但没有设置 parser，会在启动阶段 `panic`
- `INetParser` 的真实职责是装配 connector 并加载前端 agent actor，不只是编解码
- `ClusterPacket` 使用对象池，发完后通常会被 `Recycle()`，不能复用
- `profile`、`logger`、`net/nats` 都有包级全局状态，同进程多实例要格外小心
- `discovery` 的 `nats` 模式依赖后台 goroutine 与心跳，排查问题时要同时看 discovery 和 nats

## 测试与验证

- 全量回归：`go test ./...`
- 涉及 NATS / discovery / cluster 的改动，通常还需要结合真实配置与运行环境验证
- 如果修改公共契约或全局行为，应同步检查根文档和对应子模块文档是否仍然准确

## 联动检查

- 改 `application.go`：同步检查 `cherry.go`、`facade/application.go`、根 `AGENTS.md`
- 改 `cherry.go`：同步检查 `application.go`、`facade/AGENTS.md`、根 `AGENTS.md`
- 改 `facade/*`：同步检查对应实现目录和根 `AGENTS.md`
- 改 `net/actor/*`：同步检查 `facade/actor.go`、`net/cluster/`、`net/proto/`、`code/`、`error/`
- 改 `net/cluster/*`：同步检查 `net/nats/`、`net/discovery/`、`net/proto/`、`code/`、`error/`
- 改 `net/discovery/*`：同步检查 `profile/`、`net/nats/`、`net/proto/`
- 改 `net/parser/*`：同步检查 `net/connector/`、`net/proto/`、`facade/net_parser.go`
- 改 `profile/*`：同步检查 `application.go`、`logger/`、`net/discovery/`
