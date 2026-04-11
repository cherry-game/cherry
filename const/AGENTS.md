# const - 常量定义

## 概述

定义框架核心常量，包括版本号、Logo、路径分隔符等。

## 文件结构

```
const/
├── const.go      # 常量定义
└── const_test.go # 单元测试
```

## 核心常量

### 版本号

```go
const version = "1.4.19"

func Version() string {
    return version
}
```

### Logo

```go
var logo = `
░█████╗░██╗░░██╗███████╗██████╗░██████╗░██╗░░░██╗
██╔══██╗██║░░██║██╔════╝██╔══██╗██╔══██╗╚██╗░██╔╝
██║░░╚═╝███████║█████╗░░██████╔╝██████╔╝░╚████╔╝░
██║░░██╗██╔══██║██╔══╝░░██╔══██╗██╔══██╗░░╚██╔╝░░
╚█████╔╝██║░░██║███████╗██║░░██║██║░░██║░░░██║░░░
░╚════╝░╚═╝░░╚═╝╚══════╝╚═╝░░╚═╝╚═╝░░╚═╝░░░╚═╝░░░ 
game sever framework @v%s
`

func GetLOGO() string {
    return fmt.Sprintf(logo, Version())
}
```

### 路径分隔符

```go
const DOT = "." // ActorPath 的分隔符
```

## 使用示例

```go
import cconst "github.com/cherry-game/cherry/const"

// 获取版本号
version := cconst.Version()

// 获取 Logo
logo := cconst.GetLOGO()

// 构建路径
path := nodeID + cconst.DOT + actorID + cconst.DOT + childID
```

## Actor 路径格式

使用 `DOT` 分隔：

| 场景 | 格式 | 示例 |
|------|------|------|
| Actor 路径 | `{nodeID}.{actorID}` | `game-1.player` |
| 子 Actor 路径 | `{nodeID}.{actorID}.{childID}` | `game-1.player.room-1` |

## 项目约定

- 使用 `cconst` 作为包别名
- 路径分隔符统一使用 `DOT`
- 版本号在发布时更新

## 注意事项

- 版本号定义在 `const.go`，修改后需同步更新 Git Tag
- Logo 在应用启动时打印（通过 `GetLOGO()`）