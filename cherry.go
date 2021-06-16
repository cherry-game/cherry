package cherry

import (
	"flag"
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/cluster"
	"github.com/cherry-game/cherry/profile"
)

var (
	thisApp *Application
)

func App() *Application {
	return thisApp
}

func NewDefaultApp() *Application {
	var configPath, profile, nodeId string
	flag.StringVar(&configPath, "path", "./profile", "-path=~/git/project/profile")
	flag.StringVar(&profile, "profile", "local", "-profile=local")
	flag.StringVar(&nodeId, "node", "game-1", "-node=game-1")
	flag.Parse()

	return NewApp(configPath, profile, nodeId)
}

// NewApp create new application instance
func NewApp(profilePath, profileName, nodeId string) *Application {
	config, err := cherryProfile.Init(profilePath, profileName)
	if err != nil {
		panic(fmt.Sprintf("init profile fail. error = %s", err))
	}

	err = cherryCluster.Load(config)
	if err != nil {
		panic(fmt.Sprintf("load node config fail. error = %s", err))
	}

	node, err := cherryCluster.GetNode(nodeId)
	if err != nil {
		panic(fmt.Sprintf("nodeId = %s not found. error = %s", nodeId, err))
	}

	// set logger
	cherryLogger.SetNodeLogger(node)

	// print version info
	cherryLogger.Info(cherryConst.GetLOGO())

	thisNode := &Application{
		INode:     node,
		startTime: cherryTime.Now(),
		running:   0,
		die:       make(chan bool),
	}
	return thisNode
}
