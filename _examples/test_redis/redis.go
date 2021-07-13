package main

import (
	"context"
	"encoding/json"
	"fmt"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"github.com/go-redis/redis/v8"
	"time"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()

	data := struct {
		UserName string `json:"user_name"`
		Address  string `json:"address"`
	}{
		UserName: "tom",
		Address:  "china shenzhen",
	}

	jsonData, err := json.Marshal(&data)
	if err != nil {
		return
	}

	cmd := rdb.Set(ctx, "data_config:user_data:user", jsonData, 10*time.Hour)
	cherryLogger.Debug(cmd.Val(), cmd.Err())

	val := rdb.Get(ctx, "data_config:user_data:user")
	cherryLogger.Debug(val.Val(), val.Err())

	go func() {
		pubsub := rdb.Subscribe(ctx, "aaaa")
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			fmt.Println(msg.Channel, msg.Payload)
		}
	}()

	time.Sleep(1 * time.Hour)

}
