# cherry环境配置

## 安装Go
任选一种安装方式：
- 官方安装包 https://golang.google.cn/dl/
- 国内镜像包 https://studygolang.com/dl
- brew安装  `brew install go`
- Go多版本管理工具 https://github.com/voidint/g

设置代理：
- 在命令行输入 `go env -w GOPROXY=https://goproxy.cn,direct`

> 建议安装1.18+以上的版本 \
> \
> 输入命令 `go version` 确认是否安装成功 \
> \
> 输入命令 `go env` 查看环境参数

## 安装nats
任选一种安装方式：
- 官方安装指南 https://docs.nats.io/running-a-nats-service/introduction/installation
- [推荐]docker安装 https://docs.nats.io/running-a-nats-service/introduction/installation#installing-via-docker
- 安装包下载 https://github.com/nats-io/nats-server/releases
- brew安装 `brew install nats-server`
- [推荐]项目附带的windows版本 `examples/3rd/nats-server`

> 启动后,窗口显示`Listening for client connections on 0.0.0.0:4222` 代表`nats`启动成功，默认监听`4222`端口 \
> \
> 正式环境请使用nats集群部署

## 开发工具准备
优先使用熟悉的开发工具

#### Visual Studio Code
- 免费、可配置性
-  安装Go工具集，按信`Shift + CTRL/CMD + P` 输入 `Go: Install/Update Tools`, 勾选所有项，点击确认
- 点击侧方栏`扩展`进入插件安装：
  - 输入`Go`，安装`Go Team at Google`发布的插件
  - 输入`Tooltitude for Go`，安装`Tooltitude Team`发布的插件
- Debug配置(launch.json)：
  - 参考[demo_chat] `examples/demo_chat/launch.json`
  - 参考[demo_game_cluster] `examples/demo_game_cluster/launch.json`

#### Goland
- 收费，易用
- 推荐使用[GoLand](https://www.jetbrains.com/go/) 进行项目开发，不管是编码还是调试都非常方便。