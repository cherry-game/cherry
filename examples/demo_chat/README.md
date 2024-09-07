# 多节点精简版聊天室示例

- 使用cherry引擎构建一个简单的多人聊天室程序
- 推荐使用GoLand IDE
- 本示例为h5客户端，使用`pomelo-ws-client`做为客户端sdk，连接类型为`websocket`，序列化类型为`json`

## 要求

- GO版本 >= 1.18

未使用过 Golang 的开发者，请参考 [环境安装与配置](../../_docs/env-setup.md) 进行准备工作。

## 操作步骤

### 克隆

- git clone https://github.com/cherry-game/cherry.git
- 或通过github下面下载源码的方式。点击`code`按钮`Download zip`文件

### 用 GoLand 开发调试 - 推荐

- 找到`/examples/chat/`目录
- 启动 nats 服务
  - Windows：`3rd/nat-server/run_nats.bat`
  - Mac：`nats-server`
- 启动 master 节点`master/main.go`
- 启动 log 节点`log/main.go`
- 启动 room 节点`room/main.go`
- 访问`http://127.0.0.1:端口号`

### 用 Visual Studio Code 开发调试

- 打开`/examples/chat/`目录。
- 安装必要的插件和依赖
- 配置`launch.json`
- 启动 nats 服务
  - Windows：`3rd/nat-server/run_nats.bat`
  - Mac：`nats-server`
- 启动 `chat-master`
- 启动 `chat-log`
- 启动 `chat-room`
- 访问`http://127.0.0.1:8081`  
  - 端口以 room 进程打印的具体的值为准，如果发现端口被占用，请搜索并替换。

### 测试

- 从 GoLand 或者 Visual Sutdio Code 的`Console`面板可以看到，启动 H5 聊天室客户端的地址 `http://127.0.0.1:80`(不一定是 80，以实际端口为准)
- 客户端连接服务器的 Websocket 地址为`ws://:34590`
- 打开两个页面，在文本框中输入聊天内容并点击`send`按钮发送，两个页面将会收到聊天内容的广播

### 源码目录说明

- `log` 目录为日志服节点(演示actor收发消息)
- `master` 发现服务, master节点
- `protocol` 通信协议,json序列化
- `room` 房间服节点(为了测试方便，包含了网关和web客户端)
- `static`目录为h5客户端静态文件，包含html和js版的客户端协议

### 配置

- 涉及的环境配置profile文件在 `/examples/config/profile-chat.json`
- `profile-chat.json`文件的注释通过`@xxx`表示，详见文件

### 关于actor model的使用

- 从`room/main.go`文件可得知，节点启动时通过`pomelo.NewActor("user")`创建了一个`user actor`. 该`actor`用于管理客户端连接.
- 通过`app.AddActors(...)`可得知，注册了`room`actor，用于房间管理
- 如果需要创建多个聊天房间，可以通过room的子actor实现