package cherryTime

import (
	"fmt"
	"testing"
)

func TestCherryTime_IsNow(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsNow()))
}

func TestCherryTime_IsFuture(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsFuture()))
}

func TestCherryTime_IsPast(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsPast()))
}

func TestCherryTime_IsLeapYear(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsLeapYear()))
}

func TestCherryTime_IsLongYear(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsLongYear()))
}

func TestCherryTime_IsJanuary(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsJanuary()))
}

func TestCherryTime_IsFebruary(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsFebruary()))
}

func TestCherryTime_IsMarch(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsMarch()))
}

func TestCherryTime_IsApril(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsApril()))
}

func TestCherryTime_IsMay(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsMay()))
}

func TestCherryTime_IsJune(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsJune()))
}

func TestCherryTime_IsJuly(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsJuly()))
}

func TestCherryTime_IsAugust(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsAugust()))
}

func TestCherryTime_IsSeptember(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsSeptember()))
}

func TestCherryTime_IsOctober(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsOctober()))
}

func TestCherryTime_IsDecember(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsDecember()))
}

func TestCherryTime_IsMonday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsMonday()))
}

func TestCherryTime_IsTuesday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsTuesday()))
}

func TestCherryTime_IsWednesday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsWednesday()))
}

func TestCherryTime_IsThursday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsThursday()))
}

func TestCherryTime_IsFriday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsFriday()))
}

func TestCherryTime_IsSaturday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsSaturday()))
}

func TestCherryTime_IsSunday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsSunday()))
}

func TestCherryTime_IsWeekday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsWeekday()))
}

func TestCherryTime_IsWeekend(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsWeekend()))
}

func TestCherryTime_IsYesterday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsYesterday()))
}

func TestCherryTime_IsToday(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsToday()))
}

func TestCherryTime_IsTomorrow(t *testing.T) {
	now := Now()
	t.Log(fmt.Sprintf("result = %v", now.IsTomorrow()))
}
