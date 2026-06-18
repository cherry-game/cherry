# Discovery 发现服务

集群节点自动发现与成员管理，支持多种后端实现。

## 架构

```
IDiscovery (业务接口)
├── ComponentDefault (基类)     — 成员存储 + listener 通知
│   ├── ComponentMaster (nats)  — NATS 主从模式
│   └── Component (etcd)        — etcd 分布式模式 (独立仓库)
```

| 模式 | Mode 值 | 适用场景 |
|------|---------|---------|
| `default` | 读取 profile 配置 | 单进程开发/测试 |
| `nats` | NATS 主从发现 | 多节点生产环境 |
| `etcd` | etcd lease + watch | 多节点生产环境，依赖 etcd |

## Install

```go
import cherryDiscovery "github.com/cherry-game/cherry/net/discovery"
```

## Quick Start

### default 模式（开发测试）

在 profile 中配置节点信息，无需额外部署：

```json
{
    "cluster": {
        "discovery": {
            "mode": "default"
        }
    },
    "node": {
        "game": [
            {
                "node_id": "game-1",
                "rpc_address": "127.0.0.1:10001",
                "__settings__": {
                    "region": "us-west"
                }
            }
        ],
        "gate": [
            {
                "node_id": "gate-1",
                "rpc_address": "127.0.0.1:20001"
            }
        ]
    }
}
```

### nats 模式（生产环境）

**启动 master 节点：**

master 负责接收注册、心跳检测、广播成员变更。

```json
{
    "cluster": {
        "discovery": {
            "mode": "nats"
        },
        "nats": {
            "prefix": "node",
            "master_node_id": "master-1"
        }
    }
}
```

**启动 worker 节点：**

worker 向 master 注册并定期发送心跳。

```json
{
    "cluster": {
        "discovery": {
            "mode": "nats"
        },
        "nats": {
            "prefix": "node",
            "master_node_id": "master-1"
        }
    }
}
```

`master_node_id` 必须与运行 master 的节点 `node_id` 一致。

## NATS 协议流程

```
1. 先启动 master 节点
2. Worker 启动 → NATS Request  → master 收到 register
3. Master 回复完整成员列表 + NATS Publish add 广播新成员
4. Worker 定期心跳 (1s) → master 心跳超时 (3s) → 移除 + RemoveMember 广播
5. Worker 退出 → NATS Publish remove → 所有节点移除该成员
```

NATS subject 格式：`cherry.<prefix>.discovery.<masterID>.<type>`

| Type | 方向 | 说明 |
|------|------|------|
| register | Worker→Master | 注册请求，携带自身 Member 数据 |
| add | Master→Worker | 新成员广播 |
| update | 双向 | Settings 变更广播 |
| remove | 双向 | 成员移除广播 |
| heartbeat | Worker→Master | 心跳，Master 更新 LastAt 或回复 registerRequired 触发重新注册 |

## 业务层 API (IDiscovery)

```go
discovery := app.Discovery()

// 成员查询
all := discovery.Map()                           // 所有成员
games := discovery.ListByType("game")            // 按类型过滤
games := discovery.ListByType("game", "self")    // 排除指定节点
member, ok := discovery.Random("gate")           // 随机选取
member, ok := discovery.GetMember("node-id")     // 按 ID 查找

// Settings 同步（更新当前节点，自动同步到其它节点）
discovery.UpdateSetting("region", "us-east")
discovery.UpdateSettings(map[string]string{"region": "us-east", "zone": "a"})

// 成员变更监听
discovery.OnAddMember(func(member IMember) {
    // 新节点加入
})
discovery.OnUpdateMember(func(member IMember) {
    // 节点 settings 变更
})
discovery.OnRemoveMember(func(member IMember) {
    // 节点离开
})
```

## 实现自定义后端

### 方式一：直接实现 IDiscovery

```go
type MyDiscovery struct {
    cfacade.Component
}

func (m *MyDiscovery) Mode() string { return "my-mode" }
func (m *MyDiscovery) Init() { /* ... */ }
// ... 实现 IDiscovery 全部方法
```

### 方式二：组合 ComponentDefault

复用 memberMap 存储和 listener 通知逻辑，只需关注传输层：

```go
type MyDiscovery struct {
    cherryDiscovery.ComponentDefault
    thisMember *cproto.Member
}

func (m *MyDiscovery) Mode() string { return "my-mode" }

func (m *MyDiscovery) Init() {
    m.thisMember = cherryDiscovery.NewMemberWithApp(m.App())
    m.AddMember(m.thisMember)
    // 初始化传输层...
}

func (m *MyDiscovery) UpdateSetting(key, value string) {
    m.thisMember.UpdateSetting(key, value)
    // 通过传输层广播变更...
}
```

注册：

```go
func init() {
    cherryDiscovery.Register(&MyDiscovery{})
}
```

## 配置参考

### cluster.discovery

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| mode | string | - | 发现服务模式：`default` / `nats` / `etcd` |

### cluster.nats (nats 模式)

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| prefix | string | "node" | NATS subject 前缀 |
| master_node_id | string | - | master 节点 ID（必须与运行的 master 节点 node_id 一致） |

### cluster.etcd (etcd 模式，独立仓库)

见 [components/etcd](https://github.com/cherry-game/components/tree/master/etcd)
