# error - 错误定义

## 概述

定义框架统一的错误类型和错误构造函数，用于路由、数据包、消息、集群、Actor 等场景。

## 文件结构

```
error/
└── error.go  # 错误定义和构造函数
```

## 错误构造函数

```go
// 创建错误
func Error(text string) error

// 格式化创建错误
func Errorf(format string, a ...interface{}) error

// 包装错误
func Wrap(err error, text string) error

// 格式化包装错误
func Wrapf(err error, format string, a ...interface{}) error
```

## 预定义错误

### 路由错误

| 常量名 | 说明 |
|--------|------|
| `RouteFieldCantEmpty` | 路由字段不能为空 |
| `RouteInvalid` | 路由无效 |

### 数据包错误

| 常量名 | 说明 |
|--------|------|
| `PacketWrongType` | 数据包类型错误 |
| `PacketSizeExceed` | 数据包大小超限 |
| `PacketConnectClosed` | 连接已关闭 |
| `PacketInvalidHeader` | Header 无效 |
| `PacketMsgSmallerThanExpected` | 接收数据小于预期 |

### 消息错误

| 常量名 | 说明 |
|--------|------|
| `MessageWrongType` | 消息类型错误 |
| `MessageInvalid` | 消息无效 |
| `MessageRouteNotFound` | 路由未在字典中找到 |

### Protobuf 错误

| 常量名 | 说明 |
|--------|------|
| `ProtobufWrongValueType` | Protobuf 值类型转换错误 |

### 集群错误

| 常量名 | 说明 |
|--------|------|
| `ClusterClientIsStop` | 集群客户端已停止 |
| `ClusterRequestTimeout` | 集群请求超时 |
| `ClusterPacketMarshalFail` | 集群数据包序列化失败 |
| `ClusterPacketUnmarshalFail` | 集群数据包反序列化失败 |
| `ClusterPublishFail` | 集群发布失败 |
| `ClusterRequestFail` | 集群请求失败 |
| `ClusterNodeTypeIsNil` | 集群节点类型为空 |
| `ClusterNodeTypeMemberNotFound` | 集群节点类型成员未找到 |

### 发现服务错误

| 常量名 | 说明 |
|--------|------|
| `DiscoveryNotFoundNode` | 节点未在 Discovery 中找到 |

### Actor 错误

| 常量名 | 说明 |
|--------|------|
| `ActorPathError` | Actor 路径错误 |

### 函数错误

| 常量名 | 说明 |
|--------|------|
| `FuncIsNil` | 函数为空 |
| `FuncTypeError` | 函数类型错误 |

## 使用示例

```go
import cerr "github.com/cherry-game/cherry/error"

// 创建错误
err := cerr.Error("invalid parameter")

// 格式化错误
err := cerr.Errorf("node not found: %s", nodeID)

// 包装错误
err := cerr.Wrap(originalErr, "process failed")

// 判断预定义错误
if err == cerr.ClusterRequestTimeout {
    // 处理超时
}
```

## 项目约定

- 使用 `cerr` 作为包别名
- 使用 `Wrap` 包装底层错误
- 预定义错误使用 `Error()` 创建（支持 `==` 比较）

## 扩展建议

开发者可在业务层定义自定义错误：

```go
// 业务错误
var (
    PlayerNotFoundError = cerr.Error("player not found")
    ItemNotFoundError   = cerr.Error("item not found")
)
```