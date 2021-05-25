package cherry

import (
	"flag"
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

func NewNodeApp(profileName, nodeId string) *Application {
	return NewApp("./profile", profileName, nodeId)
}

// NewApp create new application instance
func NewApp(profilePath, profileName, nodeId string) *Application {
	config, err := cherryProfile.Init(profilePath, profileName)
	if err != nil {
		panic(err)
	}

	err = cherryCluster.Load(config)
	if err != nil {
		panic(err)
	}

	node, err := cherryCluster.GetNode(nodeId)
	if err != nil {
		panic(err)
	}

	// set logger
	cherryLogger.SetNodeLogger(node)

	// print version info
	cherryLogger.Info(cherryConst.GetLOGO())

	thisNode := &Application{
		INode:     node,
		startTime: cherryTime.Now(),
		running:   false,
		die:       make(chan bool),
	}
	return thisNode
}
