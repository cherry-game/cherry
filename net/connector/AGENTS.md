# net/connector - 网络连接器

## 概述

提供 TCP 和 WebSocket 网络连接器实现，支持 TLS 加密、连接回调、并发连接处理。

## 目录结构

```
net/connector/
├── connector.go         # 连接器基类（Connector）
├── tcp_connector.go     # TCP 连接器实现
├── ws_connector.go      # WebSocket 连接器实现
├── options.go           # 配置选项
├── tcp_connector_test.go # TCP 测试
└── ws_connector_test.go  # WebSocket 测试
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 创建 TCP 连接器 | `tcp_connector.go:27` | `NewTCP(address, opts...)` |
| 创建 WebSocket 连接器 | `ws_connector.go:41` | `NewWS(address, opts...)` |
| 设置连接回调 | `connector.go:28` | `OnConnect(fn)` |
| 启动连接器 | `tcp_connector.go:51`, `ws_connector.go:72` | `Start()` |
| 停止连接器 | `connector.go:50` | `Stop()` |
| TLS 配置 | `options.go:18` | `WithCert(certFile, keyFile)` |
| 通道大小配置 | `options.go:29` | `WithChanSize(size)` |

## 关键接口

**IConnector** (`facade/connector.go:6-11`):
```go
type IConnector interface {
    IComponent
    Start()                     // 启动连接器
    Stop()                      // 停止连接器
    OnConnect(fn OnConnectFunc) // 建立新连接时触发
}
```

**OnConnectFunc** (`facade/connector.go:14`):
```go
type OnConnectFunc func(conn net.Conn)
```

## 两种连接器

### TCP Connector

- 基于 Go 标准库 `net` 实现
- 支持 TLS 加密（通过证书配置）
- 默认通道大小: 256

```go
// 创建 TCP 连接器
tcp := cherryConnector.NewTCP("127.0.0.1:8080")
tcp.OnConnect(func(conn net.Conn) {
    // 处理新连接
})
app.Register(tcp)

// 带 TLS
tcp := cherryConnector.NewTCP("127.0.0.1:8080",
    cherryConnector.WithCert("server.crt", "server.key"),
)
```

### WebSocket Connector

- 基于 `gorilla/websocket` 实现
- 支持 HTTP 协议升级
- 自定义 Upgrader 配置

```go
// 创建 WebSocket 连接器
ws := cherryConnector.NewWS("127.0.0.1:8080")
ws.OnConnect(func(conn net.Conn) {
    // 处理新连接
})
app.Register(ws)

// 自定义 Upgrader
ws.SetUpgrade(&websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        return true // 允许跨域
    },
})
```

## WSConn 适配器

WebSocket 连接通过 `WSConn` 适配为标准 `net.Conn` 接口：

```go
type WSConn struct {
    *websocket.Conn
    typ    int       // message type
    reader io.Reader
}
```

支持方法：
- `Read(b []byte)` - 读取数据
- `Write(b []byte)` - 写入数据（BinaryMessage）
- `SetDeadline(t time.Time)` - 设置超时

## 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithCert(cert, key)` | TLS 证书配置 | "" |
| `WithChanSize(size)` | 连接通道大小 | 256 |

## 组件名称

| 连接器 | Name() 返回值 |
|--------|---------------|
| TCPConnector | `"tcp_connector"` |
| WSConnector | `"websocket_connector"` |

## 连接处理流程

```
┌──────────────┐
│   Listener   │  监听端口
└──────┬───────┘
       │ Accept()
       ▼
┌──────────────┐
│   connChan   │  连接通道（并发缓冲）
└──────┬───────┘
       │ goroutine
       ▼
┌──────────────┐
│ onConnectFunc│  用户回调处理
└──────────────┘
```

## 项目约定

- 连接器作为组件注册到 Application
- 新连接通过 channel 异步处理
- `OnConnect` 回调必须设置，否则 Start 时 panic

## 使用示例

```go
// TCP 示例
func main() {
    appBuilder := cherry.Configure(...)
    
    tcp := cherryConnector.NewTCP("0.0.0.0:8080")
    tcp.OnConnect(func(conn net.Conn) {
        // 创建 Agent 处理连接
        agent := pomelo.NewAgent(conn, ...)
        agent.Run()
    })
    
    appBuilder.Register(tcp)
    appBuilder.Startup()
}

// WebSocket 示例
func main() {
    appBuilder := cherry.Configure(...)
    
    ws := cherryConnector.NewWS("0.0.0.0:8080")
    ws.OnConnect(func(conn net.Conn) {
        // 创建 Agent 处理连接
        agent := pomelo.NewAgent(conn, ...)
        agent.Run()
    })
    
    appBuilder.Register(ws)
    appBuilder.Startup()
}
```

## 注意事项

- 连接器地址为空时会返回 nil
- TLS 需同时配置 certFile 和 keyFile
- WebSocket 默认允许跨域（CheckOrigin: true）
- 停止时关闭 listener 并等待连接处理完成