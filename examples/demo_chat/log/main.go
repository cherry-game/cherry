package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cserializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	app := cherry.Configure(
		"./examples/config/profile-chat.json",
		"log-1",
		false,
		cherry.Cluster,
	)
	app.SetSerializer(cserializer.NewJSON())

	app.SetActorInvoke(pomelo.LocalInvokeFunc, pomelo.RemoteInvokeFunc)

	app.AddActors(
		&ActorLog{},
	)

	app.Startup()
}
