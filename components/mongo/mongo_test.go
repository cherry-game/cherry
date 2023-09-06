package cherryMongo

import (
	"context"
	"fmt"
	"testing"

	clog "github.com/cherry-game/cherry/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Student struct {
	Name string
	Age  int
}

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

	student := &Student{
		Name: "aaa222",
		Age:  111,
	}

	res, err := collection.InsertOne(context.Background(), student)
	insertID := res.InsertedID
	clog.Infof("id = %v, err = %v", insertID, err)

	//id, _ := primitive.ObjectIDFromHex("649160b6c637f5773cc1e818")
	id, ok := insertID.(primitive.ObjectID)
	if !ok {
		return
	}

	findFilter := bson.M{"_id": id}
	findResult := collection.FindOne(context.Background(), findFilter)

	findStudent := Student{}
	findResult.Decode(&findStudent)
	fmt.Println(findStudent)
}
