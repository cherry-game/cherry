package cherryTime

// DaysInYear 获取本年的总天数
func (c CherryTime) DaysInYear() int {
	if c.IsZero() {
		return 0
	}
	days := DaysPerNormalYear
	if c.IsLeapYear() {
		days = DaysPerLeapYear
	}
	return days
}

// DaysInMonth 获取本月的总天数
func (c CherryTime) DaysInMonth() int {
	if c.IsZero() {
		return 0
	}
	return c.EndOfMonth().Day()
}

// MonthOfYear 获取本年的第几月(从1开始)
func (c CherryTime) MonthOfYear() int {
	if c.IsZero() {
		return 0
	}
	return int(c.Time.Month())
}

// DayOfYear 获取本年的第几天(从1开始)
func (c CherryTime) DayOfYear() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.YearDay()
}

// DayOfMonth 获取本月的第几天(从1开始)
func (c CherryTime) DayOfMonth() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Day()
}

// DayOfWeek 获取本周的第几天(从1开始)
func (c CherryTime) DayOfWeek() int {
	if c.IsZero() {
		return 0
	}
	day := int(c.Time.Weekday())
	if day == 0 {
		return DaysPerWeek
	}
	return day
}

// WeekOfYear 获取本年的第几周(从1开始)
func (c CherryTime) WeekOfYear() int {
	if c.IsZero() {
		return 0
	}
	_, week := c.Time.ISOWeek()
	return week
}

// WeekOfMonth 获取本月的第几周(从1开始)
func (c CherryTime) WeekOfMonth() int {
	if c.IsZero() {
		return 0
	}
	day := c.Time.Day()
	if day < DaysPerWeek {
		return 1
	}
	return day%DaysPerWeek + 1
}

// Year 获取当前年
func (c CherryTime) Year() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Year()
}

// Quarter 获取当前季度
func (c CherryTime) Quarter() int {
	if c.IsZero() {
		return 0
	}
	switch {
	case c.Month() >= 10:
		return 4
	case c.Month() >= 7:
		return 3
	case c.Month() >= 4:
		return 2
	case c.Month() >= 1:
		return 1
	default:
		return 0
	}
}

// Month 获取当前月
func (c CherryTime) Month() int {
	if c.IsZero() {
		return 0
	}
	return c.MonthOfYear()
}

// Week 获取当前周(从0开始)
func (c CherryTime) Week() int {
	if c.IsZero() {
		return -1
	}
	return int(c.Time.Weekday())
}

// Day 获取当前日
func (c CherryTime) Day() int {
	if c.IsZero() {
		return 0
	}
	return c.DayOfMonth()
}

// Hour 获取当前小时
func (c CherryTime) Hour() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Hour()
}

// Minute 获取当前分钟数
func (c CherryTime) Minute() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Minute()
}

// Second 获取当前秒数
func (c CherryTime) Second() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Second()
}

// Millisecond 获取当前毫秒数
func (c CherryTime) Millisecond() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Nanosecond() / 1e6
}

// Microsecond 获取当前微秒数
func (c CherryTime) Microsecond() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Nanosecond() / 1e9
}

// Nanosecond 获取当前纳秒数
func (c CherryTime) Nanosecond() int {
	if c.IsZero() {
		return 0
	}
	return c.Time.Nanosecond()
}

// Timezone 获取时区
func (c CherryTime) Timezone() string {
	return c.Location().String()
}
