package cherryHandler

type IExecutor interface {
	Invoke()
	String() string
}
