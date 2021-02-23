package cherryTime

import "time"

// StartOfYear 本年开始时间
func (c CherryTime) StartOfYear() CherryTime {
	c.Time = time.Date(c.Time.Year(), 1, 1, 0, 0, 0, 0, c.Location())
	return c
}

// EndOfYear 本年结束时间
func (c CherryTime) EndOfYear() CherryTime {
	c.Time = time.Date(c.Time.Year(), 12, 31, 23, 59, 59, 0, c.Location())
	return c
}

// StartOfMonth 本月开始时间
func (c CherryTime) StartOfMonth() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), 1, 0, 0, 0, 0, c.Location())
	return c
}

// EndOfMonth 本月结束时间
func (c CherryTime) EndOfMonth() CherryTime {
	t := time.Date(c.Time.Year(), c.Time.Month(), 1, 23, 59, 59, 0, c.Location())
	c.Time = t.AddDate(0, 1, -1)
	return c
}

// StartOfWeek 本周开始时间
func (c CherryTime) StartOfWeek() CherryTime {
	days := c.Time.Weekday()
	if days == 0 {
		days = DaysPerWeek
	}
	t := time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), 0, 0, 0, 0, c.Location())
	c.Time = t.AddDate(0, 0, int(1-days))
	return c
}

// EndOfWeek 本周结束时间
func (c CherryTime) EndOfWeek() CherryTime {
	days := c.Time.Weekday()
	if days == 0 {
		days = DaysPerWeek
	}
	t := time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), 23, 59, 59, 0, c.Location())
	c.Time = t.AddDate(0, 0, int(DaysPerWeek-days))
	return c
}

// StartOfDay 本日开始时间
func (c CherryTime) StartOfDay() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), 0, 0, 0, 0, c.Location())
	return c
}

// EndOfDay 本日结束时间
func (c CherryTime) EndOfDay() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), 23, 59, 59, 0, c.Location())
	return c
}

// StartOfHour 小时开始时间
func (c CherryTime) StartOfHour() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), c.Time.Hour(), 0, 0, 0, c.Location())
	return c
}

// EndOfHour 小时结束时间
func (c CherryTime) EndOfHour() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), c.Time.Hour(), 59, 59, 0, c.Location())
	return c
}

// StartOfMinute 分钟开始时间
func (c CherryTime) StartOfMinute() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), c.Time.Hour(), c.Time.Minute(), 0, 0, c.Location())
	return c
}

// EndOfMinute 分钟结束时间
func (c CherryTime) EndOfMinute() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), c.Time.Hour(), c.Time.Minute(), 59, 0, c.Location())
	return c
}

// StartOfSecond 秒开始时间
func (c CherryTime) StartOfSecond() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), c.Time.Hour(), c.Time.Minute(), c.Time.Second(), 0, c.Location())
	return c
}

// EndOfSecond 秒结束时间
func (c CherryTime) EndOfSecond() CherryTime {
	c.Time = time.Date(c.Time.Year(), c.Time.Month(), c.Time.Day(), c.Time.Hour(), c.Time.Minute(), c.Time.Second(), 999999999, c.Location())
	return c
}
