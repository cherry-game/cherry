# net/nats - NATS 连接池

## 概述

NATS 连接池管理，支持多连接、自动重连、请求超时、统计监控等功能。

## 目录结构

```
net/nats/
├── pool.go       # 连接池管理
├── connect.go    # 单连接实现（Subscribe, Request, Publish）
└── msg_pool.go   # 消息对象池
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 创建连接池 | `pool.go:19` | `NewPool(replySubject, config, isConnect)` |
| 获取连接 | `pool.go:58` | `GetConnect()` 轮询获取 |
| 关闭所有连接 | `pool.go:63` | `ConnectClose()` |
| 连接服务器 | `connect.go:55` | `Connect()` 自动重连 |
| 发布消息 | `connect.go:170` | `PublishMsg(msg)` |
| 同步请求 | `connect.go:156` | `RequestSync(subject, data, timeout)` |
| 订阅消息 | `connect.go:195` | `Subscribe(subject, cb)` |

## 连接池特性

**轮询获取** (`pool.go:58-61`):
```go
func GetConnect() *Connect {
    index := atomic.AddUint64(roundIndex, 1)
    return connectPool[index%connectSize]  // 轮询分配
}
```

**自动重连** (`connect.go:60-76`):
- 连接失败后每 3 秒重试
- 支持配置最大重连次数

**请求超时**:
- 默认超时: 2 秒
- 可通过配置 `request_timeout` 调整

## 配置参数

```json
{
  "address": "nats://127.0.0.1:4222",
  "user": "",               // 用户名（可选）
  "password": "",           // 密码（可选）
  "pool_size": 1,           // 连接池大小
  "max_reconnects": 10,     // 最大重连次数
  "reconnect_delay": 1,     // 重连延迟（秒）
  "request_timeout": 2,     // 请求超时（秒）
  "is_stats": false         // 是否开启统计
}
```

## 关键结构

**Connect** (`connect.go:19-28`):
```go
type Connect struct {
    *nats.Conn       // NATS 连接
    id      int      // 连接 ID
    seq     uint64   // 请求序列号
    waiters sync.Map // 等待响应的请求
    subs    []*nats.Subscription // 订阅列表
    reply   string   // 响应 subject
}
```

## RequestSync 实现原理

1. 生成唯一 reqID
2. 创建等待 channel
3. 发送消息，Header 携带 reqID
4. 监听 reply subject，匹配 reqID 返回

## 统计监控

开启 `is_stats: true` 后，每 30 秒输出：
- InMsgs / OutMsgs: 收发消息数
- InBytes / OutBytes: 收发字节
- Reconnects: 重连次数

## 项目约定

- 使用对象池 (`msg_pool.go`) 减少 GC
- 所有订阅在连接关闭时自动取消

## 注意事项

- 连接池大小建议与并发请求量匹配
- `RequestSync` 超时后会清理等待 channel
- 重连时会输出警告日志