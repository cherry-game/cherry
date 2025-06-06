package cherryTime

import (
	"testing"
)

func TestCherryTime_DaysInYear(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.DaysInYear())
}

func TestCherryTime_DaysInMonth(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.DaysInMonth())
}

func TestCherryTime_MonthOfYear(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.MonthOfYear())
}

func TestCherryTime_DayOfYear(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.DayOfYear())
}

func TestCherryTime_DayOfMonth(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.DayOfMonth())
}

func TestCherryTime_DayOfWeek(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.DayOfWeek())
}

func TestCherryTime_WeekOfYear(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.WeekOfYear())
}

func TestCherryTime_WeekOfMonth(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.WeekOfMonth())
}

func TestCherryTime_Year(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Year())
}

func TestCherryTime_Quarter(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Quarter())
}

func TestCherryTime_Month(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Month())
}

func TestCherryTime_Week(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Week())
}

func TestCherryTime_Day(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Day())
}

func TestCherryTime_Hour(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Hour())
}

func TestCherryTime_Minute(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Minute())
}

func TestCherryTime_Second(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Second())
}

func TestCherryTime_Millisecond(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Millisecond())
}

func TestCherryTime_Microsecond(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Microsecond())
}

func TestCherryTime_Nanosecond(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Nanosecond())
}

func TestCherryTime_Timezone(t *testing.T) {
	now := Now()
	t.Logf("result = %v", now.Timezone())
}
