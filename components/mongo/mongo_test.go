package cherryMongo

import (
	"context"
	clog "github.com/cherry-game/cherry/logger"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	clog.Info("test connect mongodb")

	uri := "mongodb://localhost:27017"
	dbName := "test"

	mdb, err := CreateDatabase(uri, dbName)
	if err != nil {
		clog.Warn(err)
		return
	}

	collection := mdb.Collection("numbers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	id := res.InsertedID
	clog.Infof("id = %v, err = %v", id, err)
}
