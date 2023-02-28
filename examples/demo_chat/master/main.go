package main

import (
	"github.com/cherry-game/cherry"
	cserializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	app := cherry.Configure(
		"./examples/config/profile-chat.json",
		"chat-master",
		false,
		cherry.Cluster,
	)
	app.SetSerializer(cserializer.NewJSON())

	app.Startup()
}
