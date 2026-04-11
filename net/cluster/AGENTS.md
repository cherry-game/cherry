# net/cluster - 集群通信组件

## 概述

基于 NATS 的集群 RPC 通信实现，支持本地消息、远程消息、节点类型广播等多种发布模式。

## 目录结构

```
net/cluster/
├── component.go       # 集群组件定义
└── nats_cluster/
    ├── cluster.go     # NATS 集群实现（核心）
    └── const.go       # Subject 生成函数
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 创建集群组件 | `component.go:17` | `New()` 创建组件 |
| 初始化集群 | `nats_cluster/cluster.go:52` | `Init()` 加载配置、订阅消息 |
| 发布本地消息 | `nats_cluster/cluster.go:159` | `PublishLocal(nodeID, packet)` |
| 发布远程消息 | `nats_cluster/cluster.go:197` | `PublishRemote(nodeID, packet)` |
| 按类型广播 | `nats_cluster/cluster.go:235` | `PublishRemoteType(nodeType, packet)` |
| 同步请求 | `nats_cluster/cluster.go:271` | `RequestRemote(nodeID, packet, timeout)` |

## 关键接口

**ICluster** (`facade/cluster.go:37-44`):
```go
type ICluster interface {
    Init()                                                                                               // 初始化
    PublishLocal(nodeID string, packet *cproto.ClusterPacket) error                                      // 发布本地消息
    PublishRemote(nodeID string, packet *cproto.ClusterPacket) error                                     // 发布远程消息
    PublishRemoteType(nodeType string, cpacket *cproto.ClusterPacket) error                              // 按节点类型广播
    RequestRemote(nodeID string, packet *cproto.ClusterPacket, timeout ...time.Duration) ([]byte, int32) // 同步请求
    Stop()                                                                                               // 停止
}
```

## 消息发布模式

| 方法 | 用途 | Subject 格式 |
|------|------|-------------|
| `PublishLocal` | 发送到目标节点的本地队列 | `{prefix}.{nodeType}.local.{nodeID}` |
| `PublishRemote` | 发送到目标节点的远程队列 | `{prefix}.{nodeType}.remote.{nodeID}` |
| `PublishRemoteType` | 广播到某类型所有节点 | `{prefix}.{nodeType}.remote.type` |
| `RequestRemote` | 同步 RPC，等待响应 | 使用 reply subject |

## Subject 订阅

集群启动时订阅三种 Subject：
- **localSubject**: 接收本地消息，投递到 ActorSystem 的 Local 队列
- **remoteSubject**: 接收远程消息，投递到 ActorSystem 的 Remote 队列
- **remoteTypeSubject**: 接收类型广播消息

## 配置要求

需在 profile 中配置 `cluster.nats`:
```json
{
  "cluster": {
    "nats": {
      "prefix": "node",
      "address": "nats://127.0.0.1:4222",
      "pool_size": 1,
      "request_timeout": 2
    }
  }
}
```

## 项目约定

- 所有发布方法在完成后调用 `packet.Recycle()` 回收对象
- 通过 Discovery 服务获取目标节点的 nodeType
- 使用 protobuf 序列化 ClusterPacket

## 注意事项

- 调用 `PublishRemote` 前需确保目标节点在 Discovery 中已注册
- `RequestRemote` 超时默认 2 秒，可传入自定义 timeout
- 组件名称: `cluster_component`