package cherryConst

import (
	"fmt"
)

// component name
const (
	HandlerComponent    = "handler_component"
	ConnectorComponent  = "connector_component"
	ClusterComponent    = "cluster_component"
	DataConfigComponent = "data_config_component"
	ORMComponent        = "db_orm_component"
)

const (
	ProfileNameFormat = "profile-%s.json"
	version           = "1.1.20"
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
