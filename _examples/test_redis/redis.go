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

	cmd := rdb.Set(ctx, "data_config:test", jsonData, 10*time.Hour)
	cherryLogger.Debug(cmd.Val(), cmd.Err())

	keysVal := rdb.Keys(ctx, "node_list*")
	cherryLogger.Debug(keysVal.Val(), keysVal.Err())

	scanVal := rdb.Scan(ctx, 0, "node_list*", 0)
	cherryLogger.Debug(scanVal.Result())

	go func() {
		subscribe := rdb.Subscribe(ctx, "aaaa")
		defer subscribe.Close()

		ch := subscribe.Channel()
		for msg := range ch {
			fmt.Println(msg.Channel, msg.Payload)
		}
	}()

	time.Sleep(1 * time.Hour)

}
