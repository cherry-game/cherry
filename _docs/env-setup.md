# Cherry 环境安装与配置

## 安装 NATS

cherry 使用了 [nats.io](https://nats.io/) 作为消息中间件。

### Windows

Windows 用户直接前往 [https://nats.io/](https://nats.io/) 下载并安装。

### Mac

Mac 用户使用 `brew install nats` 进行安装。

## 安装 Go

前往 [https://golang.google.cn/dl/](https://golang.google.cn/dl/) 安装 Go 语言。

如果是 Mac 用户，可以使用 `brew install go` 进行安装。

安装完成后，使用 `go version` 即可检查是否安装成功，以及查看对应的 Go 版本。

## 设置代理（可选）

`go get` 命令在执行的时候容易出现 `timeout`，通过设置代理可以解决。

在控制台中输入以下脚本可以解决：

```sh
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/
```

可通过 `go env` 命令查看是否设置成功。


## 安装编程环境

### GoLand

推荐使用 [GoLand](https://www.jetbrains.com/go/) 进行项目开发，不管是编程还是调试都非常方便。

### Visual Studio Code

如果想使用 Visual Studio Code 进行开发则需要进行如下配置。

- 1、安装 Go 语言插件：Tooltitude for Go
- 2、Shift + CTRL/CMD + P -> Go: Install/Update Tools
- 3、配置项目的 launch.json （具体配置以具体项目为准）

以下为 demo_chat/.vscode/launch.json 的配置：

```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name":"chat-master",
            "type":"go",
            "request":"launch",
            "mode":"debug",
            "program":"${workspaceFolder}/examples/demo_chat/master",
            "internalConsoleOptions": "openOnSessionStart",
        },
        {
            "name":"chat-log",
            "type":"go",
            "request":"launch",
            "mode":"debug",
            "program":"${workspaceFolder}/examples/demo_chat/log",
            "internalConsoleOptions": "openOnSessionStart",
        },
        {
            "name":"chat-room",
            "type":"go",
            "request":"launch",
            "mode":"debug",
            "program":"${workspaceFolder}/examples/demo_chat/room",
            "internalConsoleOptions": "openOnSessionStart",
        }
    ],
}
```
