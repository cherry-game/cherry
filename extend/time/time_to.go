package cherryTime

import (
	cstring "github.com/cherry-game/cherry/extend/string"
	"time"
)

// ToSecond 输出秒级时间戳
func (c CherryTime) ToSecond() int64 {
	return c.Unix()
}

// ToMillisecond 输出毫秒级时间戳
func (c CherryTime) ToMillisecond() int64 {
	return c.Time.UnixNano() / int64(time.Millisecond)
}

func (c CherryTime) ToMillisecondString() string {
	t := c.ToMillisecond()
	return cstring.ToString(t)
}

// ToMicrosecond 输出微秒级时间戳
func (c CherryTime) ToMicrosecond() int64 {
	return c.UnixNano() / int64(time.Microsecond)
}

// ToNanosecond 输出纳秒级时间戳
func (c CherryTime) ToNanosecond() int64 {
	return c.UnixNano()
}

// ToDateTimeFormat 2006-01-02 15:04:05
func (c CherryTime) ToDateTimeFormat() string {
	return c.Format(DateTimeFormat)
}

// ToDateFormat 2006-01-02
func (c CherryTime) ToDateFormat() string {
	return c.Format(DateFormat)
}

// ToTimeFormat 15:04:05
func (c CherryTime) ToTimeFormat() string {
	return c.Format(TimeFormat)
}

//ToShortDateTimeFormat 20060102150405
func (c CherryTime) ToShortDateTimeFormat() string {
	return c.Format(ShortDateTimeFormat)
}

// ToShortDateFormat 20060102
func (c CherryTime) ToShortDateFormat() string {
	return c.Format(ShortDateFormat)
}

// ToShortIntDateFormat 20060102
func (c CherryTime) ToShortIntDateFormat() int32 {
	strDate := c.ToShortDateFormat()
	intDate, _ := cstring.ToInt32(strDate)
	return intDate
}

// ToShortTimeFormat 150405
func (c CherryTime) ToShortTimeFormat() string {
	return c.Format(ShortTimeFormat)
}
