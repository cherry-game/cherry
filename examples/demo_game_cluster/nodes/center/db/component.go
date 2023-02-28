package db

import (
	cherryUtils "github.com/cherry-game/cherry/extend/utils"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
)

var (
	onLoadFuncList []func() // db初始化时加载函数列表
)

type Component struct {
	cherryFacade.Component
}

func (c *Component) Name() string {
	return "db_center_component"
}

// Init 组件初始化函数
// 为了简化部署的复杂性，本示例取消了数据库连接相关的逻辑
func (c *Component) Init() {
}

func (c *Component) OnAfterInit() {
	addOnload(loadDevAccount)

	for _, fn := range onLoadFuncList {
		cherryUtils.Try(fn, func(errString string) {
			cherryLogger.Warnf(errString)
		})
	}
}

func (*Component) OnStop() {
	//组件停止时触发逻辑
}

func New() *Component {
	return &Component{}
}

func addOnload(fn func()) {
	onLoadFuncList = append(onLoadFuncList, fn)
}
