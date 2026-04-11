# logger - Zap 日志封装

## 概述

基于 uber-go/zap 封装的高性能日志库，支持控制台输出、文件输出、日志切割（RotateLogs）、软链接等功能。

## 目录结构

```
logger/
├── logger.go          # 日志核心实现（CherryLogger）
├── logger_config.go   # 日志配置结构
├── logger_test.go     # 单元测试
└── rotatelogs/        # 日志切割实现
│   ├── rotatelogs.go  # 核心切割逻辑
│   ├── interface.go   # 接口定义（RotateLogs, Clock, Handler）
│   ├── options.go     # 配置选项函数
│   └── event.go       # 事件定义（FileRotatedEvent）
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 初始化节点日志 | `logger.go:40` | `SetNodeLogger(node)` |
| 创建日志实例 | `logger.go:67` | `NewLogger(refLoggerName)` |
| 创建切割日志 | `logger.go:130` | `rotatelogs.New()` |
| 日志配置 | `logger_config.go:47` | `NewConfig(jsonConfig)` |
| 获取日志级别 | `logger.go:300` | `GetLevel(level)` |
| 刷新日志 | `logger.go:59` | `Flush()` |

## 日志输出函数

```go
// 格式化输出
clog.Debug(args...)
clog.Info(args...)
clog.Warn(args...)
clog.Error(args...)
clog.Panic(args...)
clog.Fatal(args...)

// 模板输出
clog.Debugf(template, args...)
clog.Infof(template, args...)
clog.Warnf(template, args...)
clog.Errorf(template, args...)
clog.Panicf(template, args...)
clog.Fatalf(template, args...)

// 带上下文输出
clog.Infow(msg, keyAndValues...)
clog.Errorw(msg, keyAndValues...)
```

## 日志级别

| 级别 | 说明 |
|------|------|
| `debug` | 调试信息 |
| `info` | 常规信息 |
| `warn` | 警告信息 |
| `error` | 错误信息 |
| `panic` | 严重错误，触发 panic |
| `fatal` | 致命错误，调用 os.Exit |

## RotateLogs 日志切割

### 概述

RotateLogs 实现日志文件自动切割，避免单个日志文件过大，支持按时间、按大小切割，自动清理过期日志。

### 核心特性

**按时间切割** (`WithRotationTime`):
- 默认每 24 小时切割一次
- 支持自定义切割间隔

**按大小切割** (`WithRotationSize`):
- 当文件大小达到指定阈值时切割
- 单位: bytes

**最大保留时间** (`WithMaxAge`):
- 默认保留 7 天
- 超过时间的日志自动删除

**最大保留数量** (`WithRotationCount`):
- 保留指定数量的日志文件
- 与 MaxAge 互斥，只能设置其一

**软链接** (`WithLinkName`):
- 创建指向当前日志文件的软链接
- 方便实时查看最新日志

### 使用示例

```go
// 创建切割日志
hook, err := rotatelogs.New(
    "logs/debug_%Y%m%d%H%M.log",  // 文件名格式（strftime）
    rotatelogs.WithLinkName("logs/debug.log"),     // 软链接路径
    rotatelogs.WithMaxAge(7 * 24 * time.Hour),     // 保留 7 天
    rotatelogs.WithRotationTime(24 * time.Hour),   // 每天切割
)
```

### 文件名格式

使用 strftime 格式：
- `%Y`: 年（4位）
- `%m`: 月（01-12）
- `%d`: 日（01-31）
- `%H`: 时（00-23）
- `%M`: 分（00-59）

示例: `logs/debug_202604101530.log`

### 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithLinkName(path)` | 软链接路径 | "" |
| `WithMaxAge(duration)` | 最大保留时间 | 7天 |
| `WithRotationTime(duration)` | 切割间隔 | 24小时 |
| `WithRotationSize(bytes)` | 按大小切割 | 0（禁用） |
| `WithRotationCount(n)` | 保留文件数 | 0（禁用） |
| `WithClock(clock)` | 时钟（UTC/Local） | Local |
| `WithHandler(handler)` | 切割事件回调 | nil |
| `ForceNewFile()` | 强制创建新文件 | false |

### 事件回调

```go
rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(e Event) {
    if e.Type() == rotatelogs.FileRotatedEventType {
        event := e.(*rotatelogs.FileRotatedEvent)
        fmt.Printf("日志切割: %s -> %s\n", event.PreviousFile(), event.CurrentFile())
    }
}))
```

## 配置文件格式

```json
{
  "logger": {
    "debug": {
      "level": "debug",
      "stack_level": "error",
      "enable_console": true,
      "enable_write_file": false,
      "max_age": 7,
      "time_format": "15:04:05.000",
      "print_caller": true,
      "rotation_time": 86400,
      "file_link_path": "logs/debug.log",
      "file_path_format": "logs/debug_%Y%m%d%H%M.log"
    }
  }
}
```

## Config 结构

```go
type Config struct {
    LogLevel        string  // 输出日志等级
    StackLevel      string  // 堆栈输出日志等级
    EnableConsole   bool    // 是否控制台输出
    EnableWriteFile bool    // 是否输出文件
    MaxAge          int     // 最大保留天数
    TimeFormat      string  // 时间输出格式
    PrintCaller     bool    // 是否打印调用函数
    RotationTime    int     // 日期分割时间（秒）
    FileLinkPath    string  // 日志文件软链接路径
    FilePathFormat  string  // 日志文件路径格式
    IncludeStdout   bool    // 是否包含 stdout 输出
    IncludeStderr   bool    // 是否包含 stderr 输出
}
```

## 文件名变量替换

支持在文件路径中使用变量：
- `%nodeid`: 节点 ID
- `%nodetype`: 节点类型

```go
SetFileNameVar("nodeid", node.NodeID())
SetFileNameVar("nodetype", node.NodeType())

// 配置
"file_path_format": "logs/%nodetype/%nodeid_%Y%m%d.log"
// 结果: logs/game/game-1_20260410.log
```

## 项目约定

- 使用 `clog` 作为包别名调用日志函数
- 默认使用控制台输出，需配置 `ref_logger` 启用文件输出
- 日志切割默认每天执行，文件保留 7 天

## 注意事项

- `MaxAge` 和 `RotationCount` 不能同时设置
- 日志切割时自动创建目录
- 删除过期日志在后台 goroutine 执行
- 软链接使用相对路径（同目录时）