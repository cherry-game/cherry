# net/nats - NATS 连接池

## 概述

NATS 连接池管理，支持多连接、自动重连、请求超时、统计监控等功能。

## 目录结构

```
net/nats/
├── connect_pool.go  # 连接池管理
├── connect.go       # 单连接实现（Subscribe, Request, Publish）
├── msg_pool.go      # NatsMsg 消息对象池
└── timer_pool.go    # Timer 对象池
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 创建连接池 | `connect_pool.go:19` | `NewConnectPool(replySubject, config, isConnect)` |
| 获取连接池列表 | `connect_pool.go:54` | `GetConnectPool()` |
| 获取单个连接 | `connect_pool.go:58` | `GetConnect()` 轮询获取 |
| 关闭连接池 | `connect_pool.go:63` | `CloseConnectPool()` |
| 连接服务器 | `connect.go:55` | `Connect()` 自动重连 |
| 发布消息 | `connect.go:180` | `PublishMsg(msg)` |
| 同步请求 | `connect.go:167` | `RequestSync(subject, data, timeout)` |
| 订阅消息 | `connect.go:207` | `Subscribe(subject, cb)` |
| 获取消息对象 | `msg_pool.go:21` | `GetNatsMsg()` 从池获取 |
| 释放消息对象 | `msg_pool.go:29` | `NatsMsg.Release()` 归还池 |
| 获取 Timer | `timer_pool.go` | `acquireTimer(timeout)` |
| 释放 Timer | `timer_pool.go` | `releaseTimer(timer)` |

## 连接池特性

**轮询获取** (`connect_pool.go:58-61`):
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

**NatsMsg** (`msg_pool.go:17-19`):
```go
type NatsMsg struct {
    *nats.Msg
}

func (m *NatsMsg) Release()  // 归还对象池
```

## RequestSync 实现原理

1. 生成唯一 reqID（每个 Connect 独立计数）
2. 创建等待 channel
3. 从 NatsMsg 池获取消息对象
4. 发送消息，Header 携带 reqID
5. 从 Timer 池获取 Timer，等待响应或超时
6. 归还 NatsMsg 和 Timer 到池

**并发安全设计**:
- `LoadAndDelete` 原子操作，防止 channel 双重关闭
- 阻塞发送响应，防止消息丢失
- Disconnect/Close 回调清理 waiters

## 对象池设计

**NatsMsg 池** (`msg_pool.go`):
```go
msg := GetNatsMsg()
msg.Subject = subject
msg.Data = data
p.PublishMsg(msg.Msg)
msg.Release()  // 归还池
```

**Timer 池** (`timer_pool.go`):
```go
timer := acquireTimer(timeout)
defer releaseTimer(timer)
select {
case resp := <-ch:
    return resp.Data
case <-timer.C:
    // timeout
}
```

## 统计监控

开启 `is_stats: true` 后，每 30 秒输出：
- InMsgs / OutMsgs: 收发消息数
- InBytes / OutBytes: 收发字节
- Reconnects: 重连次数

## 项目约定

- 使用对象池 (`msg_pool.go`, `timer_pool.go`) 减少 GC
- 所有订阅在连接关闭时自动取消
- reqID 在 Reply Subject 级别唯一，非全局唯一

## 注意事项

- 连接池大小建议与并发请求量匹配
- `RequestSync` 超时后会清理等待 channel
- 重连时 NATS 自动恢复订阅，无需重新订阅
- Disconnect 回调会清理所有等待中的请求