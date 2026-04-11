# extend - 扩展工具库

## 概述

通用工具库集合，提供时间处理、定时器、队列、ID生成、加密、序列化等功能。

## 目录结构

```
extend/
├── time/           # 时间处理（CherryTime 自定义类型）
├── time_wheel/     # 时间轮定时器实现
├── utils/          # 通用工具（Try, Panic 捕获）
├── queue/          # 消息队列
├── nuid/           # NATS NUID 实现
├── snowflake/      # Snowflake ID 生成
├── sync/           # 同步工具（Limit）
├── string/         # 字符串工具
├── slice/          # 切片工具
├── regex/          # 正则缓存
├── reflect/        # 反射工具（FuncInfo）
├── map/            # Map 工具（StringAnyMap）
├── mapstructure/   # 结构体映射（mitchellh/mapstructure fork）
├── json/           # JSON 工具
├── gob/            # Gob 序列化
├── file/           # 文件工具
├── crypto/         # 加密工具
├── compress/       # 压缩工具
├── base58/         # Base58 编码
├── http/           # HTTP Client
├── net/            # 网络工具
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|-------|
| 时间格式化 | `time/time_to.go` | `ToDateTimeFormat()` |
| 时间轮定时器 | `time_wheel/timer.go` | 高性能定时器 |
| Try-Panic 捕获 | `utils/utils.go` | `Try(fn, errFn)` |
| Snowflake ID | `snowflake/snowflake.go` | 分布式 ID 生成 |
| 反射调用 | `reflect/func.go` | `FuncInfo` 函数信息提取 |
| Map 映射 | `mapstructure/mapstructure.go` | 配置解析到结构体 |

## 核心工具

**CherryTime** (`time/time.go`):
- 自定义时间类型，支持多种格式输出
- `Now()`, `ToMillisecond()`, `ToDateTimeFormat()`
- 时间差计算 `NowDiffMillisecond()`

**Try** (`utils/utils.go`):
```go
cutils.Try(func() {
    // 可能 panic 的代码
}, func(errString string) {
    // 错误处理
})
```

**TimeWheel** (`time_wheel/`):
- 基于时间轮的高性能定时器
- `AddTimer()`, `RemoveTimer()`

**Snowflake** (`snowflake/`):
- 分布式唯一 ID 生成
- 默认实现 `snowflake_default.go`

## 项目约定

- 工具函数以包名前缀调用: `ctime.Now()`, `cutils.Try()`
- CherryTime 是自定义类型，非标准 `time.Time`
- mapstructure fork 用于 Profile 配置解析

## 反模式

- 避免直接使用 `time.Time`，使用 `ctime.CherryTime`
- mapstructure 保持原始 API 兼容性