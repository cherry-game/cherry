package cherryTime

import "time"

// IsNow 是否是当前时间
func (c CherryTime) IsNow() bool {
	now := Now()
	return c.ToSecond() == now.ToSecond()
}

// IsFuture 是否是未来时间
func (c CherryTime) IsFuture() bool {
	now := Now()
	return c.ToSecond() > now.ToSecond()
}

// IsPast 是否是过去时间
func (c CherryTime) IsPast() bool {
	now := Now()
	return c.ToSecond() < now.ToSecond()
}

// IsLeapYear 是否是闰年
func (c CherryTime) IsLeapYear() bool {
	year := c.Year()
	if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
		return true
	}
	return false
}

// IsLongYear 是否是长年
func (c CherryTime) IsLongYear() bool {
	_, w := time.Date(c.Year(), time.December, 31, 0, 0, 0, 0, c.Location()).ISOWeek()
	return w == WeeksPerLongYear
}

// IsJanuary 是否是一月
func (c CherryTime) IsJanuary() bool {
	return c.Time.Month() == time.January
}

// IsFebruary 是否是二月
func (c CherryTime) IsFebruary() bool {
	return c.Time.Month() == time.February
}

// IsMarch 是否是三月
func (c CherryTime) IsMarch() bool {
	return c.Time.Month() == time.March
}

// IsApril 是否是四月
func (c CherryTime) IsApril() bool {
	return c.Time.Month() == time.April
}

// IsMay 是否是五月
func (c CherryTime) IsMay() bool {
	return c.Time.Month() == time.May
}

// IsJune 是否是六月
func (c CherryTime) IsJune() bool {
	return c.Time.Month() == time.June
}

// IsJuly 是否是七月
func (c CherryTime) IsJuly() bool {
	return c.Time.Month() == time.July
}

// IsAugust 是否是八月
func (c CherryTime) IsAugust() bool {
	return c.Time.Month() == time.August
}

// IsSeptember 是否是九月
func (c CherryTime) IsSeptember() bool {
	return c.Time.Month() == time.September
}

// IsOctober 是否是十月
func (c CherryTime) IsOctober() bool {
	return c.Time.Month() == time.October
}

// IsNovember 是否是十一月
func (c CherryTime) IsNovember() bool {
	return c.Time.Month() == time.November
}

// IsDecember 是否是十二月
func (c CherryTime) IsDecember() bool {
	return c.Time.Month() == time.December
}

// IsMonday 是否是周一
func (c CherryTime) IsMonday() bool {
	return c.Time.Weekday() == time.Monday
}

// IsTuesday 是否是周二
func (c CherryTime) IsTuesday() bool {
	return c.Time.Weekday() == time.Tuesday
}

// IsWednesday 是否是周三
func (c CherryTime) IsWednesday() bool {
	return c.Time.Weekday() == time.Wednesday
}

// IsThursday 是否是周四
func (c CherryTime) IsThursday() bool {
	return c.Time.Weekday() == time.Thursday
}

// IsFriday 是否是周五
func (c CherryTime) IsFriday() bool {
	return c.Time.Weekday() == time.Friday
}

// IsSaturday 是否是周六
func (c CherryTime) IsSaturday() bool {
	return c.Time.Weekday() == time.Saturday
}

// IsSunday 是否是周日
func (c CherryTime) IsSunday() bool {
	return c.Time.Weekday() == time.Sunday
}

// IsWeekday 是否是工作日
func (c CherryTime) IsWeekday() bool {
	return !c.IsSaturday() && !c.IsSunday()
}

// IsWeekend 是否是周末
func (c CherryTime) IsWeekend() bool {
	return c.IsSaturday() || c.IsSunday()
}

// IsYesterday 是否是昨天
func (c CherryTime) IsYesterday() bool {
	yesterday := Now()
	yesterday.SubDay()
	return c.ToDateFormat() == yesterday.ToDateFormat()
}

// IsToday 是否是今天
func (c CherryTime) IsToday() bool {
	now := Now()
	return c.ToDateFormat() == now.ToDateFormat()
}

// IsTomorrow 是否是明天
func (c CherryTime) IsTomorrow() bool {
	tomorrow := Now()
	tomorrow.AddDay()
	return c.ToDateFormat() == tomorrow.ToDateFormat()
}
