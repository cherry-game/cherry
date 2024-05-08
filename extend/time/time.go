// Package cherryTime code from: https://github.com/golang-module/carbon
package cherryTime

import (
	"strconv"
	"time"

	cerr "github.com/cherry-game/cherry/error"
)

const (
	YearsPerMillennium         = 1000     // 每千年1000年
	YearsPerCentury            = 100      // 每世纪100年
	YearsPerDecade             = 10       // 每十年10年
	QuartersPerYear            = 4        // 每年4季度
	MonthsPerYear              = 12       // 每年12月
	MonthsPerQuarter           = 3        // 每季度3月
	WeeksPerNormalYear         = 52       // 每常规年52周
	weeksPerLongYear           = 53       // 每长年53周
	WeeksPerMonth              = 4        // 每月4周
	DaysPerLeapYear            = 366      // 每闰年366天
	DaysPerNormalYear          = 365      // 每常规年365天
	DaysPerWeek                = 7        // 每周7天
	HoursPerWeek               = 168      // 每周168小时
	HoursPerDay                = 24       // 每天24小时
	MinutesPerDay              = 1440     // 每天1440分钟
	MinutesPerHour             = 60       // 每小时60分钟
	SecondsPerWeek             = 604800   // 每周604800秒
	SecondsPerDay              = 86400    // 每天86400秒
	SecondsPerHour             = 3600     // 每小时3600秒
	SecondsPerMinute           = 60       // 每分钟60秒
	MillisecondsPerSecond      = 1000     // 每秒1000毫秒
	MillisecondsPerMinute      = 60000    // 每分钟60000毫秒
	MillisecondsPerDay         = 86400000 // 每天86400000毫秒
	MillisecondsPerHour        = 3600000  // 每小时3600000毫秒
	MicrosecondsPerMillisecond = 1000     // 每毫秒1000微秒
	MicrosecondsPerSecond      = 1000000  // 每秒1000000微秒

	WeeksPerLongYear = 53 // 每长年53周

	DateTimeMillisecondFormat = "2006-01-02 15:04:05.000"
	DateTimeFormat            = "2006-01-02 15:04:05"
	DateFormat                = "2006-01-02"
	TimeFormat                = "15:04:05"
	ShortDateTimeFormat       = "20060102150405"
	ShortDateFormat           = "20060102"
	ShortTimeFormat           = "150405"
)

type CherryTime struct {
	time.Time
}

func NewTime(tt time.Time, setGlobal bool) CherryTime {
	ct := CherryTime{
		Time: tt,
	}

	if setGlobal {
		ct.Time = tt.In(offsetLocation).Add(offsetTime)
	}
	return ct
}

func NewSecond(second int64) CherryTime {
	return NewTime(time.Unix(second, 0), true)
}

func NewMillisecond(millisecond int64) CherryTime {
	return NewTime(time.UnixMilli(millisecond), true)
}

func Now() CherryTime {
	return NewTime(time.Now(), true)
}

func Yesterday() CherryTime {
	t := time.Now().AddDate(0, 0, -1)
	return NewTime(t, true)
}

func Tomorrow() CherryTime {
	t := Now().AddDate(0, 0, 1)
	return NewTime(t, true)
}

// CreateFromTimestamp 从时间戳创建 Carbon 实例
func CreateFromTimestamp(timestamp int64) CherryTime {
	var ts int64

	switch len(strconv.FormatInt(timestamp, 10)) {
	case 10:
		ts = timestamp
	case 13:
		ts = timestamp / 1e3
	case 16:
		ts = timestamp / 1e6
	case 19:
		ts = timestamp / 1e9
	default:
		ts = 0
	}

	t := time.Unix(ts, 0)
	return NewTime(t, false)
}

// CreateFromDateTime 从年月日时分秒创建 Carbon 实例
func CreateFromDateTime(year int, month int, day int, hour int, minute int, second int) CherryTime {
	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, offsetLocation)
	return NewTime(t, false)
}

// CreateFromDate 从年月日创建 Carbon 实例(默认时区)
func CreateFromDate(year int, month int, day int) CherryTime {
	now := Now()
	t := time.Date(year, time.Month(month), day, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	return NewTime(t, false)
}

// CreateFromTime 从时分秒创建 Carbon 实例(默认时区)
func CreateFromTime(hour int, minute int, second int) CherryTime {
	now := Now()
	t := time.Date(now.Year(), now.Time.Month(), now.Day(), hour, minute, second, 0, now.Location())
	return NewTime(t, false)
}

// ParseByDuration 通过持续时间解析
func ParseByDuration(duration string) (time.Duration, error) {
	td, err := time.ParseDuration(duration)
	if err != nil {
		err = cerr.Errorf("invalid duration %s", duration)
	}
	return td, err
}

// GetAbsValue 获取绝对值
func GetAbsValue(value int64) int64 {
	return (value ^ value>>31) - value>>31
}
