package cherryORM

import (
	"fmt"
	cconst "github.com/cherry-game/cherry/const"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	goSqlDriver "github.com/go-sql-driver/mysql"
	"github.com/json-iterator/go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

const (
	connectFormat = "%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local"
)

type (
	Component struct {
		cfacade.Component
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

	// HashDb hash by group id
	HashDb func(dbMaps map[string]*gorm.DB) string
)

func NewComponent() *Component {
	return &Component{
		ormMap: make(map[string]map[string]*gorm.DB),
	}
}

func (s *Component) Name() string {
	return cconst.ORMComponent
}

func parseMysqlConfig(groupId string, item jsoniter.Any) *mySqlConfig {
	return &mySqlConfig{
		GroupId:        groupId,
		Id:             item.Get("db_id").ToString(),
		DbName:         item.Get("db_name").ToString(),
		Host:           item.Get("host").ToString(),
		UserName:       item.Get("user_name").ToString(),
		Password:       item.Get("password").ToString(),
		MaxIdleConnect: item.Get("max_idle_connect").ToInt(),
		MaXOpenConnect: item.Get("max_open_connect").ToInt(),
		LogMode:        item.Get("log_mode").ToBool(),
		Enable:         item.Get("enable").ToBool(),
	}
}

func (s *Component) Init() {
	// load only the database contained in the `db_id_list`
	dbIdList := s.App().Settings().Get("db_id_list")
	if dbIdList.LastError() != nil || dbIdList.Size() < 1 {
		clog.Warnf("[nodeId = %s] `db_id_list` property not exists.", s.NodeId())
		return
	}

	dbConfig := cprofile.GetConfig("db")
	if dbConfig.LastError() != nil {
		panic("`db` property not exists in profile file.")
	}

	for _, groupId := range dbConfig.Keys() {
		s.ormMap[groupId] = make(map[string]*gorm.DB)

		dbGroup := dbConfig.Get(groupId)

		for i := 0; i < dbGroup.Size(); i++ {
			item := dbGroup.Get(i)
			mysqlConfig := parseMysqlConfig(groupId, item)

			for j := 0; j < dbIdList.Size(); j++ {
				dbId := dbIdList.Get(j).ToString()
				if mysqlConfig.Id != dbId {
					continue
				}

				if mysqlConfig.Enable == false {
					panic(fmt.Sprintf("[dbName = %s] is disabled!", mysqlConfig.DbName))
				}

				db, err := s.createORM(mysqlConfig)
				if err != nil {
					panic(fmt.Sprintf("[dbName = %s] create orm fail. error = %s", mysqlConfig.DbName, err))
				}

				s.ormMap[groupId][mysqlConfig.Id] = db
				clog.Infof("[dbGroup =%s, dbName = %s] is connected.", mysqlConfig.GroupId, mysqlConfig.Id)
			}
		}
	}

	goSqlDriver.SetLogger(clog.DefaultLogger)
}

func (s *Component) createORM(cfg *mySqlConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(connectFormat, cfg.UserName, cfg.Password, cfg.Host, cfg.DbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: getLogger(),
	})

	if err != nil {
		return nil, err
	}

	if cfg.LogMode {
		return db.Debug(), nil
	}

	return db, nil
}

func getLogger() logger.Interface {
	return logger.New(
		ormLogger{log: clog.DefaultLogger},
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      true,
		},
	)
}

func (s *Component) GetDb(id string) *gorm.DB {
	for _, group := range s.ormMap {
		for k, v := range group {
			if k == id {
				return v
			}
		}
	}
	return nil
}

func (s *Component) GetHashDb(groupId string, hashFn HashDb) (*gorm.DB, bool) {
	dbGroup, found := s.GetDbMap(groupId)
	if found == false {
		clog.Warnf("groupId = %s not found.", groupId)
		return nil, false
	}

	dbId := hashFn(dbGroup)
	db, found := dbGroup[dbId]
	return db, found
}

func (s *Component) GetDbMap(groupId string) (map[string]*gorm.DB, bool) {
	dbGroup, found := s.ormMap[groupId]
	return dbGroup, found
}
