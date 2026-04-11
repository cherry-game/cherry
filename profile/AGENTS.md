# profile - 配置管理

## 概述

承载项目所有参数配置，基于 jsoniter 实现 JSON 配置解析，支持多环境切换、配置文件合并、正则节点匹配。

## 目录结构

```
profile/
├── profile.go      # 配置初始化、加载、合并
├── node.go         # 节点信息结构（实现 INode 接口）
├── config.go       # Config 结构体（ProfileJSON 实现）
└── profile_test.go # 单元测试
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 初始化配置 | `profile.go:44` | `Init(filePath, nodeID)` 入口 |
| 读取配置 | `profile.go:80` | `GetConfig(path...)` |
| 加载文件 | `profile.go:84` | `LoadFile()` 支持 include 合并 |
| 获取节点 | `node.go:58` | `GetNodeWithConfig()` 解析节点配置 |
| 节点匹配 | `node.go:93` | 支持正则匹配 nodeID |

## 核心结构

**全局配置变量** (`profile.go:14-22`):
```go
cfg = &struct {
    profilePath string  // 配置根目录
    profileName string  // 配置文件名
    jsonConfig  *Config // JSON 配置对象
    env         string  // 环境名称
    debug       bool    // 调试模式
    printLevel  string  // 日志打印级别
}
```

**Node 结构体** (`node.go:13-20`):
```go
type Node struct {
    nodeID     string            // 节点 ID
    nodeType   string            // 节点类型
    address    string            // 对外地址
    rpcAddress string            // RPC 地址
    settings   cfacade.ProfileJSON // 节点专属配置
    enabled    bool              // 是否启用
}
```

**Config 结构体** (`config.go:11-14`):
```go
type Config struct {
    jsoniter.Any  // 继承 jsoniter.Any
}
```

## 配置文件格式

**主配置文件** (`profile-{name}.json`):
```json
{
  "env": "dev",
  "debug": true,
  "print_level": "debug",
  "include": ["cluster.json", "database.json"],
  "node": {
    "game": [
      {
        "node_id": "game-1",
        "address": "127.0.0.1:8080",
        "rpc_address": "127.0.0.1:9090",
        "enabled": true,
        "__settings__": {}
      }
    ]
  }
}
```

**include 文件合并规则**:
- include 文件先加载，主配置后合并
- 相同 key 时，主配置覆盖 include
- 嵌套 map 递归合并

## 项目约定

- 配置路径: `profile/settings/{env}/node.json`
- 环境切换通过 `env` 字段
- 节点 ID 支持正则匹配（如 `^game-\d+$`）
- 节点专属配置放在 `__settings__` 字段

## 特性说明

**正则节点匹配** (`node.go:112-123`):
- nodeID 以 `^` 开头、`$` 结尾视为正则
- 用于动态节点 ID 场景（如 `game-1`, `game-2`）

**配置合并** (`profile.go:114-130`):
- 递归合并嵌套 map
- 主配置优先级高于 include

## 使用示例

```go
// 初始化配置
node, err := cprofile.Init("profile/settings/dev/node.json", "game-1")

// 读取配置
dbConfig := cprofile.GetConfig("database")
dbHost := dbConfig.GetString("host", "localhost")

// 获取节点信息
nodeID := node.NodeID()
nodeType := node.NodeType()
settings := node.Settings()
```

## 注意事项

- 配置文件必须包含 `node` 字段
- include 文件路径相对于主配置文件
- `__settings__` 用于节点专属配置，避免污染全局