# 分布式多节点示例

- 建议在windows环境进行调试(示例自带`nats-server.exe`)，其他操作系统需自行搭建nats server
- 本示例没有使用数据库，进程重启会还原所有数据
- 客户端演示分为两种：
    - `robot_client` 为go实现的游戏压测客户端，使用`tcp/protobuf`协议
    - `nodes/web/view/` 为h5实现的游戏客户端，使用`websocket/protobuf`协议
- 欢迎开发者一起入群讨论，构建更好的demo

## 要求

- 安装GO版本 >= 1.17
- 安装nats.io >= 2.0

## 配置

- profile文件在`examples/config/profile-gc.json`
- 策划配置文件在 `examples/config/data/`

## 操作步骤

### 克隆

- `git clone https://github.com/cherry-game/cherry.git`
- 或者点击github.com页面的`code`按钮`Download zip`下载源码包

### 调试

#### 0x00 打开项目

> 打开项目源码，找到`examples/game_cluster`目录

#### 0x01 启动nats server

> nats为高性能的分布式消息中间件，详情可通过`https://github.com/nats-io/nats-server` 进行了解 <br />
> 本框架中所有节点都基于nats进行消息通信 <br />
> 单机版nats执行程序在`examples/3rd/nats-server`目录中 <br />
> 正式环境请使用集群nats-streaming-server进行部署 `https://github.com/nats-io/nats-streaming-server`

- 操作步骤:
- 运行`examples/3rd/nats-server/run_nats.bat`单机版
- 窗口显示`Listening for client connections on 0.0.0.0:4222` 代表nats启动成功，nats默认监听`4222`端口

#### 0x02 启动参数配置

> 找到`exmaples/game_cluster/nodes/main.go`，所有节点都从`main.go`启动 <br />
> 以下操作的启动参数配置以`goland`开发工具为例 <br />
> 附:`vs code`启动参数配置 [launch.json](launch.json) <br />

#### 0x03 启动master节点

> master节点主要用于实现最基础的发现服务,基于nats构建 <br />
> 正式环境也可配置为etcd方式提供发现服务 <br />
> 相关的代码在`examples/game_cluster/master/`目录

- 操作步骤:
    - 在`Program arguments:`选项填入参数:`master --path=./examples/config/profile-gc.json --node=gc-master`

#### 0x04 启动center节点

> center节点目前主要用于处理帐号相关的业务或全局唯一的业务

- 操作步骤:
    - 在`Program arguments:`选项填入参数:`center --path=./examples/config/profile-gc.json --node=gc-center`

#### 0x05 启动web节点

> web节点主要对外提供一些http的接口，可横向扩展，多节点部署 <br />
> 目前用于开发者帐号注册、区服列表、sdk登陆/支付回调、验证token生成等业务

- 操作步骤:
  - 在`Program arguments:`选项填入参数:`web --path=./examples/config/profile-gc.json --node=gc-web-1`

#### 0x06 启动gate节点

> gate节点为游戏对外网关，可横向扩展，多节点部署 <br />
> 主要用于管理客户端的连接、消息路由与转发

- 操作步骤:
    - 在`Program arguments:`选项填入参数:`gate --path=./examples/config/profile-gc.json --node=gc-gate-1`

#### 0x07 启动game节点

> game节点为具体的游戏逻辑业务，根据业务需求可多节点部署 <br />
> 在分服的游戏中可提供游戏内的各种逻辑实现

- 操作步骤:
    - 在`Program arguments:`选项填入参数:`game --path=./examples/config/profile-gc.json --node=10001`

### 测试

- 使用go实现客户端，通过tcp协议连接gate网关进行压力测试
- 使用h5实现客户端，通过websocket协议连接gate网关进行功能的展示

#### 启动压测机器人

- 找到`examples/game_cluster/robot_client/main.go` 文件,并执行
- 机器人执行逻辑为：`注册帐号`，`登陆获取token`、`连接网关`、`用户登录游戏服`、`查看角色`、`创建角色`、`进入角色`
- 默认设定为创建1000个帐号，可自行调整`maxRobotNum`参数进行测试
- 执行完成后，从game节点的`Console`可以查看到`onlineCount = 10000`字样，表示1万帐号已经进入游戏

#### 启动h5客户端

- 直接访问`http://127.0.0.1`，按照界面步骤提示操作

### 源码讲解

- `internal` 内部业务逻辑
    - `code` 定义一些业务的状态码
    - `component` 组件目录，
        - `check_center`组件, 用于在启动前节点先检查`center`节点是否已启动
    - `constant` 一些常用定义
    - `data` 策划配表包装的struct，用于读取`../../config/data`目录的策划配表
    - `event` 游戏事件
    - `guid` 生成全局id
    - `pb` protobuf生成的协议结构
    - `protocol` protobuf结构定义目录
    - `rpc` 跨节点rpc函数封装
    - `session_key` 一些session相关的常量定义
    - `token` 登录token逻辑，包含生成token、验证token
    - `types` 各种自定义类型封装,方便struct从配置文件、数据库读取数据时进行序列化、反序列化

- `nodes` 分布式节点目录
    - `center`节点
    - `game` 节点
    - `gate` 节点
    - `master` 节点
    - `web` 节点(为了演示方便，包含了h5客户端)
- `robot_client` 压测机器人(tcp/protobuf协议)
- `build_protocol.bat` 生成protobuf结构代码到`internal/pb/`目录

### 运行截图

![screenshot](screenshot.png)