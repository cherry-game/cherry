package cherryORM

import (
    cherryLogger "github.com/cherry-game/cherry/logger"
)

type ormLogger struct {
    log *cherryLogger.CherryLogger
}

func (l ormLogger) Printf(s string, i ...interface{}) {
    l.log.Debugf(s, i...)
}
