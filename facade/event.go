package cherryFacade

// IEvent 事件接口
type IEvent interface {
	Name() string     // 事件名称
	UniqueId() string // 事件唯一id
}

// EventFn 事件注册函数
type EventFn func(e IEvent)
