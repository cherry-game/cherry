# 多节点精简版聊天室示例

- 使用cherry引擎构建一个简单的多人聊天室程序
- 推荐使用GoLand IDE
- 本示例为h5客户端，使用`pomelo-ws-client`做为客户端sdk，连接类型为`websocket`，序列化类型为`json`

## 要求

- GO版本 >= 1.17

## 操作步骤

### 克隆

- git clone https://github.com/cherry-game/cherry.git
- 或通过github下面下载源码的方式。点击`code`按钮`Download zip`文件

### 调试

- 推荐使用`GoLand`打开源码
- 找到`/examples/chat/`目录
- 启动nats服务`3rd/nat-server/run_nats.bat`
- 启动master节点`master/main.go`
- 启动log节点`log/main.go`
- 启动room节点`room/main.go`
- 访问`http://127.0.0.1:80`

### 测试

- 从GoLand的`Console`面板可以看到，启动h5聊天室客户端的地址`http://127.0.0.1:80`
- 客户端连接服务器的websocket地址为`ws://:34590`
- 打开两个页面，在文本框中输入聊天内容并点击`send`按钮发送，两个页面将会收到聊天内容的广播

### 源码

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