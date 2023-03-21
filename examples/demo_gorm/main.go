package main

import (
	"github.com/cherry-game/cherry"
	cherryGORM "github.com/cherry-game/cherry/components/gorm"
)

func main() {
	app := cherry.Configure(
		"./examples/config/profile-dev.json", //使用dev环境的配置
		"game-1",                             // 使用game-1 的节点id
		false,
		cherry.Standalone,
	)

	// 注册gorm组件，数据库具体配置请查看 config/profile-dev.json文件
	app.Register(cherryGORM.NewComponent())

	app.AddActors(
		&ActorDB{},
	)

	app.Startup()
}
