package cherryORM

import (
	clog "github.com/cherry-game/cherry/logger"
	"strings"
)

type ormLogger struct {
	log *clog.CherryLogger
}

func (l ormLogger) Printf(s string, i ...interface{}) {
	l.log.Debugf(strings.ReplaceAll(s, "\n", ""), i...)
}
