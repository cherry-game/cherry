# net/nats

## 角色

`net/nats` 是框架底层 NATS 客户端封装，负责连接池、请求响应、订阅管理、对象池和部分并发清理。

## 真实入口

- [connect_pool.go](./connect_pool.go:19)
- [connect.go](./connect.go:20)

## 全局状态

这里不是实例化对象模型，而是包级全局：

- `connectPool`
- `connectSize`
- `reconnectDelay`
- `requestTimeout`

这意味着同进程默认共享一个连接池，重新初始化时要考虑旧状态残留。

## 关键行为

- `Connect()`：
  - 首次连接失败会按固定间隔重试
  - 成功后初始化 reply 订阅
  - 如果开启统计，会启动独立 goroutine
- `Request()`：
  - 直接使用 NATS 自带请求
- `RequestSync()`：
  - 在本连接内分配唯一 `reqID`
  - 把 waiter 登记到 `waiters`
  - 通过 reply subject 发请求
  - 等待响应、超时或断线清理
- `Subscribe()`：
  - 会记录到 `subs`
  - 便于关闭时统一取消订阅

## 并发约束

- `waiters.LoadAndDelete` 用来避免重复关闭 channel
- 断线、关闭、超时都会尝试清理 waiter
- `Connect.Conn` 在首次成功连接后保持稳定；运行期重连由 `nats.Conn` 自己处理
- `mu` 主要保护本 wrapper 的局部可变状态，如 `subs` 和 `stopStats`
- 不要在持有 `mu` 时调用 `Subscribe()`，避免重入死锁

## 对象池

- `msg_pool.go` 复用 `nats.Msg`
- `timer_pool.go` 复用 `time.Timer`

如果改请求逻辑，要一起检查对象池释放路径。

## 常见坑

- `RequestSync` 超时后，晚到响应会变成无主响应
- 断线时 `waiters` 会被清空，上层调用方要自己考虑重试
- reply subject 是每个连接一个，不是全局唯一队列
- `CloseConnectPool()` 不会自动重置所有包级变量

## 联动检查

- 改 `connect_pool.go`：同步检查 `net/cluster/`、`net/discovery/`
- 改 `connect.go`：同步检查 `RequestSync` 调用方和超时语义
- 改 reply subject：同步检查 `net/cluster/nats_cluster/cluster.go`
- 改对象池：同步检查 `msg_pool.go`、`timer_pool.go`
## 2026-04-13 mutex note

- `Connect` now uses `sync.Mutex` instead of `sync.RWMutex`.
- Hot paths such as `Request()` and `RequestSync()` should continue to use the stable `*nats.Conn` pointer directly and must not take `mu`.
- `mu` is only for wrapper-owned mutable state like `subs` and `stopStats`.
