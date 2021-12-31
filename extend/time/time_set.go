package cherryTime

import "time"

func (c *CherryTime) SetTimezone(timezone string) error {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return err
	}

	c.In(loc)
	return nil
}

// SetYear 设置年
func (c *CherryTime) SetYear(year int) {
	c.Time = time.Date(year, c.Time.Month(), c.Day(), c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
}

// SetMonth 设置月
func (c *CherryTime) SetMonth(month int) {
	c.Time = time.Date(c.Year(), time.Month(month), c.Day(), c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
}

// SetDay 设置日
func (c *CherryTime) SetDay(day int) {
	c.Time = time.Date(c.Year(), c.Time.Month(), day, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
}

// SetHour 设置时
func (c *CherryTime) SetHour(hour int) {
	c.Time = time.Date(c.Year(), c.Time.Month(), c.Day(), hour, c.Minute(), c.Second(), c.Nanosecond(), c.Location())
}

// SetMinute 设置分
func (c *CherryTime) SetMinute(minute int) {
	c.Time = time.Date(c.Year(), c.Time.Month(), c.Day(), c.Hour(), minute, c.Second(), c.Nanosecond(), c.Location())
}

// SetSecond 设置秒
func (c *CherryTime) SetSecond(second int) {
	c.Time = time.Date(c.Year(), c.Time.Month(), c.Day(), c.Hour(), c.Minute(), second, c.Nanosecond(), c.Location())
}

// SetNanoSecond 设置纳秒
func (c *CherryTime) SetNanoSecond(nanoSecond int) {
	c.Time = time.Date(c.Year(), c.Time.Month(), c.Day(), c.Hour(), c.Minute(), c.Second(), nanoSecond, c.Location())
}
