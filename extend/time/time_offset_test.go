package cherryTime

import (
	"testing"
	"time"
)

func TestGlobalOffset(t *testing.T) {
	now := Now()
	t.Log(now.ToMillisecond(), " - ", now.ToDateTimeFormat())

	SubOffsetTime(SecondsPerDay * time.Second)
	nowOffset := Now()
	t.Log(nowOffset.ToMillisecond(), " - ", nowOffset.ToDateTimeFormat())

	nowBackup := nowOffset
	nowOffset.SubDay()
	t.Log(nowOffset.ToMillisecond(), " - ", nowOffset.ToDateTimeFormat())
	t.Log(nowBackup.ToMillisecond(), " - ", nowBackup.ToDateTimeFormat())

	nowOffset.AddDays(1)
	t.Log(nowOffset.ToMillisecond(), " - ", nowOffset.ToDateTimeFormat())
}
