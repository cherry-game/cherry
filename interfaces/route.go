package cherryInterfaces

type IRoute interface {
	NodeType() string
	HandlerName() string
	Method() string
}

type RouteFunction func(session ISession, packet interface{}, ctx IApplication) error
