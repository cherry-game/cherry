package cherryFacade

//IComponent 组件接口
type IComponent interface {
	IAppContext    // IAppContext实例上下文对象
	Name() string  // 组件唯一名称
	Init()         // 初始化
	OnAfterInit()  // 初始化后执行
	OnBeforeStop() // 开始停止
	OnStop()       // 停止
}

// Component
type Component struct {
	AppContext
}

func (*Component) Name() string {
	return ""
}

func (*Component) Init() {
}

func (*Component) OnAfterInit() {
}

func (*Component) OnBeforeStop() {
}

func (*Component) OnStop() {
}
