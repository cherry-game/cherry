package cherryFacade

// IEvent 事件接口
type IEvent interface {
	Name() string    // 事件名称
	UniqueId() int64 // 事件唯一id
}

// EventFn 事件注册函数
type EventFn func(e IEvent)

// EventInfo 事件注册信息
type EventInfo struct {
	QueueHash QueueHashFn
	List      []EventFn
}
