# code - 返回码定义

## 概述

定义框架统一的返回码常量，用于 Actor 通信、RPC 调用、集群消息等场景。

## 文件结构

```
code/
└── code.go  # 返回码常量定义
```

## 返回码分类

### 基础返回码

| 码值 | 常量名 | 说明 |
|------|--------|------|
| 0 | `OK` | 成功 |
| 10 | `SessionUIDNotBind` | Session 未绑定 UID |
| 11 | `DiscoveryNotFoundNode` | 节点未在 Discovery 中注册 |
| 12 | `NodeRequestError` | 节点请求错误 |

### RPC 返回码

| 码值 | 常量名 | 说明 |
|------|--------|------|
| 20 | `RPCNetError` | RPC 网络错误 |
| 21 | `RPCUnmarshalError` | RPC 反序列化错误 |
| 22 | `RPCMarshalError` | RPC 序列化错误 |
| 23 | `RPCRemoteExecuteError` | RPC 远程执行错误 |

### Actor 返回码

| 码值 | 常量名 | 说明 |
|------|--------|------|
| 24 | `ActorPathIsNil` | Actor 路径为空 |
| 25 | `ActorFuncNameError` | Actor 函数名错误 |
| 26 | `ActorConvertPathError` | Actor 路径转换错误 |
| 27 | `ActorMarshalError` | Actor 参数序列化错误 |
| 28 | `ActorUnmarshalError` | Actor 参数反序列化错误 |
| 29 | `ActorInvokeResultIsNil` | Actor 调用结果为空 |
| 30 | `ActorSourceEqualTarget` | Actor 源与目标相同 |
| 31 | `ActorPublishRemoteError` | Actor 远程发布错误 |
| 32 | `ActorChildIDNotFound` | Actor 子 ID 未找到 |
| 33 | `ActorCallTimeout` | Actor 调用超时 |
| 34 | `ActorIDIsNil` | Actor ID 为空 |
| 35 | `ActorNotFound` | Actor 未找到 |
| 36 | `ActorInvokeRemoteError` | Actor 远程调用错误 |
| 37 | `ActorResponseIsError` | Actor 响应错误 |

## 辅助函数

```go
// 判断是否成功
func IsOK(code int32) bool {
    return code == OK
}

// 判断是否失败
func IsFail(code int32) bool {
    return code != OK
}
```

## 使用示例

```go
import ccode "github.com/cherry-game/cherry/code"

// Actor 调用
code := actor.Call(target, "OnMessage", args)
if ccode.IsFail(code) {
    clog.Errorf("调用失败: code = %d", code)
    return
}

// RPC 响应
rsp := &cproto.Response{
    Code: ccode.OK,
    Data: responseData,
}
```

## 项目约定

- 成功返回码: `OK = 0`
- 所有失败返回码 > 0
- 使用 `ccode.IsFail()` 判断失败
- 开发者可扩展自定义返回码（建议从 100 开始）

## 扩展建议

开发者可在业务层定义自定义返回码：

```go
// 业务返回码（建议从 100 开始）
const (
    PlayerNotFound    int32 = 100 // 玩家未找到
    ItemNotEnough     int32 = 101 // 物品不足
    RoomAlreadyExist  int32 = 102 // 房间已存在
)
```