package cherryTime

import (
	"testing"
)

func TestCherryTime_IsNow(t *testing.T) {
	t.Logf("result = %v", Now().IsNow())
}

func TestCherryTime_IsFuture(t *testing.T) {
	t.Logf("result = %v", Now().IsFuture())
}

func TestCherryTime_IsPast(t *testing.T) {
	t.Logf("result = %v", Now().IsPast())
}

func TestCherryTime_IsLeapYear(t *testing.T) {
	t.Logf("result = %v", Now().IsLeapYear())
}

func TestCherryTime_IsLongYear(t *testing.T) {
	t.Logf("result = %v", Now().IsLongYear())
}

func TestCherryTime_IsJanuary(t *testing.T) {
	t.Logf("result = %v", Now().IsJanuary())
}

func TestCherryTime_IsFebruary(t *testing.T) {
	t.Logf("result = %v", Now().IsFebruary())
}

func TestCherryTime_IsMarch(t *testing.T) {
	t.Logf("result = %v", Now().IsMarch())
}

func TestCherryTime_IsApril(t *testing.T) {
	t.Logf("result = %v", Now().IsApril())
}

func TestCherryTime_IsMay(t *testing.T) {
	t.Logf("result = %v", Now().IsMay())
}

func TestCherryTime_IsJune(t *testing.T) {
	t.Logf("result = %v", Now().IsJune())
}

func TestCherryTime_IsJuly(t *testing.T) {
	t.Logf("result = %v", Now().IsJuly())
}

func TestCherryTime_IsAugust(t *testing.T) {
	t.Logf("result = %v", Now().IsAugust())
}

func TestCherryTime_IsSeptember(t *testing.T) {
	t.Logf("result = %v", Now().IsSeptember())
}

func TestCherryTime_IsOctober(t *testing.T) {
	t.Logf("result = %v", Now().IsOctober())
}

func TestCherryTime_IsDecember(t *testing.T) {
	t.Logf("result = %v", Now().IsDecember())
}

func TestCherryTime_IsMonday(t *testing.T) {
	t.Logf("result = %v", Now().IsMonday())
}

func TestCherryTime_IsTuesday(t *testing.T) {
	t.Logf("result = %v", Now().IsTuesday())
}

func TestCherryTime_IsWednesday(t *testing.T) {
	t.Logf("result = %v", Now().IsWednesday())
}

func TestCherryTime_IsThursday(t *testing.T) {
	t.Logf("result = %v", Now().IsThursday())
}

func TestCherryTime_IsFriday(t *testing.T) {
	t.Logf("result = %v", Now().IsFriday())
}

func TestCherryTime_IsSaturday(t *testing.T) {
	t.Logf("result = %v", Now().IsSaturday())
}

func TestCherryTime_IsSunday(t *testing.T) {
	t.Logf("result = %v", Now().IsSunday())
}

func TestCherryTime_IsWeekday(t *testing.T) {
	t.Logf("result = %v", Now().IsWeekday())
}

func TestCherryTime_IsWeekend(t *testing.T) {
	t.Logf("result = %v", Now().IsWeekend())
}

func TestCherryTime_IsYesterday(t *testing.T) {
	t.Logf("result = %v", Now().IsYesterday())
}

func TestCherryTime_IsYesterday1(t *testing.T) {
	now := Now()
	now.SubDay()
	t.Logf("result = %v", now.IsYesterday())
}

func TestCherryTime_IsToday(t *testing.T) {
	t.Logf("result = %v", Now().IsToday())
}

func TestCherryTime_IsTomorrow(t *testing.T) {
	t.Logf("result = %v", Now().IsTomorrow())
}
