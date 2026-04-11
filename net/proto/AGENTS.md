# net/proto - Protobuf 协议定义

## 概述

定义跨节点通信的 Protobuf 消息结构，包括 Member、ClusterPacket、Session、Response 等。

## 目录结构

```
net/proto/
├── proto.proto          # Protobuf 源定义（编辑此文件）
├── proto.pb.go          # 自动生成代码（禁止编辑）
├── proto.go             # 扩展方法（Member.IsTimeout）
├── cluster_packet.go    # ClusterPacket 对象池、序列化
├── session.go           # Session 扩展方法
├── build.bat            # Windows 编译脚本
└── README.md            # 编译说明
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 编辑协议 | `proto.proto` | 定义消息结构 |
| 编译协议 | `Makefile` 或 `build.bat` | `make protoc` |
| 获取 ClusterPacket | `cluster_packet.go:19` | `GetClusterPacket()` 从对象池获取 |
| 回收 ClusterPacket | `cluster_packet.go:39` | `Recycle()` 归还对象池 |
| 解析数据包 | `cluster_packet.go:25` | `UnmarshalPacket(data)` |
| Session 判断绑定 | `session.go:12` | `IsBind()` 判断是否绑定 UID |

## 核心消息结构

**Member** - 节点成员信息:
```protobuf
message Member {
  string nodeID = 1;           // 节点 ID
  string nodeType = 2;         // 节点类型
  string address = 3;          // RPC 地址
  map<string, string> settings = 4;  // 节点设置
  int64 lastAt = 5;            // 最后心跳时间
  int64 heartbeatTimeout = 6;  // 心跳超时（毫秒）
}
```

**ClusterPacket** - 跨节点通信包:
```protobuf
message ClusterPacket {
  int64 buildTime = 1;      // 构建时间戳
  string sourcePath = 2;    // 源 Actor 路径
  string targetPath = 3;    // 目标 Actor 路径
  string funcName = 4;      // 调用函数名
  bytes argBytes = 5;       // 参数序列化数据
  Session session = 6;      // 会话信息
}
```

**Session** - 用户会话:
```protobuf
message Session {
  string sid = 1;           // Session ID
  int64 uid = 2;            // 用户 ID
  string agentPath = 3;     // 前端 Agent 路径
  string ip = 4;            // IP 地址
  map<string, string> data = 7;  // 扩展数据
}
```

**Response** - RPC 响应:
```protobuf
message Response {
  int32 code = 1;           // 返回码
  bytes data = 2;           // 响应数据
}
```

## Pomelo 相关消息

**PomeloResponse** - Pomelo 响应:
```protobuf
message PomeloResponse {
  string sid = 1;
  uint32 mid = 2;           // 客户端消息 ID
  bytes data = 3;
  int32 code = 4;
}
```

**PomeloPush** - Pomelo 推送:
```protobuf
message PomeloPush {
  string sid = 1;
  int64 uid = 2;
  string route = 3;         // 路由路径
  bytes data = 4;
}
```

**PomeloBroadcast** - Pomelo 广播:
```protobuf
message PomeloBroadcast {
  PushType pushType = 1;    // 广播类型
  repeated int64 uidList = 2;
  string route = 3;
  bytes data = 4;
}
```

## 对象池使用

```go
// 获取 ClusterPacket
packet := cproto.GetClusterPacket()
packet.SourcePath = "node1.actor1"
packet.TargetPath = "node2.actor2"
packet.FuncName = "OnMessage"

// 使用后回收
packet.Recycle()
```

## Session 扩展方法

```go
// 判断是否绑定用户
session.IsBind()  // uid > 0

// Actor 路径
session.ActorPath()  // agentPath + "." + sid

// 数据操作
session.Set("key", "value")
session.GetString("key")
session.GetMID()  // 获取消息 ID
```

## 编译协议

```bash
# 安装 protoc-gen-go
make init

# 编译
make protoc

# 或 Windows 下执行
build.bat
```

## 项目约定

- 编辑 `proto.proto`，禁止编辑 `proto.pb.go`
- ClusterPacket 使用对象池减少 GC
- 使用 `Recycle()` 回收对象

## 反模式

- **禁止编辑** `proto.pb.go` - protoc 自动生成文件