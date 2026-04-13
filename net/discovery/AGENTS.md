# net/discovery

## 角色

`net/discovery` 负责节点成员发现与成员表维护。当前核心实现是 `default` 和 `nats`。

## 真实入口

- [component.go](./component.go:13)
- [discovery.go](./discovery.go:18)

## 模式

配置入口见 [component.go:27](./component.go:27)：

- `default`
- `nats`
- `etcd`

协作时优先关注前两者；`etcd` 更像扩展点。

## 关键实现

- `default`：
  - 直接从 profile 读取节点表
  - 适合开发、测试、静态配置
  - 不做动态注册和心跳剔除
- `nats`：
  - master 维护成员表
  - client 向 master 注册并持续心跳
  - master 后台检查超时节点并广播移除事件

关键后台循环在 [discovery_master.go](./discovery_master.go)：

- `send2Master()`
- `heartbeatCheck()`

## 配置重点

- `cluster.discovery.mode`
- `cluster.nats.prefix`
- `cluster.nats.master_node_id`
- 当前节点 `__settings__.cluster_heartbeat_timeout`

## 局部约束

- `nats` 模式排障要同时看 discovery 和 NATS 连接状态
- 上层感知成员变化主要依赖 `OnAddMember` / `OnRemoveMember`
- 节点成员结构变更会影响 `facade/cluster.go` 和 `net/proto/`

## 常见坑

- `mode` 配错时 discovery 不会正常加载
- `nats` 模式必须配置 `master_node_id`
- `default` 模式不会动态感知新节点

## 联动检查

- 改 `component.go`：同步检查 `profile/`、`cherry.go`
- 改 `discovery_default.go`：同步检查 `profile/node.go`
- 改 `discovery_master.go`：同步检查 `net/nats/`、`net/proto/`、`profile/`
