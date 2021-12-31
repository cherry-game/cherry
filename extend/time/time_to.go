package cherryTime

import (
	cherryString "github.com/cherry-game/cherry/extend/string"
	"time"
)

// ToTimestamp 输出秒级时间戳
func (c CherryTime) ToTimestamp() int64 {
	return c.ToTimestampWithSecond()
}

// ToTimestampWithSecond 输出秒级时间戳
func (c CherryTime) ToTimestampWithSecond() int64 {
	return c.Unix()
}

// ToMillisecond 输出毫秒级时间戳
func (c CherryTime) ToMillisecond() int64 {
	return c.ToTimestampWithMillisecond()
}

// ToTimestampWithMillisecond 输出毫秒级时间戳
func (c CherryTime) ToTimestampWithMillisecond() int64 {
	return c.Time.UnixNano() / int64(time.Millisecond)
}

// ToTimestampWithMicrosecond 输出微秒级时间戳
func (c CherryTime) ToTimestampWithMicrosecond() int64 {
	return c.UnixNano() / int64(time.Microsecond)
}

// ToTimestampWithNanosecond 输出纳秒级时间戳
func (c CherryTime) ToTimestampWithNanosecond() int64 {
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
	intDate, _ := cherryString.ToInt32(strDate)
	return intDate
}

// ToShortTimeFormat 150405
func (c CherryTime) ToShortTimeFormat() string {
	return c.Format(ShortTimeFormat)
}
