package cherryGORM

import (
	"fmt"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	Name = "gorm_component"
	dsn  = "%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local"
)

type (
	Component struct {
		cfacade.Component
		// key:groupID,value:{key:id,value:*gorm.Db}
		ormMap map[string]map[string]*gorm.DB
	}

	mySqlConfig struct {
		Enable         bool
		GroupID        string
		ID             string
		DbName         string
		Host           string
		UserName       string
		Password       string
		MaxIdleConnect int
		MaxOpenConnect int
		LogMode        bool
		DSN            string
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

func parseMysqlConfig(groupID string, item cfacade.ProfileJSON) *mySqlConfig {
	return &mySqlConfig{
		GroupID:        groupID,
		ID:             item.GetString("db_id"),
		DSN:            item.GetString("dsn", ""),
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
	dbIDList := s.App().Settings().Get("db_id_list")
	if dbIDList.LastError() != nil || dbIDList.Size() < 1 {
		clog.Warnf("[nodeID = %s] `db_id_list` property not exists.", s.App().NodeID())
		return
	}

	dbConfig := cprofile.GetConfig("db")
	if dbConfig.LastError() != nil {
		clog.Panic("`db` property not exists in profile file.")
	}

	for _, groupID := range dbConfig.Keys() {
		s.ormMap[groupID] = make(map[string]*gorm.DB)

		dbGroup := dbConfig.GetConfig(groupID)
		for i := 0; i < dbGroup.Size(); i++ {
			item := dbGroup.GetConfig(i)
			mysqlConfig := parseMysqlConfig(groupID, item)

			for _, key := range dbIDList.Keys() {
				if dbIDList.Get(key).ToString() != mysqlConfig.ID {
					continue
				}

				if !mysqlConfig.Enable {
					clog.Panicf("[dbName = %s] is disabled!", mysqlConfig.DbName)
				}

				db, err := s.createORM(mysqlConfig)
				if err != nil {
					clog.Panicf("[dbName = %s] create orm fail. error = %s", mysqlConfig.DbName, err)
				}

				s.ormMap[groupID][mysqlConfig.ID] = db
				clog.Infof("[dbGroup =%s, dbName = %s] is connected.", mysqlConfig.GroupID, mysqlConfig.ID)
			}
		}
	}
}

func (s *Component) createORM(cfg *mySqlConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{
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

func (s *Component) GetHashDb(groupID string, hashFn HashDb) (*gorm.DB, bool) {
	dbGroup, found := s.GetDbMap(groupID)
	if !found {
		clog.Warnf("groupID = %s not found.", groupID)
		return nil, false
	}

	dbID := hashFn(dbGroup)
	db, found := dbGroup[dbID]
	return db, found
}

func (s *Component) GetDbMap(groupID string) (map[string]*gorm.DB, bool) {
	dbGroup, found := s.ormMap[groupID]
	return dbGroup, found
}

func (s *mySqlConfig) GetDSN() string {
	if s.DSN == "" {
		s.DSN = dsn
	}

	return fmt.Sprintf(s.DSN, s.UserName, s.Password, s.Host, s.DbName)
}
