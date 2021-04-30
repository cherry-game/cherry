package cherryInterfaces

//IComponent 组件接口
type IComponent interface {
	IAppContext   // IAppContext实例上下文对象
	Name() string // 组件唯一名称
	Init()        // 初始化
	AfterInit()   // 初始化后执行
	BeforeStop()  // 开始停止
	Stop()        // 停止
}

// BaseComponent
type BaseComponent struct {
	AppContext
}

func (*BaseComponent) Name() string {
	return ""
}

func (*BaseComponent) Init() {
}

func (*BaseComponent) AfterInit() {
}

func (*BaseComponent) BeforeStop() {
}

func (*BaseComponent) Stop() {
}
