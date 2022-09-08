package main

import (
	"github.com/cherry-game/cherry"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	cherry.Configure("../../config/", "sample1", "cross-1")
	cherry.SetSerializer(cherrySerializer.NewJSON())

	cherry.Run(false, cherry.Cluster)
}
