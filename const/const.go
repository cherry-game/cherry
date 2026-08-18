// Package cherryConst holds framework-wide constants: version, logo, and separators.
package cherryConst

import (
	"fmt"
)

const (
	version = "1.6.1"
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

// GetLOGO returns the framework ASCII logo with version.
func GetLOGO() string {
	return fmt.Sprintf(logo, Version())
}

// Version returns the current framework version.
func Version() string {
	return version
}

const (
	DOT = "." // ActorPath separator
)
