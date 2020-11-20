package cherryConst

import (
	"fmt"
	"github.com/cherry-game/cherry/logger"
)

// component name
const (
	HandlerComponent = "handler_component"
	SessionComponent = "session_component"
	ORMComponent     = "db_orm_component"
	QueueComponent   = "db_queue_component"
)

const (
	ProfileFileName = "profile-%s.json"
	version         = "1.0.0"
)

var logo = `
░█████╗░██╗░░██╗███████╗██████╗░██████╗░██╗░░░██╗
██╔══██╗██║░░██║██╔════╝██╔══██╗██╔══██╗╚██╗░██╔╝
██║░░╚═╝███████║█████╗░░██████╔╝██████╔╝░╚████╔╝░
██║░░██╗██╔══██║██╔══╝░░██╔══██╗██╔══██╗░░╚██╔╝░░
╚█████╔╝██║░░██║███████╗██║░░██║██║░░██║░░░██║░░░
░╚════╝░╚═╝░░╚═╝╚══════╝╚═╝░░╚═╝╚═╝░░╚═╝░░░╚═╝░░░
`

var versionInfo = `game sever framework@v%s
-------------------------------------------------
`

func PrintVersion() {
	cherryLogger.Info(logo, fmt.Sprintf(versionInfo, version))
}

func Version() string {
	return version
}
