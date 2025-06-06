package cherryTime

import "time"

func (c *CherryTime) SetTimezone(timezone string) error {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return err
	}

	c.Time = c.Time.In(loc)
	return nil
}

func (c *CherryTime) SetDate(year, month, day int) {
	c.Time = time.Date(year, time.Month(month), day, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
}

func (c *CherryTime) SetTime(hour, mintue, second, nanoSecond int) {
	c.Time = time.Date(c.Year(), c.Time.Month(), c.Day(), hour, mintue, second, nanoSecond, c.Location())
}

// SetYear 设置年
func (c *CherryTime) SetYear(year int) {
	c.SetDate(year, c.Month(), c.Day())
}

// SetMonth 设置月
func (c *CherryTime) SetMonth(month int) {
	c.SetDate(c.Year(), month, c.Day())
}

// SetDay 设置日
func (c *CherryTime) SetDay(day int) {
	c.SetDate(c.Year(), c.Month(), day)
}

// SetHour 设置时
func (c *CherryTime) SetHour(hour int) {
	c.SetTime(hour, c.Minute(), c.Second(), c.Nanosecond())
}

// SetMinute 设置分
func (c *CherryTime) SetMinute(minute int) {
	c.SetTime(c.Hour(), minute, c.Second(), c.Nanosecond())
}

// SetSecond 设置秒
func (c *CherryTime) SetSecond(second int) {
	c.SetTime(c.Hour(), c.Minute(), second, c.Nanosecond())
}

// SetNanoSecond 设置纳秒
func (c *CherryTime) SetNanoSecond(nanoSecond int) {
	c.SetTime(c.Hour(), c.Minute(), c.Second(), nanoSecond)
}
