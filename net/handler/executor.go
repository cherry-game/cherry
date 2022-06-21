package cherryHandler

type IExecutor interface {
	Index() int         // execute goroutine index
	SetIndex(index int) // set goroutine index
	Invoke()            // invoke method
	String() string     // string
}
