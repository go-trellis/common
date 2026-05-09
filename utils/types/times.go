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
	"net/http"
	"time"

	"github.com/go-trellis/common.v3/errors/errcode"
)

// Datas
const (
	Date  = "2006-01-02"
	ZDate = "2006-1-2"

	TimeStr  = "15:04:05"
	DashTime = Date + "-15-04-05"
	DateTime = Date + " " + TimeStr

	ChineseDate      = "2006年01月02日"
	ChineseZDate     = "2006年1月2日"
	ChineseDateTime  = "2006年01月02日15时04分05秒"
	ChineseZDateTime = "2006年1月2日15时4分5秒"

	DefaultDateTime = "0001-01-01 00:00:00"
)

// MonthDays
const (
	MonthLunarDays   int = 30
	MonthSolarDays   int = 31
	MonthFebLeapDays int = 29
	MonthFebDays     int = 28
)

const (
	// WeekStartDay start day of week
	WeekStartDay = time.Sunday
)

// FormatLayoutTime format layout time to string time.
func FormatLayoutTime(t time.Time, layout string) string {
	return t.Format(layout)
}

///// Chinese display format /////
///// Format time to Chinese string time /////

// FormatChineseDate format layout chinese date to string date.
func FormatChineseDate(t time.Time) string {
	return FormatLayoutTime(t, ChineseDate)
}

// FormatChineseZDate format layout chinese zdate to string zdate.
func FormatChineseZDate(t time.Time) string {
	return FormatLayoutTime(t, ChineseZDate)
}

// FormatChineseDateTime format layout chinese datetime to string datetime.
func FormatChineseDateTime(t time.Time) string {
	return FormatLayoutTime(t, ChineseDateTime)
}

// FormatChineseZDateTime format layout chinese zdatetime to string zdatetime.
func FormatChineseZDateTime(t time.Time) string {
	return FormatLayoutTime(t, ChineseZDateTime)
}

///// English display format /////
///// Format time to string /////

// FormatDate format date string.
func FormatDate(t time.Time) string {
	return FormatLayoutTime(t, Date)
}

// FormatZDate format zdate string.
func FormatZDate(t time.Time) string {
	return FormatLayoutTime(t, ZDate)
}

// FormatTime format time string
func FormatTime(t time.Time) string {
	return FormatLayoutTime(t, TimeStr)
}

// FormatDateTime format datetime string
func FormatDateTime(t time.Time) string {
	return FormatLayoutTime(t, DateTime)
}

// FormatDashTime format datetime string with dash
func FormatDashTime(t time.Time) string {
	return FormatLayoutTime(t, DashTime)
}

// FormatRFC3339 format RFC3339 string
func FormatRFC3339(t time.Time) string {
	return FormatLayoutTime(t, time.RFC3339)
}

// FormatRFC3339Nano format RFC3339Nano string
func FormatRFC3339Nano(t time.Time) string {
	return FormatLayoutTime(t, time.RFC3339Nano)
}

// FormatHTTPGMT format GMT string
func FormatHTTPGMT(t time.Time) string {
	return FormatLayoutTime(t, http.TimeFormat)
}

// IsZero judge time is zero
func IsZero(t time.Time) bool {
	return t.IsZero() || FormatTime(t) == DefaultDateTime
}

// GetTimeMonthDays get time's month days
func GetTimeMonthDays(t time.Time) int {
	return GetMonthDays(t.Year(), int(t.Month()))
}

// GetMonthDays get year's month days
func GetMonthDays(year, month int) int {
	switch month {
	case 4, 6, 9, 11:
		return MonthLunarDays
	case 1, 3, 5, 7, 8, 10, 12:
		return MonthSolarDays
	case 2:
		if ((year%4 == 0) && (year%100 != 0)) || (year%400) == 0 {
			return MonthFebLeapDays
		}
		return MonthFebDays
	}
	return 0
}

///// Convert string to time /////
///// Parse string to time /////

// StringToDate parse string to date, but is deprecated, use ParseDate
func StringToDate(t string) (time.Time, error) {
	return time.Parse(Date, t)
}

// StringToDateTime parse string to datetime, but is deprecated, use ParseDateTime
func StringToDateTime(t string) (time.Time, error) {
	return time.Parse(DateTime, t)
}

// ParseDate parse string to date using layout.
func ParseDate(t string) (time.Time, error) {
	return ParseLayoutTime(Date, t)
}

// ParseDateTime parse string to datetime using layout.
func ParseDateTime(t string) (time.Time, error) {
	return ParseLayoutTime(DateTime, t)
}

// ParseChineseDate parse string to chinese date using layout.
func ParseChineseDate(t string) (time.Time, error) {
	return ParseLayoutTime(Date, t)
}

// ParseChineseDateTime parse string to chinese datetime using layout.
func ParseChineseDateTime(t string) (time.Time, error) {
	return ParseLayoutTime(ChineseDateTime, t)
}

func Parse(s string) (*time.Time, error) {
	if s == "0" {
		return &time.Time{}, nil
	}

	p, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &p, nil
	}

	p, err = time.Parse("2006-01-02T15:04", s)
	if err == nil {
		return &p, nil
	}

	p, err = time.Parse("2006-01-02", s)
	if err == nil {
		return &p, nil
	}

	return nil, errcode.Newf("failed to parse time: %q", s)
}

// // ParseInLocation parse datetime in local
// func ParseInLocation(t, layout string, local *time.Location) (time.Time, error) {
// 	return time.ParseInLocation(DateTime, t, local)
// }

// UnixToTime parse unix to time
func UnixToTime(unix int64) time.Time {
	return time.Unix(unix, 0)
}
