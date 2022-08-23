package cherryConst

import (
	"fmt"
)

const (
	ProfileNameFormat = "profile-%s.json"
	version           = "1.2.2"
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
