# profile

## 角色

`profile` 负责读取配置文件、解析节点、合并 include，并把结果挂到全局配置状态上。

## 真实入口

- [profile.go](./profile.go:44)
- [node.go](./node.go:58)
- [config.go](./config.go:10)

## 全局状态

`profile` 不是纯函数库，它维护包级状态，例如：

- `profilePath`
- `profileName`
- `jsonConfig`
- `env`
- `debug`
- `printLevel`

因此同进程多配置实例要格外小心。

## 初始化流程

1. `Init(filePath, nodeID)`
2. 校验配置文件路径
3. `LoadFile()` 读取主配置
4. 读取 include 并合并
5. 从 `node` 节点中找匹配的 `nodeID`
6. 填充全局 `cfg`

## 关键语义

- include 规则：
  - 先读 include
  - 再用主配置覆盖
  - map 类型递归合并
- 节点匹配支持：
  - 单个字符串
  - 字符串数组
  - 以 `^` 开头、`$` 结尾的正则表达式字符串
- 节点私有运行参数放在 `__settings__`
- `Config.GetDuration()` 只是把整数转成 `time.Duration`
  - 单位通常由调用方再乘 `time.Second` 等常量决定

## 常见坑

- `node` 节点缺失会导致初始化失败
- include 路径是相对主配置文件目录
- `GetDuration()` 本身不带单位语义
- 全局 `cfg` 会让同进程多配置实例互相影响

## 联动检查

- 改 `profile.go`：同步检查 `application.go`、`logger/`、`net/discovery/`
- 改 `node.go`：同步检查 `facade/application.go` 和节点配置样式
- 改 `config.go`：同步检查所有 `GetDuration/GetInt/GetString` 调用点
