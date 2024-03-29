/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package types

import (
	"time"
)

var _ Times = (*Now)(nil)

/*
	获取一些日期的函数
	More Time functions
*/

// Times 时间处理函数
type Times interface {
	Now() time.Time
	Monday() time.Time
	Sunday() time.Time
	BeginOfHour() time.Time
	EndOfHour() time.Time
	BeginOfDay() time.Time
	EndOfDay() time.Time
	BeginOfWeek() time.Time
	EndOfWeek() time.Time
	BeginOfMonth() time.Time
	EndOfMonth() time.Time
	BeginOfYear() time.Time
	EndOfYear() time.Time
	BeginOfDuration(d time.Duration) time.Time
	ParseLayoutTime(layout, timestring string) (time.Time, error)
	ParseInLocation(layout, timestring string, loc *time.Location) (time.Time, error)
	WithLocation(loc *time.Location)
	DayOfWeek() int
	WeekOfYear() (int, int)
	Add(time.Duration) Times
	AddDate(years, months, days int) Times
	SetTime(t time.Time) Times
}

// NowOption 执行函数
type NowOption func(*Now)

// NowTime 获取当前时间
func NowTime(t *time.Time) NowOption {
	return func(n *Now) {
		n.Time = t
	}
}

// NowWeekStartDay 设置一周开始时间
func NowWeekStartDay(d time.Weekday) NowOption {
	return func(n *Now) {
		n.Config.WeekStartDay = d
	}
}

// NowLocation 设置时间Location
func NowLocation(loc *time.Location) NowOption {
	return func(n *Now) {
		n.Config.Location = loc
	}
}

// NowConfig 设置时间的配置
func NowConfig(cfg Config) NowOption {
	return func(n *Now) {
		n.Config = cfg
	}
}

// Now the time
type Now struct {
	*time.Time
	Config
}

// Config configuration for now package
type Config struct {
	WeekStartDay time.Weekday
	Location     *time.Location
}

func initConfig() Config {
	return Config{
		WeekStartDay: WeekStartDay,
	}
}

// GetNow initialise by input time
// 初始化当前时间
func GetNow(opts ...NowOption) Times {
	n := &Now{
		Config: initConfig(),
	}

	for _, o := range opts {
		o(n)
	}

	if n.Time == nil {
		t := time.Now()
		n.Time = &t
	}

	return n
}

// BeginOfDuration 以当前的时间作为起始时间，抹掉一部分时间
func BeginOfDuration(d time.Duration) time.Time {
	return GetNow().BeginOfDuration(d)
}

// ParseLayoutTime 解析时间
func ParseLayoutTime(layout, timestring string) (time.Time, error) {
	return GetNow().ParseLayoutTime(layout, timestring)
}

// ParseInLocation 解析时间
func ParseInLocation(layout, timestring string, loc *time.Location) (time.Time, error) {
	return GetNow().ParseInLocation(layout, timestring, loc)
}

// BeginOfHour 当前小时的起始时间
func BeginOfHour() time.Time {
	return GetNow().BeginOfHour()
}

// EndOfDay 当前小时的结束时间
func EndOfHour() time.Time {
	return GetNow().EndOfHour()
}

// BeginOfDay 当前日期的起始时间
func BeginOfDay() time.Time {
	return GetNow().BeginOfDay()
}

// EndOfDay 当前日期的终止时间
func EndOfDay() time.Time {
	return GetNow().EndOfDay()
}

// BeginOfWeek 当前日期的起始时间
func BeginOfWeek() time.Time {
	return GetNow().BeginOfWeek()
}

// EndOfWeek 当前日期的终止时间
func EndOfWeek() time.Time {
	return GetNow().EndOfWeek()
}

// BeginOfMonth 当前月的起始时间
func BeginOfMonth() time.Time {
	return GetNow().BeginOfMonth()
}

// EndOfMonth 当前月的终止时间
func EndOfMonth() time.Time {
	return GetNow().EndOfMonth()
}

// BeginOfYear 当前年的起始时间
func BeginOfYear() time.Time {
	return GetNow().BeginOfYear()
}

// EndOfYear 当前年的终止时间
func EndOfYear() time.Time {
	return GetNow().EndOfYear()
}

// WithLocation 返回带自定义Location的时间
func WithLocation(loc *time.Location) Times {
	return GetNow(NowLocation(loc))
}

///// Times functions /////

// BeginOfDuration 获取当前时间，并抹掉一部分时间
func (p *Now) BeginOfDuration(d time.Duration) time.Time {
	return p.Time.Truncate(d)
}

// WithLocation 设置时区
func (p *Now) WithLocation(loc *time.Location) {
	p.Config.Location = loc
}

// Now 当前时间
func (p *Now) Now() time.Time {
	return *p.Time
}

// ParseLayoutTime 解析时间
func (p *Now) ParseLayoutTime(layout, s string) (time.Time, error) {
	if p.Config.Location == nil {
		return p.ParseInLocation(layout, s, p.Time.Location())
	}
	return p.ParseInLocation(layout, s, p.Config.Location)
}

// ParseInLocation 解析时间
func (*Now) ParseInLocation(layout, timestring string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(layout, timestring, loc)
}

// Monday get monday
func (p *Now) Monday() time.Time {
	t := p.BeginOfDay()
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -weekday+1)
}

// Sunday get sunday
func (p *Now) Sunday() time.Time {
	t := p.BeginOfDay()
	weekday := int(t.Weekday())
	if weekday == 0 {
		return t
	}
	return t.AddDate(0, 0, 7-weekday)
}

// BeginOfHour begin of hour
func (p *Now) BeginOfHour() time.Time {
	y, m, d := p.Date()
	return time.Date(y, m, d, p.Time.Hour(), 0, 0, 0, p.Time.Location())
}

// EndOfHour end of hour
func (p *Now) EndOfHour() time.Time {
	return p.BeginOfHour().Add(time.Hour - time.Nanosecond)
}

// BeginOfDay begin of day
func (p *Now) BeginOfDay() time.Time {
	y, m, d := p.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, p.Time.Location())
}

// EndOfDay end of day
func (p *Now) EndOfDay() time.Time {
	return p.BeginOfDay().AddDate(0, 0, 1).Add(-time.Nanosecond)
}

// BeginOfWeek begin of week
func (p *Now) BeginOfWeek() time.Time {
	t := p.BeginOfDay()
	weekday := int(t.Weekday())
	if p.WeekStartDay != WeekStartDay {
		beginInt := int(p.WeekStartDay)
		if weekday < beginInt {
			weekday = weekday + 7 - beginInt
		} else {
			weekday = weekday - beginInt
		}
	}
	return t.AddDate(0, 0, -weekday)
}

// EndOfWeek end of week
func (p *Now) EndOfWeek() time.Time {
	begin := p.BeginOfWeek()
	return begin.AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// BeginOfMonth begin of month
func (p *Now) BeginOfMonth() time.Time {
	y, m, _ := p.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, p.Time.Location())
}

// EndOfMonth begin of month
func (p *Now) EndOfMonth() time.Time {
	return p.BeginOfMonth().AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// BeginOfYear begin of year
func (p *Now) BeginOfYear() time.Time {
	y, _, _ := p.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, p.Time.Location())
}

// EndOfYear begin of year
func (p *Now) EndOfYear() time.Time {
	return p.BeginOfYear().AddDate(1, 0, 0).Add(-time.Nanosecond)
}

// DayOfWeek day of week
func (p *Now) DayOfWeek() int {
	day := int(p.Weekday())
	beginInt := int(p.WeekStartDay)
	if p.WeekStartDay != WeekStartDay {
		if day < beginInt {
			day = day + 7 - beginInt
		} else {
			day = day - beginInt
		}
	}
	return day + 1
}

// WeekOfYear week of year
// time ISOWeek()
func (p *Now) WeekOfYear() (int, int) {
	return p.Time.ISOWeek()
}

// Add added duration
func (p *Now) Add(duration time.Duration) Times {
	*p.Time = p.Time.Add(duration)
	return p
}

// AddDate added days
func (p *Now) AddDate(years, months, days int) Times {
	*p.Time = p.Time.AddDate(years, months, days)
	return p
}

// SetTime set custom time
func (p *Now) SetTime(t time.Time) Times {
	*p.Time = t
	return p
}
