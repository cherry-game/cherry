package cherryTime

import (
	"time"
)

// AddDuration 按照持续时间字符串增加时间
// 支持整数/浮点数和符号ns(纳秒)、us(微妙)、ms(毫秒)、s(秒)、m(分钟)、h(小时)的组合
func (c *CherryTime) AddDuration(duration string) error {
	td, err := ParseByDuration(duration)
	if err != nil {
		return err
	}

	c.Time = c.Add(td)
	return nil
}

// SubDuration 按照持续时间字符串减少时间
// 支持整数/浮点数和符号ns(纳秒)、us(微妙)、ms(毫秒)、s(秒)、m(分钟)、h(小时)的组合
func (c *CherryTime) SubDuration(duration string) error {
	return c.AddDuration("-" + duration)
}

// AddCenturies N世纪后
func (c *CherryTime) AddCenturies(centuries int) {
	c.AddYears(YearsPerCentury * centuries)
}

// AddCenturiesNoOverflow N世纪后(月份不溢出)
func (c *CherryTime) AddCenturiesNoOverflow(centuries int) {
	c.AddYearsNoOverflow(centuries * YearsPerCentury)
}

// AddCentury 1世纪后
func (c *CherryTime) AddCentury() {
	c.AddCenturies(1)
}

// AddCenturyNoOverflow 1世纪后(月份不溢出)
func (c *CherryTime) AddCenturyNoOverflow() {
	c.AddCenturiesNoOverflow(1)
}

// SubCenturies N世纪前
func (c *CherryTime) SubCenturies(centuries int) {
	c.SubYears(YearsPerCentury * centuries)
}

// SubCenturiesNoOverflow N世纪前(月份不溢出)
func (c *CherryTime) SubCenturiesNoOverflow(centuries int) {
	c.SubYearsNoOverflow(centuries * YearsPerCentury)
}

// SubCentury 1世纪前
func (c *CherryTime) SubCentury() {
	c.SubCenturies(1)
}

// SubCenturyNoOverflow 1世纪前(月份不溢出)
func (c *CherryTime) SubCenturyNoOverflow() {
	c.SubCenturiesNoOverflow(1)
}

// AddYears N年后
func (c *CherryTime) AddYears(years int) {
	c.Time = c.Time.AddDate(years, 0, 0)
}

// AddYearsNoOverflow N年后(月份不溢出)
func (c *CherryTime) AddYearsNoOverflow(years int) {
	// 获取N年后本月的最后一天
	last := time.Date(c.Year()+years, c.Time.Month(), 1, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location()).AddDate(0, 1, -1)

	day := c.Day()
	if c.Day() > last.Day() {
		day = last.Day()
	}

	c.Time = time.Date(last.Year(), last.Month(), day, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
}

// AddYear 1年后
func (c *CherryTime) AddYear() {
	c.AddYears(1)
}

// AddYearNoOverflow 1年后(月份不溢出)
func (c *CherryTime) AddYearNoOverflow() {
	c.AddYearsNoOverflow(1)
}

// SubYears N年前
func (c *CherryTime) SubYears(years int) {
	c.AddYears(-years)
}

// SubYearsNoOverflow N年前(月份不溢出)
func (c *CherryTime) SubYearsNoOverflow(years int) {
	c.AddYearsNoOverflow(-years)
}

// SubYear 1年前
func (c *CherryTime) SubYear() {
	c.SubYears(1)
}

// SubYearNoOverflow 1年前(月份不溢出)
func (c *CherryTime) SubYearNoOverflow() {
	c.SubYearsNoOverflow(1)
}

// AddQuarters N季度后
func (c *CherryTime) AddQuarters(quarters int) {
	c.AddMonths(quarters * MonthsPerQuarter)
}

// AddQuartersNoOverflow N季度后(月份不溢出)
func (c *CherryTime) AddQuartersNoOverflow(quarters int) {
	c.AddMonthsNoOverflow(quarters * MonthsPerQuarter)
}

// AddQuarter 1季度后
func (c *CherryTime) AddQuarter() {
	c.AddQuarters(1)
}

// AddQuarterNoOverflow 1季度后(月份不溢出)
func (c *CherryTime) AddQuarterNoOverflow() {
	c.AddQuartersNoOverflow(1)
}

// SubQuarters N季度前
func (c *CherryTime) SubQuarters(quarters int) {
	c.AddQuarters(-quarters)
}

// SubQuartersNoOverflow N季度前(月份不溢出)
func (c *CherryTime) SubQuartersNoOverflow(quarters int) {
	c.AddMonthsNoOverflow(-quarters * MonthsPerQuarter)
}

// SubQuarter 1季度前
func (c *CherryTime) SubQuarter() {
	c.SubQuarters(1)
}

// SubQuarterNoOverflow 1季度前(月份不溢出)
func (c *CherryTime) SubQuarterNoOverflow() {
	c.SubQuartersNoOverflow(1)
}

// AddMonths N月后
func (c *CherryTime) AddMonths(months int) {
	c.Time = c.Time.AddDate(0, months, 0)
}

// AddMonthsNoOverflow N月后(月份不溢出)
func (c *CherryTime) AddMonthsNoOverflow(months int) {
	month := c.Time.Month() + time.Month(months)

	// 获取N月后的最后一天
	last := time.Date(c.Year(), month, 1, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location()).AddDate(0, 1, -1)

	day := c.Day()
	if c.Day() > last.Day() {
		day = last.Day()
	}

	c.Time = time.Date(last.Year(), last.Month(), day, c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.Location())
}

// AddMonth 1月后
func (c *CherryTime) AddMonth() {
	c.AddMonths(1)
}

// AddMonthNoOverflow 1月后(月份不溢出)
func (c *CherryTime) AddMonthNoOverflow() {
	c.AddMonthsNoOverflow(1)
}

// SubMonths N月前
func (c *CherryTime) SubMonths(months int) {
	c.AddMonths(-months)
}

// SubMonthsNoOverflow N月前(月份不溢出)
func (c *CherryTime) SubMonthsNoOverflow(months int) {
	c.AddMonthsNoOverflow(-months)
}

// SubMonth 1月前
func (c *CherryTime) SubMonth() {
	c.SubMonths(1)
}

// SubMonthNoOverflow 1月前(月份不溢出)
func (c *CherryTime) SubMonthNoOverflow() {
	c.SubMonthsNoOverflow(1)
}

// AddWeeks N周后
func (c *CherryTime) AddWeeks(weeks int) {
	c.AddDays(weeks * DaysPerWeek)
}

// AddWeek 1天后
func (c *CherryTime) AddWeek() {
	c.AddWeeks(1)
}

// SubWeeks N周后
func (c *CherryTime) SubWeeks(weeks int) {
	c.SubDays(weeks * DaysPerWeek)
}

// SubWeek 1天后
func (c *CherryTime) SubWeek() {
	c.SubWeeks(1)
}

// AddDays N天后
func (c *CherryTime) AddDays(days int) {
	c.Time = c.Time.AddDate(0, 0, days)
}

// AddDay 1天后
func (c *CherryTime) AddDay() {
	c.AddDays(1)
}

// SubDays N天前
func (c *CherryTime) SubDays(days int) {
	c.AddDays(-days)
}

// SubDay 1天前
func (c *CherryTime) SubDay() {
	c.SubDays(1)
}

// AddHours N小时后
func (c *CherryTime) AddHours(hours int) {
	td := time.Duration(hours) * time.Hour
	c.Time = c.Time.Add(td)
}

// AddHour 1小时后
func (c *CherryTime) AddHour() {
	c.AddHours(1)
}

// SubHours N小时前
func (c *CherryTime) SubHours(hours int) {
	c.AddHours(-hours)
}

// SubHour 1小时前
func (c *CherryTime) SubHour() {
	c.SubHours(1)
}

// AddMinutes N分钟后
func (c *CherryTime) AddMinutes(minutes int) {
	td := time.Duration(minutes) * time.Minute
	c.Time = c.Time.Add(td)
}

// AddMinute 1分钟后
func (c *CherryTime) AddMinute() {
	c.AddMinutes(1)
}

// SubMinutes N分钟前
func (c *CherryTime) SubMinutes(minutes int) {
	c.AddMinutes(-minutes)
}

// SubMinute 1分钟前
func (c *CherryTime) SubMinute() {
	c.SubMinutes(1)
}

// AddSeconds N秒钟后
func (c *CherryTime) AddSeconds(seconds int) {
	td := time.Duration(seconds) * time.Second
	c.Time = c.Time.Add(td)
}

// AddSecond 1秒钟后
func (c *CherryTime) AddSecond() {
	c.AddSeconds(1)
}

// SubSeconds N秒钟前
func (c *CherryTime) SubSeconds(seconds int) {
	c.AddSeconds(-seconds)
}

// SubSecond 1秒钟前
func (c *CherryTime) SubSecond() {
	c.SubSeconds(1)
}
