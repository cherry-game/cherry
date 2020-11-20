package cherryComponents

import (
	"fmt"
	"github.com/json-iterator/go"
	"github.com/phantacix/cherry/const"
	"github.com/phantacix/cherry/interfaces"
	"github.com/phantacix/cherry/profile"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	connectFormat = "%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local"
)

type (
	ORMComponent struct {
		cherryInterfaces.BaseComponent

		// key:groupId,value:{key:id,value:*gorm.Db}
		ormMap map[string]map[string]*gorm.DB
	}

	mySqlConfig struct {
		Enable         bool
		GroupId        string
		Id             string
		DbName         string
		Host           string
		UserName       string
		Password       string
		MaxIdleConnect int
		MaXOpenConnect int
		LogMode        bool
	}
)

func NewORM() *ORMComponent {
	return &ORMComponent{
		ormMap: make(map[string]map[string]*gorm.DB),
	}
}

func parseMysqlConfig(item jsoniter.Any) *mySqlConfig {
	return &mySqlConfig{
		Enable:         item.Get("enable").ToBool(),
		GroupId:        item.Get("groupId").ToString(),
		Id:             item.Get("id").ToString(),
		DbName:         item.Get("dbName").ToString(),
		Host:           item.Get("host").ToString(),
		UserName:       item.Get("userName").ToString(),
		Password:       item.Get("password").ToString(),
		MaxIdleConnect: item.Get("maxIdleConnect").ToInt(),
		MaXOpenConnect: item.Get("maxOpenConnect").ToInt(),
		LogMode:        item.Get("logMode").ToBool(),
	}
}

func (s *ORMComponent) Name() string {
	return cherryConst.ORMComponent
}

func (s *ORMComponent) Init() {
	rootJson := cherryProfile.Config().Get("db")

	for i := 0; i < rootJson.Size(); i++ {
		item := rootJson.Get(i)
		cfg := parseMysqlConfig(item)

		db, err := s.createORM(cfg)
		if err != nil {
			panic(err)
		}

		dbs := s.ormMap[cfg.GroupId]
		if dbs == nil {
			dbs = make(map[string]*gorm.DB)
			s.ormMap[cfg.GroupId] = dbs
		}
		dbs[cfg.Id] = db
	}
}

func (s *ORMComponent) createORM(cfg *mySqlConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(connectFormat, cfg.UserName, cfg.Password, cfg.Host, cfg.DbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *ORMComponent) GetDb(id string) *gorm.DB {
	for _, group := range s.ormMap {
		for k, v := range group {
			if k == id {
				return v
			}
		}
	}
	return nil
}

func (s *ORMComponent) DBWithGroupId(groupId string) map[string]*gorm.DB {
	return s.ormMap[groupId]
}
