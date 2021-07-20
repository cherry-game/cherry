package cherryORM

import (
    "fmt"
    "github.com/cherry-game/cherry/const"
    "github.com/cherry-game/cherry/facade"
    cherryLogger "github.com/cherry-game/cherry/logger"
    "github.com/cherry-game/cherry/profile"
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
        cherryFacade.Component

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

func NewComponent() *Component {
    return &Component{
        ormMap: make(map[string]map[string]*gorm.DB),
    }
}

func parseMysqlConfig(item jsoniter.Any) *mySqlConfig {
    return &mySqlConfig{
        Enable:         item.Get("enable").ToBool(),
        GroupId:        item.Get("group_id").ToString(),
        Id:             item.Get("id").ToString(),
        DbName:         item.Get("db_name").ToString(),
        Host:           item.Get("host").ToString(),
        UserName:       item.Get("user_name").ToString(),
        Password:       item.Get("password").ToString(),
        MaxIdleConnect: item.Get("max_idle_connect").ToInt(),
        MaXOpenConnect: item.Get("max_open_connect").ToInt(),
        LogMode:        item.Get("log_mode").ToBool(),
    }
}

func (s *Component) Name() string {
    return cherryConst.ORMComponent
}

func (s *Component) Init() {
    dbIdList := s.App().Settings().Get("db_id_list")
    if dbIdList.LastError() != nil {
        cherryLogger.Warnf("[nodeId = %s] `db_id_list` node not found.", s.App().NodeId())
        return
    }

    if dbIdList.Size() < 1 {
        cherryLogger.Warnf("[nodeId = %s] `db_id_list` node not config.", s.App().NodeId())
        return
    }

    rootJson := cherryProfile.GetConfig("db")
    for i := 0; i < rootJson.Size(); i++ {
        item := rootJson.Get(i)
        cfg := parseMysqlConfig(item)

        // load database by dbIdList
        for i := 0; i < dbIdList.Size(); i++ {
            dbId := dbIdList.Get(i).ToString()
            if cfg.Id != dbId {
                continue
            }

            if cfg.Enable == false {
                panic(fmt.Sprintf("[dbName = %s] is disabled. create orm fail.", cfg.DbName))
            }

            db, err := s.createORM(cfg)
            if err != nil {
                panic(fmt.Sprintf("[dbName = %s] create orm fail. error = %s", cfg.DbName, err))
            }

            dbs := s.ormMap[cfg.GroupId]
            if dbs == nil {
                dbs = make(map[string]*gorm.DB)
                s.ormMap[cfg.GroupId] = dbs
            }
            dbs[cfg.Id] = db
        }
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

    return db, nil
}

func getLogger() logger.Interface {
    return logger.New(
        ormLogger{log: cherryLogger.DefaultLogger},
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

func (s *Component) DBWithGroupId(groupId string) map[string]*gorm.DB {
    return s.ormMap[groupId]
}
