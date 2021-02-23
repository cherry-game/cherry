package cherryTime

import (
	"github.com/cherry-game/cherry/extend/utils"
	"time"
)

// parseByDuration 通过持续时间解析
func ParseByDuration(duration string) (time.Duration, error) {
	td, err := time.ParseDuration(duration)
	if err != nil {
		err = cherryUtils.Errorf("invalid duration %d", duration)
	}
	return td, err
}

// getAbsValue 获取绝对值
func GetAbsValue(value int64) int64 {
	return (value ^ value>>31) - value>>31
}

// AddDurations 按照持续时间字符串增加时间
// 支持整数/浮点数和符号ns(纳秒)、us(微妙)、ms(毫秒)、s(秒)、m(分钟)、h(小时)的组合
func (c CherryTime) AddDuration(duration string) error {
	td, err := ParseByDuration(duration)
	if err != nil {
		return err
	}

	c.Time = c.Time.Add(td)
	return nil
}

// SubDurations 按照持续时间字符串减少时间
// 支持整数/浮点数和符号ns(纳秒)、us(微妙)、ms(毫秒)、s(秒)、m(分钟)、h(小时)的组合
func (c CherryTime) SubDuration(duration string) error {
	return c.AddDuration("-" + duration)
}

// AddCenturies N世纪后
func (c CherryTime) AddCenturies(centuries int) CherryTime {
	return c.AddYears(YearsPerCentury * centuries)
}

// AddCenturiesNoOverflow N世纪后(月份不溢出)
func (c CherryTime) AddCenturiesNoOverflow(centuries int) CherryTime {
	return c.AddYearsNoOverflow(centuries * YearsPerCentury)
}

// AddCentury 1世纪后
func (c CherryTime) AddCentury() CherryTime {
	return c.AddCenturies(1)
}

// AddCenturyNoOverflow 1世纪后(月份不溢出)
func (c CherryTime) AddCenturyNoOverflow() CherryTime {
	return c.AddCenturiesNoOverflow(1)
}

// SubCenturies N世纪前
func (c CherryTime) SubCenturies(centuries int) CherryTime {
	return c.SubYears(YearsPerCentury * centuries)
}

// SubCenturiesNoOverflow N世纪前(月份不溢出)
func (c CherryTime) SubCenturiesNoOverflow(centuries int) CherryTime {
	return c.SubYearsNoOverflow(centuries * YearsPerCentury)
}

// SubCentury 1世纪前
func (c CherryTime) SubCentury() CherryTime {
	return c.SubCenturies(1)
}

// SubCenturyNoOverflow 1世纪前(月份不溢出)
func (c CherryTime) SubCenturyNoOverflow() CherryTime {
	return c.SubCenturiesNoOverflow(1)
}

// AddYears N年后
func (c CherryTime) AddYears(years int) CherryTime {
	c.Time = c.Time.AddDate(years, 0, 0)
	return c
}

// AddYearsNoOverflow N年后(月份不溢出)
func (c CherryTime) AddYearsNoOverflow(years int) CherryTime {
	// 获取N年后本月的最后一天
	last := time.Date(c.Year()+years, c.Time.Month(), 1, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location()).AddDate(0, 1, -1)

	day := c.Day()
	if c.Day() > last.Day() {
		day = last.Day()
	}

	c.Time = time.Date(last.Year(), last.Month(), day, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
	return c
}

// AddYear 1年后
func (c CherryTime) AddYear() CherryTime {
	return c.AddYears(1)
}

// AddYearNoOverflow 1年后(月份不溢出)
func (c CherryTime) AddYearNoOverflow() CherryTime {
	return c.AddYearsNoOverflow(1)
}

// SubYears N年前
func (c CherryTime) SubYears(years int) CherryTime {
	return c.AddYears(-years)
}

// SubYearsNoOverflow N年前(月份不溢出)
func (c CherryTime) SubYearsNoOverflow(years int) CherryTime {
	return c.AddYearsNoOverflow(-years)
}

// SubYear 1年前
func (c CherryTime) SubYear() CherryTime {
	return c.SubYears(1)
}

// SubYearNoOverflow 1年前(月份不溢出)
func (c CherryTime) SubYearNoOverflow() CherryTime {
	return c.SubYearsNoOverflow(1)
}

// AddQuarters N季度后
func (c CherryTime) AddQuarters(quarters int) CherryTime {
	return c.AddMonths(quarters * MonthsPerQuarter)
}

// AddQuartersNoOverflow N季度后(月份不溢出)
func (c CherryTime) AddQuartersNoOverflow(quarters int) CherryTime {
	return c.AddMonthsNoOverflow(quarters * MonthsPerQuarter)
}

// AddQuarter 1季度后
func (c CherryTime) AddQuarter() CherryTime {
	return c.AddQuarters(1)
}

// NextQuarters 1季度后(月份不溢出)
func (c CherryTime) AddQuarterNoOverflow() CherryTime {
	return c.AddQuartersNoOverflow(1)
}

// SubQuarters N季度前
func (c CherryTime) SubQuarters(quarters int) CherryTime {
	return c.AddQuarters(-quarters)
}

// SubQuartersNoOverflow N季度前(月份不溢出)
func (c CherryTime) SubQuartersNoOverflow(quarters int) CherryTime {
	return c.AddMonthsNoOverflow(-quarters * MonthsPerQuarter)
}

// SubQuarter 1季度前
func (c CherryTime) SubQuarter() CherryTime {
	return c.SubQuarters(1)
}

// SubQuarterNoOverflow 1季度前(月份不溢出)
func (c CherryTime) SubQuarterNoOverflow() CherryTime {
	return c.SubQuartersNoOverflow(1)
}

// AddMonths N月后
func (c CherryTime) AddMonths(months int) CherryTime {
	c.Time = c.Time.AddDate(0, months, 0)
	return c
}

// AddMonthsNoOverflow N月后(月份不溢出)
func (c CherryTime) AddMonthsNoOverflow(months int) CherryTime {
	month := c.Time.Month() + time.Month(months)

	// 获取N月后的最后一天
	last := time.Date(c.Year(), month, 1, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location()).AddDate(0, 1, -1)

	day := c.Day()
	if c.Day() > last.Day() {
		day = last.Day()
	}

	c.Time = time.Date(last.Year(), last.Month(), day, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
	return c
}

// AddMonth 1月后
func (c CherryTime) AddMonth() CherryTime {
	return c.AddMonths(1)
}

// AddMonthNoOverflow 1月后(月份不溢出)
func (c CherryTime) AddMonthNoOverflow() CherryTime {
	return c.AddMonthsNoOverflow(1)
}

// SubMonths N月前
func (c CherryTime) SubMonths(months int) CherryTime {
	return c.AddMonths(-months)
}

// SubMonthsNoOverflow N月前(月份不溢出)
func (c CherryTime) SubMonthsNoOverflow(months int) CherryTime {
	return c.AddMonthsNoOverflow(-months)
}

// SubMonth 1月前
func (c CherryTime) SubMonth() CherryTime {
	return c.SubMonths(1)
}

// SubMonthNoOverflow 1月前(月份不溢出)
func (c CherryTime) SubMonthNoOverflow() CherryTime {
	return c.SubMonthsNoOverflow(1)
}

// AddWeeks N周后
func (c CherryTime) AddWeeks(weeks int) CherryTime {
	return c.AddDays(weeks * DaysPerWeek)
}

// AddWeek 1天后
func (c CherryTime) AddWeek() CherryTime {
	return c.AddWeeks(1)
}

// SubWeeks N周后
func (c CherryTime) SubWeeks(weeks int) CherryTime {
	return c.SubDays(weeks * DaysPerWeek)
}

// SubWeek 1天后
func (c CherryTime) SubWeek() CherryTime {
	return c.SubWeeks(1)
}

// AddDays N天后
func (c CherryTime) AddDays(days int) CherryTime {
	c.Time = c.Time.AddDate(0, 0, days)
	return c
}

// AddDay 1天后
func (c CherryTime) AddDay() CherryTime {
	return c.AddDays(1)
}

// SubDays N天前
func (c CherryTime) SubDays(days int) CherryTime {
	return c.AddDays(-days)
}

// SubDay 1天前
func (c CherryTime) SubDay() CherryTime {
	return c.SubDays(1)
}

// AddHours N小时后
func (c CherryTime) AddHours(hours int) CherryTime {
	td := time.Duration(hours) * time.Hour
	c.Time = c.Time.Add(td)
	return c
}

// AddHour 1小时后
func (c CherryTime) AddHour() CherryTime {
	return c.AddHours(1)
}

// SubHours N小时前
func (c CherryTime) SubHours(hours int) CherryTime {
	return c.AddHours(-hours)
}

// SubHour 1小时前
func (c CherryTime) SubHour() CherryTime {
	return c.SubHours(1)
}

// AddMinutes N分钟后
func (c CherryTime) AddMinutes(minutes int) CherryTime {
	td := time.Duration(minutes) * time.Minute
	c.Time = c.Time.Add(td)
	return c
}

// AddMinute 1分钟后
func (c CherryTime) AddMinute() CherryTime {
	return c.AddMinutes(1)
}

// SubMinutes N分钟前
func (c CherryTime) SubMinutes(minutes int) CherryTime {
	return c.AddMinutes(-minutes)
}

// SubMinute 1分钟前
func (c CherryTime) SubMinute() CherryTime {
	return c.SubMinutes(1)
}

// AddSeconds N秒钟后
func (c CherryTime) AddSeconds(seconds int) CherryTime {
	td := time.Duration(seconds) * time.Second
	c.Time = c.Time.Add(td)
	return c
}

// AddSecond 1秒钟后
func (c CherryTime) AddSecond() CherryTime {
	return c.AddSeconds(1)
}

// SubSeconds N秒钟前
func (c CherryTime) SubSeconds(seconds int) CherryTime {
	return c.AddSeconds(-seconds)
}

// SubSecond 1秒钟前
func (c CherryTime) SubSecond() CherryTime {
	return c.SubSeconds(1)
}
