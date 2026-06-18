# 🍒 欢迎使用 cherry！

![cherry logo](https://img.shields.io/badge/cherry--game-cherry-red)
![cherry license](https://img.shields.io/github/license/cherry-game/cherry)
![go version](https://img.shields.io/github/go-mod/go-version/cherry-game/cherry)
![cherry tag](https://img.shields.io/github/v/tag/cherry-game/cherry)

[🌐 English Documentation](README.en.md)

- **高性能分布式的 Golang 游戏服务器框架**
- 采用 Golang + Actor Model 构建，具备高性能、可伸缩等特性
- 简单易学，让开发者更专注于游戏业务开发

## 📦 安装

```bash
go get github.com/cherry-game/cherry
```

需要 Go 1.24+。

## 🚀 快速开始

```go
package main

import (
    cherry "github.com/cherry-game/cherry"
)

func main() {
    app := cherry.Configure(
        "etc/profile/dev.json", // profile 路径
        "game-1",               // 节点 ID
        true,                   // 是否前端节点（接收客户端连接）
        cherry.Cluster,         // 集群模式
    )
    app.Startup()
}
```

## 🗺️ 架构

![game-server-architecture](_docs/game-server-architecture.jpg)

| 模块 | 路径 | 职责 |
|------|------|------|
| 应用装配 | `cherry.go` → `application.go` | 组件注册、生命周期、启动 |
| Actor 执行 | `net/actor/` | 每 Actor 独立 goroutine，串行 mailbox，本地/远程/事件分发 |
| 集群通信 | `net/cluster/` + `net/nats/` + `net/discovery/` | 跨节点 RPC，NATS 传输，成员发现 |
| 前端接入 | `net/parser/` + `net/connector/` | 协议解码，会话管理，agent Actor，WebSocket/TCP |

## 🌟 核心功能

### AppBuilder 与生命周期

通过 Builder API 将组件链式装配为运行中的服务：

```go
cherry.Configure("etc/profile/dev.json", "game-1", true, cherry.Cluster).
    Register(myComponent).
    SetSerializer(cherryFacade.NewProtobuf()).
    AddActors(myActor).
    Startup()
```

生命周期保证：

```
Register → Set → Init → OnAfterInit → (运行中) → OnBeforeStop → OnStop
```

组件按注册的逆序停止。通过 `SIGINT` / `SIGQUIT` / `SIGTERM` 信号触发优雅关闭。

### Actor 模型

每个 Actor 运行于独立 goroutine，逻辑串行处理。三种独立队列按 FIFO 原则消费：

| 队列 | 来源 | 用途 |
|------|------|------|
| **Local** | 客户端 → Actor | 处理玩家请求 |
| **Remote** | Actor → Actor（跨节点） | 服务间 RPC |
| **Event** | 系统事件 | 解耦通知 |

```go
type MyActor struct {
    capp.ActorLogger
}

func (p *MyActor) OnInit() {
    p.Local().Register("myHandler", p.handle)
    p.Remote().Register("myRemote", p.remote)
    p.EventRegister("eventName", p.onEvent)
}

func (p *MyActor) handle(session *cproto.Session, req *pb.MyReq) {
    p.Response(session, &pb.MyResp{Value: "ok"})
}
```

**子 Actor** — Actor 可创建子 Actor，消息由父 Actor 路由转发，子 Actor 与父 Actor 共享生命周期。

**Actor Timer** — 直接在 Actor 上注册定时器和定时任务，保证在 Actor 的 goroutine 上执行。

### 集群 & 注册发现

三种发现后端，通过 profile 配置切换：

| 模式 | 配置值 | 适用场景 |
|------|--------|----------|
| `default` | 从 profile 配置读取 | 单进程开发/测试 |
| `nats` | NATS 主从心跳发现 | 多节点生产环境 |
| `etcd` | etcd lease + watch | 多节点生产环境（独立仓库） |

```go
discovery := app.Discovery()

// 成员查询
all := discovery.Map()
games := discovery.ListByType("game")
member, ok := discovery.Random("gate")

// Settings 同步（自动广播到所有节点）
discovery.UpdateSetting("region", "us-east")

// 成员变更回调
discovery.OnAddMember(func(m IMember) {
    // 新节点加入
})
discovery.OnRemoveMember(func(m IMember) {
    // 节点离开
})
```

基于 NATS 实现跨节点 RPC 调用，支持同步/异步方式，可配置超时时间。

### 连接器 & 协议

内置连接器：

- **TCP** — 原生 Socket
- **WebSocket** — 浏览器客户端
- **HTTP Server** — REST API
- **HTTP Client** — 对外请求

两种协议格式：

| 协议 | 格式 | 适用场景 |
|------|------|----------|
| **Pomelo** | `type(1b) + length(3b) + data` | 兼容 [pomelo](https://github.com/NetEase/pomelo) 各平台客户端 |
| **Simple** | `id(4b) + dataLen(4b) + data` | 自定义轻量协议 |

### 消息 & 序列化

基于对象池的零分配消息传递：

```go
msg := cherryFacade.GetMessage()
msg.Source = "game-1.player-100"
msg.Target = "map-1.aoi"
msg.FuncName = "enter"
msg.Args = myPayload
```

- `sync.Pool` 消息池 + 引用计数，减少 GC 压力
- 默认 **Protobuf** 序列化，同时支持 **JSON**
- 同进程消息免序列化直传；跨节点消息通过 `ClusterPacket` 自动编解码

### Extend 工具库

框架内置 21 个工具包：

| 分类 | 包 |
|------|-----|
| **时间** | `time_wheel`（分层时间轮），`time`（辅助函数、时间偏移、时间旅行） |
| **ID 生成** | `snowflake`（分布式唯一 ID），`nuid` |
| **数据处理** | `compress`（zlib），`crypto`（MD5、CRC32、Base64），`gob`，`json` |
| **集合** | `map`，`slice`，`queue`，`string` |
| **基础设施** | `file`，`http/client`，`net`，`sync`（限流器） |
| **反射** | `reflect`，`mapstructure`，`regex`（带缓存） |
| **编码** | `base58` |

```go
// 时间轮 — 调度器、超时、延迟执行
tw := cherryTimeWheel.NewTimeWheel(time.Millisecond*10, 20)
tw.AfterFunc(time.Second, func() { /* ... */ })

// Snowflake 唯一 ID
node, _ := cherrySnowflake.NewNode(1)
id := node.Generate()
```

### 错误码

框架内置分层错误码：

```go
if code.IsFail(errCode) {
    p.ResponseCode(session, errCode)
    return
}
```

预定义错误码：`OK(0)`，会话错误（`10+`），RPC 错误（`20+`），Actor 错误（`24+`）。

### 日志

基于 [uber-go/zap](https://github.com/uber-go/zap) 封装：

- 结构化日志，支持键值对输出
- 多文件输出 + 日志切割
- 可配置日志级别和堆栈跟踪级别
- 控制台与文件可同时输出

## 🧰 扩展组件

### 已开放组件

| 组件 | 说明 |
|------|------|
| **data-config** | 策划配表读取管理，支持多种加载方式及数据查询 |
| **etcd** | 基于 etcd 的集群注册发现 |
| **gin** | 集成 gin 实现 HTTP server，支持中间件 |
| **gorm** | MySQL 数据库访问，支持多数据库配置 |
| **mongo** | MongoDB 访问，支持多数据库配置 |
| **cron** | 基于 robfig/cron 的定时任务 |

仓库地址：[cherry-game/components](https://github.com/cherry-game/components)

### 待开放组件

DB 写队列、gopher-lua 脚本、限流组件等

## 📖 示例

### 单节点精简版聊天室

适合新手熟悉项目，具备以下特性：

- 基于网页客户端，构建 HTTP 服务器
- 采用 WebSocket 作为连接器
- 使用 JSON 作为通信格式
- 实现创建房间、发送消息、广播消息等功能

源码位置：[examples/demo_chat](https://github.com/cherry-game/examples/tree/master/demo_chat)

### 多节点分布式游戏示例

适合作为基础框架构建游戏服务端，特性如下：

- 基于 H5 构建客户端
- 搭建 Web 服、网关服、中心服、游戏服等节点
- 实现区服列表、多 SDK 帐号体系、帐号注册、登录、创建角色等功能

源码位置：[examples/demo_cluster](https://github.com/cherry-game/examples/tree/master/demo_cluster)

准备步骤详见：[环境安装与配置](https://cherry-game.github.io/guides/install-go.html)

## 🎮 游戏客户端 SDK

兼容 pomelo 协议的客户端均可接入 Cherry。

| 平台 | 客户端 |
|------|--------|
| **Unity3D** | [YMoonRiver/Pomelo_UnityWebSocket](https://github.com/YMoonRiver/Pomelo_UnityWebSocket-2.7.0)、[NetEase/pomelo-unityclient](https://github.com/NetEase/pomelo-unityclient) 等 |
| **Cocos2d-x** | [NetEase/pomelo-cocos2dchat](https://github.com/NetEase/pomelo-cocos2dchat) |
| **JavaScript** | [pomelonode/pomelo-jsclient-websocket](https://github.com/pomelonode/pomelo-jsclient-websocket) 等 |
| **C** | [topfreegames/libpitaya](https://github.com/topfreegames/libpitaya)、[NetEase/libpomelo](https://github.com/NetEase/libpomelo/) 等 |
| **iOS** | [NetEase/pomelo-iosclient](https://github.com/NetEase/pomelo-iosclient) 等 |
| **Android / Java** | [NetEase/pomelo-androidclient](https://github.com/NetEase/pomelo-androidclient) 等 |
| **微信** | [wangsijie/pomelo-weixin-client](https://github.com/wangsijie/pomelo-weixin-client) |

协议格式：[协议结构图](_docs/pomelo-protocol.jpg)、[pomelo wiki 协议格式](https://github.com/NetEase/pomelo/wiki/%E5%8D%8F%E8%AE%AE%E6%A0%BC%E5%BC%8F)

## 💬 讨论与交流

- QQ 群：[191651647](https://jq.qq.com/?_wv=1027&k=vdIddlK0)

## 🙏 致谢

- [pomelo](https://github.com/NetEase/pomelo) — 初代 Node.js 游戏服务端框架
- [pitaya](https://github.com/topfreegames/pitaya) — Go 游戏服务端框架
