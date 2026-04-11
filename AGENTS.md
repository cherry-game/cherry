# 项目知识库

**生成时间:** 2026-04-10
**提交版本:** ea5d31e
**分支:** master

## 项目概述

高性能分布式 Golang 游戏服务器框架，基于 Actor Model 构建。采用 Go 1.23+，支持集群模式、单机模式，提供 TCP/WebSocket 连接器，JSON/Protobuf 序列化。

**核心特性**:
- Actor Model：每个 Actor 独立 goroutine，逻辑串行处理
- 集群通信：基于 NATS 的 RPC，支持同步/异步调用
- 节点发现：三种实现（default/nats/etcd）
- 协议可扩展：系统自带 pomelo/simple 两种协议实现，开发者可基于接口自定义协议
- 组件化架构：生命周期统一管理

**参考来源**: 
- 架构设计图: `_docs/game-server-architecture.jpg`
- 模块结构图: `_docs/module-list.jpg`
- 官方文档: https://cherry-game.github.io/

## 项目结构

```
cherry/
├── facade/       # 核心接口定义（IApplication, IActor, ICluster, IComponent 等）
├── net/          # 网络层：Actor 系统、集群通信、连接器、协议解析
│   ├── actor/    # Actor Model 实现（详见 AGENTS.md）
│   ├── cluster/  # NATS 集群 RPC（详见 AGENTS.md）
│   ├── discovery/# 节点发现服务（详见 AGENTS.md）
│   ├── nats/     # NATS 连接池（详见 AGENTS.md）
│   ├── proto/    # Protobuf 协议定义（详见 AGENTS.md）
│   ├── serializer/# 序列化器（详见 AGENTS.md）
│ ├── parser/   # pomelo/simple 协议解析（详见 AGENTS.md）
│ ├── connector/# TCP/WebSocket 连接器（详见 AGENTS.md）
├── extend/       # 扩展工具库（详见 AGENTS.md）
├── profile/      # 配置管理（基于 jsoniter，详见 AGENTS.md）
├── logger/       # Zap 日志封装（详见 AGENTS.md）
├── const/        # 常量、版本号（详见 AGENTS.md）
├── error/        # 错误定义（详见 AGENTS.md）
├── code/         # 返回码定义（详见 AGENTS.md）
├── application.go # 应用入口
├── cherry.go     # AppBuilder 配置
└── go.mod        # Go 1.23.0
```

## 子目录 AGENTS.md 汇总

项目采用分层知识库设计，核心目录均有独立的 AGENTS.md 文件：

| 目录 | AGENTS.md | 核心内容 |
|------|-----------|----------|
| `facade/` | ✓ | 核心接口合约（IApplication, IActor, IComponent 等） |
| `net/actor/` | ✓ | Actor Model 实现（生命周期、消息队列、子 Actor） |
| `net/cluster/` | ✓ | 集群 RPC（PublishLocal/Remote/Type, RequestRemote） |
| `net/discovery/` | ✓ | 节点发现（default/nats/etcd 三种实现） |
| `net/nats/` | ✓ | NATS 连接池（自动重连、请求超时、统计监控） |
| `net/proto/` | ✓ | Protobuf 协议（Member, ClusterPacket, Session） |
| `net/serializer/` | ✓ | 序列化器（JSON/Protobuf 实现） |
| `net/parser/` | ✓ | 协议解析器（可扩展接口，自带 pomelo/simple 实现） |
| `net/connector/` | ✓ | 网络连接器（TCP/WebSocket, TLS 加密） |
| `extend/` | ✓ | 扩展工具库（时间、队列、ID生成、加密等） |
| `profile/` | ✓ | 配置管理（多环境、include 合并、正则节点匹配） |
| `logger/` | ✓ | Zap 日志封装（RotateLogs 切割、软链接、多输出） |
| `const/` | ✓ | 常量定义（版本号、Logo、路径分隔符 DOT） |
| `error/` | ✓ | 错误定义（路由、数据包、集群、Actor 错误） |
| `code/` | ✓ | 返回码定义（OK/RPC/Actor 返回码，IsFail 判断） |

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|-------|
| 创建新应用 | `cherry.go` | `Configure()` + `AppBuilder.Startup()` |
| 添加组件 | `facade/component.go` | 实现 `IComponent` 接口 |
| 创建 Actor | `net/actor/` + `facade/actor.go` | 实现 `IActorHandler` |
| 跨节点 RPC | `net/cluster/nats_cluster/cluster.go` | `PublishRemote`, `RequestRemote` |
| 协议解析 | `net/parser/pomelo/` 或 `net/parser/simple/` | 两种协议可选 |
| 配置读取 | `profile/config.go` | `GetConfig().GetString()` |
| 日志输出 | `logger/logger.go` | `clog.Info`, `clog.Error` |
| 扩展工具 | `extend/` | 时间、队列、ID生成、加密等 |

## 代码地图

| 符号 | 类型 | 位置 | 作用 |
|--------|------|----------|------|
| `Application` | struct | `application.go:27` | 应用主实例，管理组件生命周期 |
| `AppBuilder` | struct | `cherry.go:10` | 配置构建器，`Configure()` 入口 |
| `IActorHandler` | interface | `facade/actor.go:41` | Actor 处理器合约 |
| `IComponent` | interface | `facade/component.go:4` | 组件生命周期合约 |
| `INode` | interface | `facade/application.go:11` | 节点信息接口 |
| `ICluster` | interface | `facade/cluster.go` | 集群 RPC 接口 |
| `Cluster` | struct | `net/cluster/nats_cluster/cluster.go:19` | NATS 集群实现 |
| `Actor` | struct | `net/actor/actor.go:40` | Actor 实例，单 goroutine 运行 |
| `System` | struct | `net/actor/system.go:17` | Actor 系统管理器 |
| `Connector` | struct | `net/connector/connector.go:12` | 连接器基类 |

## 项目约定

**组件命名**: 组件必须实现 `Name()` 方法，注册时检查重复

**Actor 路径**: 格式 `{nodeID}.{actorID}.{childID}`，使用 `const.DOT` 分隔

**消息类型**: 三种队列 - Local（客户端消息）、Remote（Actor 间）、Event（订阅发布）

**配置路径**: Profile 文件支持环境切换，路径如 `profile/settings/{env}/node.json`

**错误码**: 使用 `code/` 包定义返回码，`ccode.IsFail()` 判断失败

## 反模式（本项目特有）

- **DO NOT EDIT** `net/proto/proto.pb.go` - protoc 自动生成文件
- **组件重复注册** - `Register()` 会检查 Name 重复并报错
- **ActorID 为空** - `CreateActor()` 会返回 `ErrActorIDIsNil`
- **运行中注册组件** - `Running()` 为 true 时 `Register()` 直接返回

## 特色设计

### Actor Model

每个 Actor 独立运行在一个 goroutine，所有逻辑串行处理，FIFO 队列消费。

**三种消息队列**:
- Local Mailbox: 客户端发来的本地消息
- Remote Mailbox: Actor 间远程调用消息
- Event Queue: 订阅发布的事件消息

**子 Actor**: 可创建子 Actor，消息由父 Actor 路由转发，子 Actor 不能再创建子 Actor（层级限制）

### 组件生命周期

```
Set → Init → OnAfterInit → [运行中] → OnBeforeStop → OnStop
                    ↑                              ↓
              反向顺序停止 ←────────────────────────┘
```

### 协议解析器（可扩展）

框架提供 `INetParser` 接口，支持开发者自定义协议实现。系统自带两种协议实现作为示例：

**接口定义** (`facade/net_parser.go`):
```go
type INetParser interface {
    Package() IPackage   // 数据包解析
    Message() IMessage   // 消息解析
    Route() IRoute       // 路由解析
    Serialize() ISerializer // 序列化器
}
```

**自带协议实现**:

| 协议 | 位置 | 特点 |
|------|------|------|
| Pomelo | `net/parser/pomelo/` | 网易 Pomelo 框架协议，支持握手/心跳/路由压缩 |
| Simple | `net/parser/simple/` | 简化格式 `id(4B)+len(4B)+data`，无握手过程 |

**自定义协议示例**:
```go
// 1. 实现 INetParser 接口
type MyParser struct { ... }
func (p *MyParser) Package() IPackage { ... }
func (p *MyParser) Message() IMessage { ... }
func (p *MyParser) Route() IRoute { ... }
func (p *MyParser) Serialize() ISerializer { ... }

// 2. 注册到应用
appBuilder.SetNetParser(&MyParser{})
```

### 双运行模式

| 模式 | 说明 | 自动注册组件 |
|------|------|------------|
| `Cluster` | 集群模式，跨节点通信 | cluster + discovery |
| `Standalone` | 单机模式，无跨节点 | 仅 actorSystem |

### 集群架构

典型分布式游戏服务端架构（参考 [_docs/game-server-architecture.jpg](_docs/game-server-architecture.jpg)）：

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           客户端 (H5/Unity/Cocos)                        │
└─────────────────────────────────────────────────────────────────────────┘
         │ HTTP                    │ WebSocket
         ▼                         ▼
┌─────────────┐            ┌─────────────┐            ┌─────────────┐
│   Web 节点   │◄──────────►│  Gate 节点   │◄──────────►│ Center 节点 │
│             │   NATS     │  (网关服)    │   NATS     │  (中心服)   │
│ - 区服列表   │            │ - 连接接入   │            │ - 节点协调  │
│ - SDK 登录  │            │ - 路由转发   │            │ - 路由管理  │
│ - HTTP API │            │ - Session    │            │ - 负载均衡  │
└─────────────┘            └─────────────┘            └─────────────┘
                                   │
                                   │ NATS RPC
                                   ▼
                            ┌─────────────┐
                            │  Game 节点   │
                            │  (游戏服)    │
                            │ - 游戏逻辑  │
                            │ - Actor 系统│
                            │ - 数据存储  │
                            └─────────────┘
                                   │
                                   ▼
                            ┌─────────────┐
                            │   数据层     │
                            │ - MySQL     │
                            │ - MongoDB   │
                            │ - Redis     │
                            └─────────────┘
```

**节点类型说明**:

| 节点类型 | 职责 | 连接方式 |
|---------|------|----------|
| **Web** | HTTP 服务、区服列表、SDK 登录、账号注册 | HTTP/WebSocket |
| **Gate** | 网关服务、客户端连接接入、消息路由转发 | WebSocket |
| **Center** | 中心服务、节点协调、路由管理、负载均衡 | NATS RPC |
| **Game** | 游戏服务、游戏逻辑处理、Actor 系统 | NATS RPC |

**通信方式**:
- 客户端 ↔ Gate: WebSocket/TCP（Pomelo 或 Simple 协议）
- Gate ↔ Center ↔ Game: NATS RPC（同步/异步调用）
- Web ↔ Game: NATS RPC

**数据流向**:
1. 客户端通过 WebSocket 连接 Gate 节点
2. Gate 节点根据路由将消息转发到 Center 或 Game 节点
3. Game 节点处理游戏逻辑，存储数据到数据库
4. Center 节点负责节点协调和路由管理

**节点发现服务**:
- `default`: 从 profile 配置文件读取（适合开发测试）
- `nats`: master-client 模式，支持心跳检测（适合生产）
- `etcd`: 基于 etcd（需安装 components/etcd）

### 扩展组件

已开放组件（独立仓库 `github.com/cherry-game/components`）:
- **data-config**: 策划配表读取管理
- **etcd**: 基于 etcd 的节点发现
- **gin**: HTTP Server 组件
- **gorm**: MySQL 数据库组件
- **mongo**: MongoDB 数据库组件
- **cron**: 定时任务组件

待开放: db 队列、gopher-lua 脚本、限流组件等

## Makefile 使用说明

项目提供 Makefile 简化常用操作，执行 `make` 或 `make help` 可查看所有命令。

| 命令 | 作用 | 详细说明 |
|------|------|----------|
| `make init` | 安装 protoc 工具 | 安装 `protoc-gen-go`，用于生成 Go protobuf 代码 |
| `make protoc` | 编译 proto 文件 | 将 `net/proto/proto.proto` 编译为 `proto.pb.go` |
| `make tag` | 创建版本标签 | 执行 `tag.sh` 脚本创建 Git 版本标签 |
| `make modtidy` | 整理依赖 | 删除旧的 go.sum，重新执行 `go mod tidy`（含 components 子模块） |

**注意事项**:
- 执行 `make protoc` 前需先运行 `make init`
- `make modtidy` 会清理 components 子目录的 go.sum
- proto 文件位置: `net/proto/proto.proto`

## 注意事项

- Go 版本要求 1.23+，toolchain go1.24.2
- NATS 是集群通信核心依赖（启动命令: `nats-server`）
- 示例代码在独立仓库 `github.com/cherry-game/examples`
- 组件库在独立仓库 `github.com/cherry-game/components`
- 文档地址: https://cherry-game.github.io/
- QQ 群交流: 191651647
- 架构设计图: `_docs/game-server-architecture.jpg`
- Pomelo 协议详解: [_docs/pomelo.md](_docs/pomelo.md)
- **代码变化时需更新**: 若以下目录代码发生变化，请同步更新对应的 `AGENTS.md` 文件：
  
  **网络层** (`net/`):
  - [net/actor/](net/actor/AGENTS.md) - Actor Model 实现
  - [net/cluster/](net/cluster/AGENTS.md) - 集群 RPC
  - [net/discovery/](net/discovery/AGENTS.md) - 节点发现服务
  - [net/nats/](net/nats/AGENTS.md) - NATS 连接池
  - [net/proto/](net/proto/AGENTS.md) - Protobuf 协议定义
  - [net/serializer/](net/serializer/AGENTS.md) - 序列化器
  - [net/parser/](net/parser/AGENTS.md) - 协议解析器
  - [net/connector/](net/connector/AGENTS.md) - 网络连接器
  
  **核心模块**:
  - [facade/](facade/AGENTS.md) - 核心接口定义
  - [extend/](extend/AGENTS.md) - 扩展工具库
  - [profile/](profile/AGENTS.md) - 配置管理
  - [logger/](logger/AGENTS.md) - 日志封装
  
  **基础模块**:
  - [const/](const/AGENTS.md) - 常量定义
  - [error/](error/AGENTS.md) - 错误定义
  - [code/](code/AGENTS.md) - 返回码定义

## 环境准备

参考 [_docs/env-setup.md](_docs/env-setup.md)：

| 工具 | 安装方式 | 说明 |
|------|----------|------|
| Go 1.23+ | 官方下载 / brew | 设置代理 `GOPROXY=https://goproxy.cn,direct` |
| NATS Server | docker / brew | 默认端口 4222 |
| VS Code / GoLand | 推荐 GoLand | 安装 Go 插件和调试配置 |