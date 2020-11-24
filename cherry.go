package cherry

import (
	"github.com/cherry-game/cherry/cluster"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"time"
)

// CreateApp create new application instance
func CreateApp(configPath, profileName, nodeId string) *Application {
	err := cherryProfile.Init(configPath, profileName)
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
