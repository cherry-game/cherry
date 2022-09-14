# 单节点聊天室示例

- 使用cherry引擎构建一个简单的多人聊天室程序
- 建议使用GoLand打开源码
- 本示例为h5客户端，使用`starx-wsclient`做为客户端sdk，连接类型为`websocket`，序列化类型为`json`

## 要求

- GO版本 >= 1.17

## 操作步骤

### 克隆

- git clone https://github.com/cherry-game/cherry.git
- 或通过github下面下载源码的方式。点击`code`按钮`Download zip`文件

### 调试

- 使用`GoLand`打开源码
- 找到`/examples/chat/`目录
- 双击打开`main.go`
- 点击第11行的绿色箭头，选择`Debug go build github.com/...`
- 程序调试已经启动，查看GoLand下面的`Console`面板，显示`application is running.`文字则表示成功

### 测试

- 从GoLand的`Console`面板可以看到，启动h5聊天室客户端的地址`http://127.0.0.1:80`
- 客户端连接服务器的websocket地址为`ws://:34590`
- 打开两个页面，在文本框中输入聊天内容并点击`send`按钮发送，两个页面将会收到聊天内容的广播

### 源码

- `web`目录，放置h5的聊天室客户端文件，包含html和js版的客户端协议
- `main.go` 主函数
- `protocol.go` 与客户端通信的协议结构(json方式)
- `room_handler.go` 房间handler处理
- `topic_service.go` 服务类
- `user_handler.go` 用户handler处理

### 配置

- 涉及的环境配置profile文件在 `/examples/config/profile-chat.json`
- `profile-chat.json`文件的注释通过`@xxx`表示，详见文件
