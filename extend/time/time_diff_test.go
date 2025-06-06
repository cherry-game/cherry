package cherryTime

import (
	"testing"
)

func TestCherryTime_DiffInYears(t *testing.T) {
	ct1 := CreateFromDate(2012, 12, 1)
	ct2 := CreateFromDate(2022, 2, 1)

	years := ct1.DiffInYears(&ct2)
	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInYearsWithAbs(t *testing.T) {
	ct1 := CreateFromDate(2012, 12, 1)
	ct2 := CreateFromDate(2022, 2, 1)

	years := ct1.DiffInYearsWithAbs(&ct2)

	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInMonths(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	month := ct1.DiffInMonths(&ct2)

	t.Logf("result = %v", month)
}

func TestCherryTime_DiffInMonthsWithAbs(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	month := ct1.DiffInMonthsWithAbs(&ct2)

	t.Logf("result = %v", month)
}

func TestCherryTime_DiffInWeeks(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	week := ct1.DiffInWeeks(&ct2)

	t.Logf("result = %v", week)
}

func TestCherryTime_DiffInWeeksWithAbs(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	years := ct1.DiffInWeeksWithAbs(&ct2)

	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInDays(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	years := ct1.DiffInDays(&ct2)

	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInDaysWithAbs(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	years := ct1.DiffInDaysWithAbs(&ct2)

	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInHours(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	years := ct1.DiffInHours(&ct2)

	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInHoursWithAbs(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	years := ct1.DiffInHoursWithAbs(&ct2)

	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInSeconds(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	years := ct1.DiffInSeconds(&ct2)

	t.Logf("result = %v", years)
}

func TestCherryTime_DiffInSecondsWithAbs(t *testing.T) {
	ct1 := CreateFromDate(2021, 12, 15)
	ct2 := CreateFromDate(2022, 1, 1)

	years := ct1.DiffInSecondsWithAbs(&ct2)

	t.Logf("result = %v", years)
}
