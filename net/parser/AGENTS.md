# net/parser - 协议解析器

## 概述

框架提供 `INetParser` 可扩展接口，支持开发者自定义协议实现。系统自带两种协议实现作为示例：pomelo（网易 Pomelo 协议）和 simple（自定义简化协议）。

## 目录结构

```
net/parser/
├── pomelo/         # Pomelo 协议实现
│   ├── pomelo.go   # 主解析器
│   ├── actor.go    # Actor 处理器
│   ├── actor_base.go
│   ├── agent.go    # 连接代理
│   ├── agents.go   # 代理管理
│   ├── command.go  # 命令定义
│   ├── client/     # Pomelo 客户端
│   ├── message/    # 消息结构
│   └── packet/     # 数据包结构
└── simple/         # Simple 协议实现
│   ├── simple.go   # 主解析器
│   ├── actor.go    # Actor 处理器
│   ├── actor_base.go
│   ├── agent.go    # 连接代理
│   ├── agents.go   # 代理管理
│   ├── message.go  # 消息结构
│   ├── route.go    # 路由定义
│   └── define.go   # 常量定义
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|-------|
| Pomelo 解析 | `pomelo/pomelo.go` | `New()` 创建解析器 |
| Simple 解析 | `simple/simple.go` | `New()` 创建解析器 |
| 消息路由 | `pomelo/message/route.go` | Route 结构 |
| 数据包格式 | `pomelo/packet/packet.go` | Handshake, Data, Kick 等 |
| 代理管理 | `pomelo/agents.go`, `simple/agents.go` | Agent 生命周期 |

## 协议接口

**INetParser** (`facade/net_parser.go`):
```go
type INetParser interface {
    Package() IPackage   // 数据包解析
    Message() IMessage   // 消息解析
    Route() IRoute       // 路由解析
    Serialize() ISerializer // 序列化器
}
```

开发者可基于此接口实现自定义协议，通过 `appBuilder.SetNetParser()` 注册。

## 自带协议实现

### Pomelo 协议 (`pomelo/`)
- 基于 Pomelo 框架协议格式
- 支持握手、心跳、数据包
- 消息路由: `serverType.handler.method`
- 支持消息字典压缩

### Simple 协议 (`simple/`)
- 简化格式: `id(4bytes) + dataLen(4bytes) + data(n bytes)`
- 更轻量，适合自定义场景
- 无握手过程

## 关键结构

**pomelo/packet/packet.go**:
```go
type Packet struct {
    Type   byte   // Handshake, HandshakeAck, Heartbeat, Data, Kick
    Length int    // 数据长度
    Data   []byte // 数据内容
}
```

**pomelo/message/message.go**:
```go
type Message struct {
    Type    byte   // Request, Notify, Response, Push
    ID      uint   // 消息 ID
    Route   string // 路由路径
    Data    []byte // 数据
}
```

**simple/message.go**:
```go
type Message struct {
    MsgID uint32 // 消息 ID (4 bytes)
    Data  []byte // 数据内容
}
```

## 项目约定

- Agent 代表单个连接
- Agents 管理所有 Agent
- ActorHandler 处理路由消息

## 自定义协议

开发者可基于 `INetParser` 接口实现自定义协议：

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