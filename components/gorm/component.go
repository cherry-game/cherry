package cherryGORM

import (
	"fmt"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

const (
	Name          = "gorm_component"
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
		MaxOpenConnect int
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
	return Name
}

func parseMysqlConfig(groupId string, item cfacade.JsonConfig) *mySqlConfig {
	return &mySqlConfig{
		GroupId:        groupId,
		Id:             item.GetString("db_id"),
		DbName:         item.GetString("db_name"),
		Host:           item.GetString("host"),
		UserName:       item.GetString("user_name"),
		Password:       item.GetString("password"),
		MaxIdleConnect: item.GetInt("max_idle_connect", 4),
		MaxOpenConnect: item.GetInt("max_open_connect", 8),
		LogMode:        item.GetBool("log_mode", true),
		Enable:         item.GetBool("enable", true),
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

		dbGroup := dbConfig.GetConfig(groupId)
		for i := 0; i < dbGroup.Size(); i++ {
			item := dbGroup.GetConfig(i)
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

	if err := mysqlDriver.SetLogger(clog.DefaultLogger); err != nil {
		clog.Warnf("mysql driver setLogger error = %v", err)
	}
}

func (s *Component) createORM(cfg *mySqlConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(connectFormat, cfg.UserName, cfg.Password, cfg.Host, cfg.DbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: getLogger(),
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnect)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnect)
	sqlDB.SetConnMaxLifetime(time.Minute)

	if cfg.LogMode {
		return db.Debug(), nil
	}

	return db, nil
}

func getLogger() logger.Interface {
	return logger.New(
		gormLogger{log: clog.DefaultLogger},
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
