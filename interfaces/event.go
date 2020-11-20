package cherryInterfaces

type IEvent interface {
	// event name
	EventName() string

	//unique id
	UniqueId() string
}
