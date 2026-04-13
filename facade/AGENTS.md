# facade

## 角色

`facade` 定义框架对外契约。实现层大多在别的目录，协作时这里优先于局部实现描述。

- 如果文档和代码冲突，以 `facade` 当前接口为准
- 如果新增跨模块能力，优先先定接口，再落实现

## 关键接口

- [IApplication](./application.go:20)
- [IComponent](./component.go:4)
- [IActorSystem / IActor / IActorHandler](./actor.go:10)
- [IDiscovery / ICluster](./cluster.go:11)
- [INetParser](./net_parser.go:5)
- [IConnector](./connector.go:6)
- [ISerializer](./serializer.go:3)

协作时最容易误判的点：

- `INetParser` 当前真实职责是 `Load/AddConnector/Connectors`
- 它不是旧版的 `Package/Message/Route/Serialize` 风格接口

## 关键结构

- [Message](./message.go:12)：Actor 与 Cluster 之间的统一消息载体
- [ActorPath](./message.go:29)：路径格式为 `node.actor` 或 `node.actor.child`

## 联动检查

- 改 `application.go`：同步检查 `application.go`、`cherry.go`、根 `AGENTS.md`
- 改 `actor.go`：同步检查 `net/actor/`、`net/parser/`、`code/`
- 改 `cluster.go`：同步检查 `net/cluster/`、`net/discovery/`
- 改 `net_parser.go`：同步检查 `net/parser/`、`application.go`
- 改 `connector.go`：同步检查 `net/connector/`、`net/parser/`
- 改 `serializer.go`：同步检查 `net/serializer/`、`application.go`
- 改 `message.go`：同步检查 `net/actor/`、`net/cluster/`、`net/proto/`
