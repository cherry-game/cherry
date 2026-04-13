# net/parser

## 角色

`net/parser` 不是单纯协议编解码目录，而是前端接入层。它负责把 connector、新连接、session、agent actor 和路由处理串起来。

## 真实接口

- [INetParser](../../facade/net_parser.go:5)

当前仓库里的 parser 接口职责是：

- `Load(application IApplication)`
- `AddConnector(connector IConnector)`
- `Connectors() []IConnector`

不要再按旧版 `Package/Message/Route/Serialize` 理解它。

## 实现入口

- [pomelo/actor.go](./pomelo/actor.go:30)
- [simple/actor.go](./simple/actor.go:27)

## 真实运行链

标准接入方式：

1. 创建 parser
2. `AddConnector(...)`
3. `app.SetNetParser(parser)`
4. 启动应用后由 parser `Load()` 接管 connector

`Load()` 典型会做这些事：

- 检查 connector 是否已挂载
- 初始化协议或命令上下文
- 创建前端 `agentActor`
- 为每个 connector 注册 `OnConnect`
- 启动全部 connector

## 新连接处理

Pomelo 和 Simple 的默认入口都会：

- 创建 `Session`
- 生成 `sid`
- 构造 `Agent`
- 绑定到当前 parser actor 路径
- 启动 agent

## 常见坑

- `isFrontend == true` 时，应用启动前必须设置 parser
- parser 没有 connector 时，`Load()` 会直接 `panic`
- 绕过 parser 直接操作 connector，会失去默认的 session / agent 接入逻辑

## 联动检查

- 改 parser `Load()`：同步检查 `application.go`、`net/connector/`
- 改 agent 创建逻辑：同步检查 `net/proto/session.go`、`net/actor/`
- 改 Pomelo 路由或命令：同步检查 `net/proto/` 和客户端兼容性
- 改 `INetParser` 实现方式：同步检查 `facade/net_parser.go`
