package cherry

import (
	"flag"
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/packet"
	"github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/profile"
)

var thisNode *Application

func App() *Application {
	return thisNode
}

func NewDefaultApp() *Application {
	var configPath, profile, nodeId string
	flag.StringVar(&configPath, "path", "./config", "-path=~/git/project/config")
	flag.StringVar(&profile, "profile", "local", "-profile=local")
	flag.StringVar(&nodeId, "node", "game-1", "-node=game-1")
	flag.Parse()

	return NewApp(configPath, profile, nodeId)
}

// NewApp create new application instance
func NewApp(profilePath, profileName, nodeId string) *Application {
	_, err := cherryProfile.Init(profilePath, profileName)
	if err != nil {
		panic(fmt.Sprintf("init profile fail. error = %s", err))
	}

	node, err := cherryProfile.LoadNode(nodeId)
	if err != nil {
		panic(fmt.Sprintf("error = %s", err))
	}

	// set logger
	cherryLogger.SetNodeLogger(node)

	// print version info
	cherryLogger.Info(cherryConst.GetLOGO())

	thisNode = &Application{
		INode:        node,
		startTime:    cherryTime.Now(),
		running:      0,
		die:          make(chan bool),
		ISerializer:  cherrySerializer.NewProtobuf(),
		IPacketCodec: cherryPacket.NewPomeloCodec(),
	}

	return thisNode
}
