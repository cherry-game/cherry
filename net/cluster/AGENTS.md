# net/cluster

## 角色

`net/cluster` 是框架集群组件层，负责把 `ICluster` 契约接到 NATS 实现上。

## 真实入口

- [component.go](./component.go:12)
- [nats_cluster/cluster.go](./nats_cluster/cluster.go:18)

## 生命周期

1. `cluster_component.Init()` 创建 cluster 实例
2. 读取配置、初始化 NATS 连接池、订阅 subject
3. 停机时 `OnStop()` 调用 `Stop()`

## 关键语义

- `PublishLocal()`：
  - 会先通过 discovery 查目标节点的 `nodeType`
  - 再编码 `ClusterPacket` 并投递到目标 local subject
- `PublishRemote()`：
  - 与 `PublishLocal()` 类似，但投递到 remote subject
- `PublishRemoteType()`：
  - 按节点类型广播
  - 广播前会确认该类型在 discovery 中至少有成员
- `RequestRemote()`：
  - 返回 `[]byte` 和 `int32 code`
  - 调用方要自己继续反序列化

## 局部约束

- 对外发布接口通常都会 `defer cpacket.Recycle()`
- 调用后不要继续持有同一个 `ClusterPacket`
- `nodeType` 不是调用方直接传入，而是经 discovery 查询得到
- 底层传输协议固定为 protobuf

## 常见坑

- `PublishRemote` 失败时，常见根因其实是 discovery 里没有对应节点
- `RequestRemote` 超时不一定是网络问题，也可能是远端 Actor 没处理完
- 复用已经发过的 `ClusterPacket` 会踩对象池复用问题

## 联动检查

- 改 `component.go`：同步检查 `cherry.go`、`application.go`
- 改 `nats_cluster/cluster.go`：同步检查 `net/nats/`、`net/discovery/`、`net/proto/`、`code/`、`error/`
- 改 subject 生成规则：同步检查 `net/discovery/discovery_master.go`
