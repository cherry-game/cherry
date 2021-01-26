package cherry

import (
	"flag"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/cluster"
	"github.com/cherry-game/cherry/profile"
	"time"
)

func DefaultApp() *Application {
	var configPath, profile, nodeId string
	flag.StringVar(&configPath, "path", "./config", "-path=~/git/project/config")
	flag.StringVar(&profile, "profile", "local", "-profile=local")
	flag.StringVar(&nodeId, "node", "game-1", "-node=game-1")
	flag.Parse()

	return NewApp(configPath, profile, nodeId)
}

// NewApp create new application instance
func NewApp(configPath, profile, nodeId string) *Application {
	err := cherryProfile.Init(configPath, profile)
	if err != nil {
		panic(err)
	}

	//set logger
	cherryLogger.SetLogger(cherryProfile.Config())

	//print version info
	cherryConst.PrintVersion()

	//load nodes from config file
	cherryCluster.LoadNodes(cherryProfile.Config())

	nodeType, err := cherryCluster.Nodes().GetType(nodeId)
	if err != nil {
		cherryLogger.Panic(err)
		return nil
	}

	app := &Application{
		nodeId:    nodeId,
		nodeType:  nodeType,
		startTime: time.Now().Unix(),
		running:   false,
		die:       make(chan bool),
	}
	return app
}
