package cherryConst

import (
	"fmt"
)

// component name
const (
	HandlerComponent    = "handler_component"
	SessionComponent    = "session_component"
	ConnectorComponent  = "connector_component"
	DataConfigComponent = "data_config_component"
	ORMComponent        = "db_orm_component"

	RPCServerComponent = "rpc_server_component"
	RPCClientComponent = "rpc_client_component"
)

const (
	ProfileNameFormat = "profile-%s.json"
	version           = "1.0.0"
)

var logo = `

░█████╗░██╗░░██╗███████╗██████╗░██████╗░██╗░░░██╗
██╔══██╗██║░░██║██╔════╝██╔══██╗██╔══██╗╚██╗░██╔╝
██║░░╚═╝███████║█████╗░░██████╔╝██████╔╝░╚████╔╝░
██║░░██╗██╔══██║██╔══╝░░██╔══██╗██╔══██╗░░╚██╔╝░░
╚█████╔╝██║░░██║███████╗██║░░██║██║░░██║░░░██║░░░
░╚════╝░╚═╝░░╚═╝╚══════╝╚═╝░░╚═╝╚═╝░░╚═╝░░░╚═╝░░░ 
game sever framework @v%s
`

func GetLOGO() string {
	return fmt.Sprintf(logo, Version())
}

func Version() string {
	return version
}
