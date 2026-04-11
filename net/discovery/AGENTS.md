# net/discovery - 节点发现服务

## 概述

提供节点注册、发现、监听功能，支持三种实现方式：default（配置文件）、nats（master-client模式）、etcd（第三方组件）。

## 目录结构

```
net/discovery/
├── discovery.go           # 发现服务注册管理
├── component.go           # 发现服务组件
├── discovery_default.go   # 默认实现（读取 profile 配置）
├── discovery_master.go    # NATS master-client 实现
└── discovery_etcd.go      # etcd 实现（引用 components/etcd）
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 注册发现服务 | `discovery.go:18` | `Register(discovery)` |
| 创建组件 | `component.go:18` | `New()` 创建组件 |
| 加载节点 | `discovery_default.go:28` | 从 profile 加载节点信息 |
| 获取成员 | `discovery_default.go:126` | `GetMember(nodeID)` |
| 按类型获取 | `discovery_default.go:86` | `ListByType(nodeType)` |
| 随机获取 | `discovery_default.go:103` | `Random(nodeType)` |
| 添加成员监听 | `discovery_default.go:164` | `OnAddMember(listener)` |

## 关键接口

**IDiscovery** (`facade/cluster.go:11-24`):
```go
type IDiscovery interface {
    Load(app IApplication)
    Name() string                                                 // 发现服务名称
    Map() map[string]IMember                                      // 获取成员列表
    ListByType(nodeType string, filterNodeID ...string) []IMember // 按类型获取列表
    Random(nodeType string) (IMember, bool)                       // 随机获取
    GetType(nodeID string) (nodeType string, err error)           // 获取节点类型
    GetMember(nodeID string) (member IMember, found bool)         // 获取成员
    AddMember(member IMember)                                     // 添加成员
    RemoveMember(nodeID string)                                   // 移除成员
    OnAddMember(listener MemberListener)                          // 添加监听
    OnRemoveMember(listener MemberListener)                       // 移除监听
    Stop()
}
```

## 三种实现方式

### 1. default（默认）

- 从 profile 配置文件读取节点信息
- 适用场景：开发测试、静态节点配置
- 配置位置：`profile.json` → `node` 字段
- 模式名称：`default`

### 2. nats（master-client）

- master 节点管理所有注册信息
- client 启动时向 master 注册
- 支持心跳检测、超时移除
- 适用场景：动态节点、生产环境
- 模式名称：`nats`
- 配置：
```json
{
  "cluster": {
    "discovery": {
      "mode": "nats"
    },
    "nats": {
      "master_node_id": "master-1"
    }
  }
}
```

### 3. etcd

- 基于 etcd 实现
- 实际代码在 `components/etcd` 仓库
- 模式名称：`etcd`

## 配置切换

通过 `cluster.discovery.mode` 切换：
```json
{
  "cluster": {
    "discovery": {
      "mode": "default"  // 或 "nats"、"etcd"
    }
  }
}
```

## 成员监听

```go
// 监听节点添加
app.Discovery().OnAddMember(func(member cfacade.IMember) {
    clog.Infof("新节点加入: %s", member.GetNodeID())
})

// 监听节点移除
app.Discovery().OnRemoveMember(func(member cfacade.IMember) {
    clog.Infof("节点离线: %s", member.GetNodeID())
})
```

## 项目约定

- 组件名称: `discovery_component`
- 自动注册 default 和 nats 实现（init 函数）
- IMember 使用 `cproto.Member` 实现

## 注意事项

- default 模式仅适合开发测试，不支持动态节点
- nats 模式需配置 `master_node_id`
- 心跳超时可通过节点 `__settings__.cluster_heartbeat_timeout` 配置（毫秒）