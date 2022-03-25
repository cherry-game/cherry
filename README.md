# cherry
- 这是一款分布式的go游戏服务器框架。 
- 基于golang + nats.io + pomelo protocol技术构建。
- 它具备高性能、可伸缩、分布式、协程分组管理等特点。并且上手简单、易学。 
- 让开发者更多的关注游戏业务，高效完成功能实现。
- 文档陆续补充中。。。


# TODO

### 基础
- 多环境profile配置


### 日志
- 多文件配置输出
- 过滤配置
- LEVEL定义
- 滚动日志


### 消息&路由
- 包结构(pomelo)
- 包解码编码
- 消息路由
- 消息序列化
- 事件
- 定时器


### 网络协议
- tcp
- websocket
- http server
- http client
- kcp


### 数据配表
- 本地加载配表
- 第三方数据源加载配表(redis)
- 热更新配表
- 类型检测
- 条件查询(go-linq)


### 集群
- 文件配置方式加载节点
- etcd方式加载&更新节点
- nats.io

### 其他
- mysql db队列
- nat消息队列
- gopher-lua脚本


