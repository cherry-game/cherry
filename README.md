# 欢迎使用cherry!
![cherry logo](https://img.shields.io/badge/cherry--game-cherry-red)
![cherry license](https://img.shields.io/github/license/cherry-game/cherry)
![go version](https://img.shields.io/github/go-mod/go-version/cherry-game/cherry)
![cherry tag](https://img.shields.io/github/v/tag/cherry-game/cherry)

- 这是一款分布式的golang游戏服务器框架
- 基于golang + nats.io + pomelo protocol技术构建
- 它具备高性能、可伸缩、分布式、协程分组管理等特点。并且上手简单、易学
- 让开发者更多的关注游戏业务，高效完成功能实现
- 文档陆续补充中，欢迎加入一起建设框架


# 讨论与交流
- QQ群:191651647
- 扫码加群

![qr-code](_docs/qq-qun.png)


# 核心功能

### 组件管理
- 基于组件的方式组合功能，方便统一管理生命周期
- 可根据需求自定义组件，并注册到框架，灵活扩展
- 可配置`cluster mode`和`standalone mode`


### 环境配置
- 可配置多个环境的参数，方便切换
- 所有系统参数、组件参数都基于profile文件配置，方便扩展
- 可根据业务需求自由的拆分或组装多个profile子文件，精简配置,拒绝冗余


### 日志
- 基于`uber zap`封装，性能良好
- 可配置多文件进行日志输出
- 基于`rotatelogs`增加滚动日志


### 消息&路由
- 包结构基本pomelo protocol实现
- 包解码编码
- 消息路由
- 消息序列化
- 事件


### 连接器
- tcp
- websocket
- http server
- http client
- kcp(未实现，以后做为组件集成)

### 集群&注册发现
- 读取`profile->node`配置文件的方式实现节点加载和同步(测试用)
- 启动`master`单节点的方式，基于nats.io通信的方式实现节点加载和同步(小规模用)
- etcd方式加载和同步节点(做为组件已实现)


# 扩展组件

### [cron组件](components/cron/README.md)
- 基于`github.com/robfig/cron/v3`进行封装成组件
- 性能良好

### [data-config组件](components/data-config/README.md)
- 策划配表读取管理组件
- 可基于本地配置文件的方式加载
- 可基于redis数据的方式加载
- 可基于接口抽像自定义数据源加载
- 支持自定义文件格式读取，目前已实现`JSON`格式读取
- 支持缓存热更新
- 可自定义类型检测
- 可根据`go-linq`进行数据集合的条件查询

### [etcd组件](components/etcd/README.md)
- 基于`etcd`组件进行封装，节点集群和注册发现


### [gin组件](components/gin/README.md)
- 集成`gin`组件，实现http server功能
- 自定义`controller`，增加`PreInit()`、`Init()`、`Stop()`初始周期的管理
- 增加几个常用的`middleware`组件
    - gin zap
    - recover with zap
    - cors跨域
    - max connect限流
- 封装了部份必用的参数获取函数


### [gorm组件](components/gorm/README.md)
- 集成`gorm`组件，实现mysql的数据库访问
- 支持多个mysql数据库配置和管理

### 待开放组件
- db队列
- mongodb组件
- gopher-lua脚本



