package cherryMongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
)

const (
	Name = "mongo_component"
)

type (
	Component struct {
		cfacade.Component
		dbMap map[string]map[string]*mongo.Database
	}

	// HashDb hash by group id
	HashDb func(dbMaps map[string]*mongo.Database) string
)

func NewComponent() *Component {
	return &Component{
		dbMap: make(map[string]map[string]*mongo.Database),
	}
}

func (*Component) Name() string {
	return Name
}

func (s *Component) Init() {
	// load only the database contained in the `db_id_list`
	mongoIDList := s.App().Settings().Get("mongo_id_list")
	if mongoIDList.LastError() != nil || mongoIDList.Size() < 1 {
		clog.Warnf("[nodeID = %s] `mongo_id_list` property not exists.", s.App().NodeID())
		return
	}

	mongoConfig := cprofile.GetConfig("mongo")
	if mongoConfig.LastError() != nil {
		panic("`mongo` property not exists in profile file.")
	}

	for _, groupID := range mongoConfig.Keys() {
		s.dbMap[groupID] = make(map[string]*mongo.Database)

		dbGroup := mongoConfig.GetConfig(groupID)
		for i := 0; i < dbGroup.Size(); i++ {
			item := dbGroup.GetConfig(i)

			var (
				enable  = item.GetBool("enable", true)
				id      = item.GetString("db_id")
				dbName  = item.GetString("db_name")
				uri     = item.GetString("uri")
				timeout = time.Duration(item.GetInt64("timeout", 3)) * time.Second
			)

			if !enable {
				continue
			}

			for _, key := range mongoIDList.Keys() {
				if mongoIDList.Get(key).ToString() != id {
					continue
				}

				db, err := CreateDatabase(uri, dbName, timeout)
				if err != nil {
					panic(fmt.Sprintf("[dbName = %s] create mongodb fail. error = %s", dbName, err))
				}

				s.dbMap[groupID][id] = db
				clog.Infof("[dbGroup =%s, dbName = %s] is connected.", groupID, id)
			}
		}
	}
}

func CreateDatabase(uri, dbName string, timeout ...time.Duration) (*mongo.Database, error) {
	tt := 3 * time.Second

	if len(timeout) > 0 && timeout[0].Seconds() > 3 {
		tt = timeout[0]
	}

	o := options.Client().ApplyURI(uri)
	if err := o.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), tt)
	defer cancel()

	client, err := mongo.Connect(o)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	clog.Infof("ping db [uri = %s] is ok", uri)

	return client.Database(dbName), nil
}

func (s *Component) GetDb(id string) *mongo.Database {
	for _, group := range s.dbMap {
		for k, v := range group {
			if k == id {
				return v
			}
		}
	}
	return nil
}

func (s *Component) GetHashDb(groupID string, hashFn HashDb) (*mongo.Database, bool) {
	dbGroup, found := s.GetDbMap(groupID)
	if !found {
		clog.Warnf("groupID = %s not found.", groupID)
		return nil, false
	}

	dbID := hashFn(dbGroup)
	db, found := dbGroup[dbID]
	return db, found
}

func (s *Component) GetDbMap(groupID string) (map[string]*mongo.Database, bool) {
	dbGroup, found := s.dbMap[groupID]
	return dbGroup, found
}
