# net/serializer - 序列化器

## 概述

提供 JSON 和 Protobuf 两种序列化实现，用于消息编码解码。

## 目录结构

```
net/serializer/
├── json.go      # JSON 序列化实现（基于 jsoniter）
├── protobuf.go  # Protobuf 序列化实现
```

## 查找指南

| 任务 | 位置 | 说明 |
|------|----------|------|
| 创建 JSON 序列化器 | `json.go:9` | `NewJSON()` |
| 创建 Protobuf 序列化器 | `protobuf.go:12` | `NewProtobuf()` |
| JSON 序列化 | `json.go:14` | `Marshal(v)` |
| JSON 反序列化 | `json.go:24` | `Unmarshal(data, v)` |
| Protobuf 序列化 | `protobuf.go:17` | `Marshal(v)` |
| Protobuf 反序列化 | `protobuf.go:31` | `Unmarshal(data, v)` |

## 关键接口

**ISerializer** (`facade/serializer.go`):
```go
type ISerializer interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
    Name() string
}
```

## 两种实现

### JSON (`json.go`)

- 基于 `jsoniter` 高性能 JSON 库
- 若输入已是 `[]byte`，直接返回
- 适用场景：调试、Web 客户端通信

```go
serializer := cserializer.NewJSON()
data, err := serializer.Marshal(myData)
```

### Protobuf (`protobuf.go`)

- 基于 `google.golang.org/protobuf`
- 输入必须是 `proto.Message` 类型
- 适用场景：高性能、跨节点 RPC

```go
serializer := cserializer.NewProtobuf()
data, err := serializer.Marshal(&cproto.ClusterPacket{...})
```

## 默认配置

应用默认使用 Protobuf：
```go
// application.go:63
app := &Application{
    serializer: cserializer.NewProtobuf(),
    ...
}
```

可通过 `SetSerializer()` 切换：
```go
appBuilder.SetSerializer(cserializer.NewJSON())
```

## 序列化器名称

| 序列化器 | Name() 返回值 |
|---------|--------------|
| JSON | `"json"` |
| Protobuf | `"protobuf"` |

## 项目约定

- Protobuf 序列化要求类型实现 `proto.Message`
- JSON 使用 jsoniter（比标准库更快）
- 应用启动日志会显示序列化器名称

## 错误处理

Protobuf 序列化时类型错误返回：
```go
cerr.ProtobufWrongValueType  // "Convert on wrong type value"
```

## 注意事项

- Protobuf 序列化非 `proto.Message` 类型会报错
- 切换序列化器需在 `Startup()` 之前