package cherryTime

import (
	cherryLogger "github.com/cherry-game/cherry/logger"
	"testing"
	"time"
)

func TestGlobalOffset(t *testing.T) {
	now := Now()
	cherryLogger.Info(now.ToMillisecond(), " - ", now.ToDateTimeFormat())

	SubOffsetTime(SecondsPerDay * time.Second)
	nowOffset := Now()
	cherryLogger.Info(nowOffset.ToMillisecond(), " - ", nowOffset.ToDateTimeFormat())

	nowBackup := nowOffset
	nowOffset.SubDay()
	cherryLogger.Info(nowOffset.ToMillisecond(), " - ", nowOffset.ToDateTimeFormat())
	cherryLogger.Info(nowBackup.ToMillisecond(), " - ", nowBackup.ToDateTimeFormat())

	nowOffset.AddDays(1)
	cherryLogger.Info(nowOffset.ToMillisecond(), " - ", nowOffset.ToDateTimeFormat())
}
