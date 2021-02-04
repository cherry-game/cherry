package cherry

import (
	"flag"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/cluster"
	"github.com/cherry-game/cherry/profile"
	"time"
)

var (
	thisApp *Application
)

func App() *Application {
	return thisApp
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
func NewApp(configPath, profile, nodeId string) *Application {
	config, err := cherryProfile.Init(configPath, profile)
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
		nodeId:    nodeId,
		nodeType:  node.NodeType(),
		startTime: time.Now().Unix(),
		running:   false,
		die:       make(chan bool),
	}
	return thisNode
}
