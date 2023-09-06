package cherryTime

import (
	"testing"
)

func TestCherryTime_SetYear(t *testing.T) {
	now := Now()
	now.SetYear(2012)
	t.Logf("result = %v", now.ToDateTimeFormat())
}

func TestCherryTime_SetMonth(t *testing.T) {
	now := Now()
	now.SetMonth(12)
	t.Logf("result = %v", now.ToDateTimeFormat())
}

func TestCherryTime_SetDay(t *testing.T) {
	now := Now()
	now.SetDay(12)
	t.Logf("result = %v", now.ToDateTimeFormat())
}

func TestCherryTime_SetHour(t *testing.T) {
	now := Now()
	now.SetHour(0)
	t.Logf("result = %v", now.ToDateTimeFormat())
}

func TestCherryTime_SetMinute(t *testing.T) {
	now := Now()
	now.SetMinute(0)
	t.Logf("result = %v", now.ToDateTimeFormat())
}

func TestCherryTime_SetSecond(t *testing.T) {
	now := Now()
	now.SetSecond(59)
	t.Logf("result = %v", now.ToDateTimeFormat())

	now.SetSecond(60)
	t.Logf("result = %v", now.ToDateTimeFormat())
}
