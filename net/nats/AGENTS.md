# net/nats

## 角色

`net/nats` 是框架底层的 NATS 客户端封装，负责：

- 连接池初始化与轮询获取
- 请求 / 响应收发
- reply subject 订阅
- 订阅列表管理
- waiter 清理
- `nats.Msg` 和 `time.Timer` 的对象池复用

## 真实入口

- [connect_pool.go](./connect_pool.go:11)
- [connect.go](./connect.go:20)

## 当前状态模型

这里是“连接池 + 单连接实例状态”的组合，不再把所有运行参数都放在包级全局。

包级状态当前只有：

- `connectPool`
- `roundIndex`

每个 `Connect` 实例自己持有：

- `options.address`
- `options.reconnectDelay`
- `options.requestTimeout`
- `options.maxReconnects`
- `options.user/password`
- `options.isStats`
- `options.statsInterval`
- `subs`
- `stopStats`
- `waiters`
- `reply`

因此现在要区分两类状态：

- 连接池级别：由 [NewConnectPool](./connect_pool.go:16) 和 [CloseConnectPool](./connect_pool.go:68) 管理
- 单连接级别：由 [Connect](./connect.go:60) / [Close](./connect.go:104) 管理

## 关键行为

- `NewConnectPool()`：
  - 先执行 `resetConnectPool()`，避免旧连接池残留
  - 从配置读取 `address`、认证、重连、请求超时、统计开关、统计周期
  - 为每个连接通过 `OptionFunc` 注入实例级参数
  - `isConnect == true` 时立即连接全部 `Connect`
- `GetConnect()`：
  - 通过 `roundIndex` 做轮询选择
  - 连接池为空时返回 `nil`
- `CloseConnectPool()`：
  - 会关闭现有连接
  - 会清空 `connectPool`
  - 会重置 `roundIndex`

- `Connect.Connect()`：
  - 首次连接失败会循环重试
  - 成功后设置稳定的 `*nats.Conn`
  - 初始化 reply 订阅
  - 如果启用统计，则启动独立 goroutine
- `Connect.Close()`：
  - 停止统计协程
  - 取消当前连接记录过的全部订阅
  - 关闭底层 `nats.Conn`
  - 清空全部 waiter

- `Request()`：
  - 直接使用 `nats.Conn.Request`
- `RequestSync()`：
  - 为当前连接分配唯一 `reqID`
  - 把 waiter 注册到 `waiters`
  - 通过当前连接自己的 reply subject 发请求
  - 等待响应、超时，或在断线 / 关闭时被清理
- `Subscribe()` / `QueueSubscribe()`：
  - 创建订阅后会记录到 `subs`
  - 便于 `Close()` 统一取消

## 并发约束

- `Connect.mu` 当前是 `sync.Mutex`，不是 `sync.RWMutex`
- `mu` 只保护 wrapper 自己的局部可变状态，例如：
  - `subs`
  - `stopStats`
  - `Conn` 首次赋值时的竞争收敛
- 热路径如 `Request()`、`RequestSync()`、`Subscribe()` 先读取稳定的 `*nats.Conn` 指针，再执行 NATS 调用，不应把整个请求过程放到 `mu` 里
- reply 回包使用 `waiters.LoadAndDelete` 接管 waiter，避免重复关闭 channel
- 超时、断线、显式关闭都会触发 waiter 清理
- 不要在持有 `mu` 时再调用会回到订阅路径的方法，避免重入问题

## reply / waiter 语义

- reply subject 是“每个 `Connect` 一个”，格式见 [NewConnect](./connect.go:45)
- 它不是整个进程唯一队列
- `RequestSync()` 超时后，晚到响应会因为 waiter 已删除而被丢弃
- 断线和 `ClosedHandler` 都会执行 `clearWaiters()`，上层如果依赖幂等请求，要自己考虑重试语义

## 配置重点

连接池初始化主要读取这些配置：

- `address`
- `user`
- `password`
- `max_reconnects`
- `pool_size`
- `is_stats`
- `stats_interval`
- `reconnect_delay`
- `request_timeout`

当前语义：

- `stats_interval` 读取后按“秒”转换成 `time.Duration`
- `reconnect_delay` 通过 `GetDuration()` 读整数，再乘 `time.Second`
- `request_timeout` 通过 `GetDuration()` 读整数，再乘 `time.Second`

默认值要注意两层：

- 连接池初始化时，`reconnect_delay` 默认 1 秒，`request_timeout` 默认 1 秒，`stats_interval` 默认 30 秒
- 如果某个 `Connect.options` 最终仍为 `<= 0`，实例方法会再兜底：
  - `ReconnectDelay()` 默认 1 秒
  - `RequestTimeout()` 默认 2 秒
  - `StatsInterval()` 默认 30 秒

## 对象池

- [msg_pool.go](./msg_pool.go) 复用 `nats.Msg`
- [timer_pool.go](./timer_pool.go) 复用 `time.Timer`

改请求 / 超时链路时，要一起检查：

- 消息对象是否及时 `Release()`
- timer 在复用前是否保持 stopped + drained 状态

## 常见坑

- 不要再按旧实现理解 `reconnectDelay` / `requestTimeout` 为包级全局状态；它们现在属于每个 `Connect.options`
- `CloseConnectPool()` 现在会重置连接池状态，不是只关闭连接
- reply subject 是单连接级别，不是全局唯一
- `RequestSync()` 超时不代表 NATS 一定断了，也可能只是远端处理过慢
- 断线后 waiter 会被清空，调用方要自己决定是否重试
- 统计协程是否运行取决于 `is_stats`

## 联动检查

- 改 [connect_pool.go](./connect_pool.go:11)：同步检查 `net/cluster/`、`net/discovery/`
- 改 [connect.go](./connect.go:20)：同步检查 `RequestSync` 调用方和超时语义
- 改 reply subject 规则：同步检查 `../cluster/nats_cluster/cluster.go`
- 改对象池：同步检查 [msg_pool.go](./msg_pool.go) 和 [timer_pool.go](./timer_pool.go)
