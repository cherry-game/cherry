# 分布式多节点示例(保姆级教程)

- TODO

## 要求

- 安装GoLand >= 2021.1
- 安装GO版本 >= 1.17
- 安装nats.io中间件
- 本示例默认在windows环境调试(其他环境请自行查阅搭建nats的方法)

## 操作步骤

### 克隆

- git clone https://github.com/cherry-game/cherry.git
- 或者通过github下载源码的方式，点击`code`按钮`Download zip`下载源码包

### 调试

#### 打开项目

- 使用GoLand打开项目(Goland使用不熟练的请自行查阅资料)
- 找到`examples/game_cluster`目录

#### 启动nats

- 找到`run_nats.bat`，右键点击`Run cmd script`运行单机版`nats`
- 窗口显示`Listening for client connections on 0.0.0.0:4222` 代表nats启动成功，nats默认监听`4222`端口
- 正式环境请部署集群的nats!  正式环境请部署集群的nats!  正式环境请部署集群的nats!

#### 启动master节点

- 找到`exmaples/game_cluster/nodes/main.go`，该文件为本示例的main函数，所有项目都从此处启动
- 点击`func main() {`左边的绿色三角形，选择`Debug ...`运行本函数
- 此时master节点启动失败。从`Console`窗口可以看到需要传入参数
- 点击 `Edit Configurations...` 找到`Program arguments:`选项，配置参数为:`master --name=gc --node=gc-master`
- `master` 代表运行master节点，详情看`main.go:87`行
- `--name=gc` 代表选择`profile-gc.json`的环境配置参数
- `--node=gc-master` 代表master节点启动时，读取`profile-gc.json`文件中`"node_id": "gc-master"`的环境配置
- 点击`ok按钮`，继续启动调式。当`Console`窗口显示`application is running.`字样时，表示master节点启动成功。
- 如果`Console`窗口显示`nats connect fail! retrying in 3 seconds.` 则表示本地没有启动成功nats

#### 启动center节点

- 按照`master`的启动方式，再复制一个配置项 `Copy Configurations`，并配置启动参数
- 找到`Program arguments:`选项，配置参数为:`center --name=gc --node=gc-center`

#### 启动web节点

- 按照`master`的启动方式，再复制一个配置项 `Copy Configurations`，并配置启动参数
- 找到`Program arguments:`选项，配置参数为:`web --name=gc --node=gc-web-1`

#### 启动gate节点

- 按照`master`的启动方式，再复制一个配置项 `Copy Configurations`，并配置启动参数
- 找到`Program arguments:`选项，配置参数为:`gate --name=gc --node=gc-gate-1`

#### 启动game节点

- 按照`master`的启动方式，再复制一个配置项 `Copy Configurations`，并配置启动参数
- 找到`Program arguments:`选项，配置参数为:`game --name=gc --node=3000`

### 测试

- TODO

### 源码讲解

- TODO

### 配置

- TODO
