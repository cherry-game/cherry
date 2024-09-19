# 简单化版的网络数据包解析实现
- 起因是网友`@刘`在Q群中反馈，需要一个更简单的网络数据包结构。
- 本结构只是一个自定义网络数据包的演示。在实际开发中各位可根据自身需求进行定制。
- 代码未做详细测试，并且缺少对应的客户端实现。在使用中有任何问题欢迎反馈。


### 包结构
- 本结构参照`zinx`的网络包结构，通过`消息ID + 数据长度 + 数据`构建一个网络数据包
- MID uint32(4 bytes) +  DataLen uint32(4 bytes) + Data(n bytes)


### 使用方法
- 在网关节点构建一个simple的网络数据包解析器
- 通过`simple.AddNodeRoute(mid,&NodeRoute{...})`构造数据包路由策略
- [示例代码](https://github.com/cherry-game/examples/tree/master/demo_cluster/nodes/gate/gate.go)

### 示例代码
```
// 构建简单的网络数据包解析器
func buildSimpleParser(app *cherry.AppBuilder) cfacade.INetParser {
    agentActor := simple.NewActor("user")
    agentActor.AddConnector(cconnector.NewTCP(":10011"))
    agentActor.AddConnector(cconnector.NewWS(app.Address()))
	
        agentActor.SetOnNewAgent(func(newAgent *simple.Agent) {
        childActor := &ActorAgent{}
        //newAgent.AddOnClose(childActor.onSessionClose)
        agentActor.Child().Create(newAgent.SID(), childActor)
    })

    // 设置大头&小头
    agentActor.SetEndian(binary.LittleEndian)
    // 设置心跳时间
    agentActor.SetHeartbeatTime(60 * time.Second)
    // 设置积压消息数量
    agentActor.SetWriteBacklog(64)

    // 设置数据路由函数
    //agentActor.SetOnDataRoute(onSimpleDataRoute)

    // 设置消息节点路由(建议配合data-config配置表使用)
    // mid = 1 的消息路由到  gate节点.user的Actor.login函数上
    agentActor.AddNodeRoute(1, &simple.NodeRoute{
        NodeType: "gate",
        ActorID:  "user",
        FuncName: "login",
    })
	
    return agentActor
}
```