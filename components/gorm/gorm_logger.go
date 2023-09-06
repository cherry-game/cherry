package cherryGORM

import (
	"strings"

	clog "github.com/cherry-game/cherry/logger"
)

type gormLogger struct {
	log *clog.CherryLogger
}

func (l gormLogger) Printf(s string, i ...interface{}) {
	l.log.Debugf(strings.ReplaceAll(s, "\n", ""), i...)
}
