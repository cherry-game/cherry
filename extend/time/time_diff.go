package cherryTime

// DiffInYears 相差多少年
func (c CherryTime) DiffInYears(end *CherryTime) int64 {
	return c.DiffInMonths(end) / 12
}

// DiffInYearsWithAbs 相差多少年(绝对值)
func (c CherryTime) DiffInYearsWithAbs(end *CherryTime) int64 {
	return GetAbsValue(c.DiffInYears(end))
}

// DiffInMonths 相差多少月
func (c CherryTime) DiffInMonths(end *CherryTime) int64 {
	dy, dm, dd := end.Year()-c.Year(), end.Month()-c.Month(), end.Day()-c.Day()

	if dd < 0 {
		dm = dm - 1
	}
	if dy == 0 && dm == 0 {
		return 0
	}
	if dy == 0 && dm != 0 && dd != 0 {
		if int(end.DiffInHoursWithAbs(&c)) < c.DaysInMonth()*HoursPerDay {
			return 0
		}
		return int64(dm)
	}

	return int64(dy*MonthsPerYear + dm)
}

// DiffInMonthsWithAbs 相差多少月(绝对值)
func (c CherryTime) DiffInMonthsWithAbs(end *CherryTime) int64 {
	return GetAbsValue(c.DiffInMonths(end))
}

// DiffInWeeks 相差多少周
func (c CherryTime) DiffInWeeks(end *CherryTime) int64 {
	return c.DiffInDays(end) / DaysPerWeek
}

// DiffInWeeksWithAbs 相差多少周(绝对值)
func (c CherryTime) DiffInWeeksWithAbs(end *CherryTime) int64 {
	return GetAbsValue(c.DiffInWeeks(end))
}

// DiffInDays 相差多少天
func (c CherryTime) DiffInDays(end *CherryTime) int64 {
	return c.DiffInSeconds(end) / SecondsPerDay
}

// DiffInDaysWithAbs 相差多少天(绝对值)
func (c CherryTime) DiffInDaysWithAbs(end *CherryTime) int64 {
	return GetAbsValue(c.DiffInDays(end))
}

// DiffInHours 相差多少小时
func (c CherryTime) DiffInHours(end *CherryTime) int64 {
	return c.DiffInSeconds(end) / SecondsPerHour
}

// DiffInHoursWithAbs 相差多少小时(绝对值)
func (c CherryTime) DiffInHoursWithAbs(end *CherryTime) int64 {
	return GetAbsValue(c.DiffInHours(end))
}

// DiffInMinutes 相差多少分钟
func (c CherryTime) DiffInMinutes(end *CherryTime) int64 {
	return c.DiffInSeconds(end) / SecondsPerMinute
}

// DiffInMinutesWithAbs 相差多少分钟(绝对值)
func (c CherryTime) DiffInMinutesWithAbs(end *CherryTime) int64 {
	return GetAbsValue(c.DiffInMinutes(end))
}

// DiffInSeconds 相差多少秒
func (c CherryTime) DiffInSeconds(end *CherryTime) int64 {
	return end.ToSecond() - c.ToSecond()
}

// DiffInSecondsWithAbs 相差多少秒(绝对值)
func (c CherryTime) DiffInSecondsWithAbs(end *CherryTime) int64 {
	return GetAbsValue(c.DiffInSeconds(end))
}

// DiffInMillisecond 相差多少毫秒
func (c CherryTime) DiffInMillisecond(end *CherryTime) int64 {
	return end.ToMillisecond() - c.ToMillisecond()
}

// DiffInMicrosecond 相差多少微秒
func (c CherryTime) DiffInMicrosecond(end *CherryTime) int64 {
	return end.ToMicrosecond() - c.ToMicrosecond()
}

// DiffINanosecond 相差多少纳秒
func (c CherryTime) DiffInNanosecond(end *CherryTime) int64 {
	return end.ToNanosecond() - c.ToNanosecond()
}

// DiffInNowMillisecond 与当前时间相差多少毫秒
func (c CherryTime) NowDiffMillisecond() int64 {
	return Now().ToMillisecond() - c.ToMillisecond()
}

// DiffInNowMillisecond 与当前时间相差多少秒
func (c CherryTime) NowDiffSecond() int64 {
	return Now().ToSecond() - c.ToSecond()
}
