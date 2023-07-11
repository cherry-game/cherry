package main

import (
	cherryGORM "github.com/cherry-game/cherry/components/gorm"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	"gorm.io/gorm"
	"time"
)

type ActorDB struct {
	cactor.Base
	centerDB *gorm.DB
}

func (p *ActorDB) AliasID() string {
	return "db"
}

// OnInit Actor初始化前触发该函数
func (p *ActorDB) OnInit() {
	// db配置的注解
	// 打开profile-dev.json，找到"game-1"和"db"配置
	// 当前示例启动的节点id为 game-1
	// db_id_list参数配置了center_db_1，表示当前节点可以连接该数据库
	// 当前节点启时注册了gorm组件  app.Register(cherryGORM.NewComponent())
	// 通过gorm组件可以获取对应的gorm.DB对象
	// 后续操作请参考gorm的用法

	// 获取gorm组件
	gorm := p.App().Find(cherryGORM.Name).(*cherryGORM.Component)
	if gorm == nil {
		clog.DPanicf("[component = %s] not found.", cherryGORM.Name)
	}

	// 获取 db_id = "center_db_1" 的配置
	centerDbID := p.App().Settings().GetConfig("db_id_list").GetString("center_db_id")
	p.centerDB = gorm.GetDb(centerDbID)
	if p.centerDB == nil {
		clog.Panic("center_db_1 not found")
	}

	// 每秒查询一次db
	p.Timer().Add(5*time.Second, p.selectDB)
	// 1秒后进行一次分页查询
	p.Timer().AddOnce(1*time.Second, p.selectPagination)
}

func (p *ActorDB) selectDB() {
	userBindTable := &UserBindTable{}
	tx := p.centerDB.First(userBindTable)
	if tx.Error != nil {
		clog.Warn(tx.Error)
	}

	clog.Infof("%+v", userBindTable)
}

func (p *ActorDB) selectPagination() {
	list, count := p.pagination(1, 10)
	clog.Infof("count = %d", count)

	for _, table := range list {
		clog.Infof("%+v", table)
	}
}

// pagination 分页查询
func (p *ActorDB) pagination(page, pageSize int) ([]*UserBindTable, int64) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	var list []*UserBindTable
	var count int64

	p.centerDB.Model(&UserBindTable{}).Count(&count)

	if count > 0 {
		list = make([]*UserBindTable, pageSize)
		s := p.centerDB.Limit(pageSize).Offset((page - 1) * pageSize)
		if err := s.Find(&list).Error; err != nil {
			clog.Warn(err)
		}
	}

	return list, count
}
