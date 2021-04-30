package cherryInterfaces

// IEvent 事件接口
type IEvent interface {
	EventName() string // 事件名称
	UniqueId() string  // 事件唯一id(hash用)
}
