package cherryTime

import (
	"fmt"
	"testing"
)

func TestCherryTime_DaysInYear(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().DaysInYear()))
}

func TestCherryTime_DaysInMonth(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().DaysInMonth()))
}

func TestCherryTime_MonthOfYear(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().MonthOfYear()))
}

func TestCherryTime_DayOfYear(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().DayOfYear()))
}

func TestCherryTime_DayOfMonth(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().DayOfMonth()))
}

func TestCherryTime_DayOfWeek(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().DayOfWeek()))
}

func TestCherryTime_WeekOfYear(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().WeekOfYear()))
}

func TestCherryTime_WeekOfMonth(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().WeekOfMonth()))
}

func TestCherryTime_Year(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Year()))
}

func TestCherryTime_Quarter(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Quarter()))
}

func TestCherryTime_Month(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Month()))
}

func TestCherryTime_Week(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Week()))
}

func TestCherryTime_Day(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Day()))
}

func TestCherryTime_Hour(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Hour()))
}

func TestCherryTime_Minute(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Minute()))
}

func TestCherryTime_Second(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Second()))
}

func TestCherryTime_Millisecond(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Millisecond()))
}

func TestCherryTime_Microsecond(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Microsecond()))
}

func TestCherryTime_Nanosecond(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Nanosecond()))
}

func TestCherryTime_Timezone(t *testing.T) {
	t.Log(fmt.Sprintf("result = %v", Now().Timezone()))
}
