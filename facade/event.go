package cherryFacade

// IEvent 事件接口
type IEvent interface {
	Name() string     // 事件名称
	UniqueId() string // 事件唯一id
}

// EventFunc 事件注册函数
type EventFunc func(e IEvent)
