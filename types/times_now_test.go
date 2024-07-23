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

package types_test

import (
	"testing"
	"time"

	"trellis.tech/trellis/common.v2/testutils"
	"trellis.tech/trellis/common.v2/types"
)

func TestTimes(t *testing.T) {
	testMondayTime := time.Date(2016, 11, 30, 1, 33, 0, 0, &time.Location{})
	testTuesdayTime := time.Date(2016, 11, 30, 1, 33, 0, 0, &time.Location{})
	testTime := time.Date(2016, 11, 30, 1, 33, 0, 0, &time.Location{})

	tMondayNow := types.GetNow(types.NowTime(&testMondayTime), types.NowWeekStartDay(time.Monday))
	tTuesdayNow := types.GetNow(types.NowTime(&testTuesdayTime), types.NowWeekStartDay(time.Tuesday))
	tNow := types.GetNow(types.NowTime(&testTime), types.NowWeekStartDay(time.Sunday))

	var expt int64 = 1480291200000000000 // 2016-11-28 Mon
	testutils.Equals(t, expt, tMondayNow.BeginOfWeek().UnixNano())
	testutils.Equals(t, expt, tMondayNow.Monday().UnixNano())
	testutils.Equals(t, expt, tNow.Monday().UnixNano())

	expt = 1480809600000000000
	testutils.Equals(t, expt, tMondayNow.Sunday().UnixNano())
	testutils.Equals(t, expt, tNow.Sunday().UnixNano())

	expt = 1480204800000000000 // 2016-11-27
	testutils.Equals(t, expt, tNow.BeginOfWeek().UnixNano())

	expt = 1480895999999999999
	testutils.Equals(t, expt, tMondayNow.EndOfWeek().UnixNano())
	expt = 1480809599999999999
	testutils.Equals(t, expt, tNow.EndOfWeek().UnixNano())

	expt = 1477958400000000000 // 2016-11-1
	testutils.Equals(t, expt, tMondayNow.BeginOfMonth().UnixNano())
	testutils.Equals(t, expt, tNow.BeginOfMonth().UnixNano())

	expt = 1480550399999999999 // 2016-11-30
	testutils.Equals(t, expt, tMondayNow.EndOfMonth().UnixNano())
	testutils.Equals(t, expt, tNow.EndOfMonth().UnixNano())

	expt = 1451606400000000000 // 2016-1-1
	testutils.Equals(t, expt, tMondayNow.BeginOfYear().UnixNano())
	testutils.Equals(t, expt, tNow.BeginOfYear().UnixNano())

	expt = 1483228799999999999 // 2016-12-31
	testutils.Equals(t, expt, tMondayNow.EndOfYear().UnixNano())
	testutils.Equals(t, expt, tNow.EndOfYear().UnixNano())

	testutils.Equals(t, "2016-11-30 01:33:00", types.FormatDateTime(tNow.Now()))

	pst, err := time.LoadLocation("America/Los_Angeles")
	testutils.Ok(t, err)
	tNow.WithLocation(pst)

	r, err := tNow.ParseLayoutTime(types.DateTime, "2016-11-30 01:33:00")
	testutils.Ok(t, err)
	zString, offset := r.Zone()
	testutils.Equals(t, "PST", zString)
	testutils.Equals(t, -28800, offset)

	testutils.Equals(t, 3, tMondayNow.DayOfWeek())
	testutils.Equals(t, 3, int(tMondayNow.Now().Weekday()))
	testutils.Equals(t, 1, tMondayNow.Add(-time.Hour*24*2).DayOfWeek()) // Monday is 1
	year, week := tMondayNow.WeekOfYear()
	testutils.Equals(t, 2016, year)
	testutils.Equals(t, 48, week)
	testutils.Equals(t, 7, tMondayNow.Add(time.Hour*24*6).DayOfWeek()) // 1+6 = 7 Sunday is 7
	year, week = tMondayNow.WeekOfYear()
	testutils.Equals(t, 2016, year)
	testutils.Equals(t, 48, week)
	testutils.Equals(t, 1, tMondayNow.Add(time.Hour*24).DayOfWeek()) // 7+1 = 1 Monday is 1
	year, week = tMondayNow.WeekOfYear()
	testutils.Equals(t, 2016, year)
	testutils.Equals(t, 49, week)

	testutils.Equals(t, 3, int(tTuesdayNow.Now().Weekday()))
	testutils.Equals(t, 0, int(tTuesdayNow.Sunday().Weekday()))
	testutils.Equals(t, 2, tTuesdayNow.DayOfWeek()) // Wednesday is 2
	year, week = tTuesdayNow.WeekOfYear()
	testutils.Equals(t, 2016, year)
	testutils.Equals(t, 48, week)
	testutils.Equals(t, 1, tTuesdayNow.Add(-time.Hour*24).DayOfWeek()) // Tuesday is 1
	testutils.Equals(t, 7, tTuesdayNow.Add(-time.Hour*24).DayOfWeek()) // Monday is 7
	year, week = tTuesdayNow.WeekOfYear()
	testutils.Equals(t, 2016, year)
	testutils.Equals(t, 48, week)
	testutils.Equals(t, 6, tTuesdayNow.Add(-time.Hour*24).DayOfWeek()) // Sunday is 6
	year, week = tTuesdayNow.WeekOfYear()
	testutils.Equals(t, 2016, year)
	testutils.Equals(t, 47, week)
	testutils.Equals(t, 0, int(tMondayNow.Sunday().Weekday()))
	testutils.Equals(t, 0, int(tNow.Sunday().Weekday()))
	testutils.Equals(t, 1, int(tNow.Monday().Weekday()))
	testutils.Equals(t, 4, tNow.DayOfWeek())

	testHourTime := time.Date(2023, 1, 29, 16, 15, 2, 0, &time.Location{})
	tHourNow := types.GetNow(types.NowTime(&testHourTime))

	tHBegin := tHourNow.BeginOfHour()
	testutils.Equals(t, 29, tHBegin.Day())
	testutils.Equals(t, 29, tHBegin.Day())
	testutils.Equals(t, 16, tHBegin.Hour())
	testutils.Equals(t, 0, tHBegin.Minute())
	testutils.Equals(t, 0, tHBegin.Second())

	tHEnd := tHourNow.EndOfHour()
	testutils.Equals(t, 29, tHEnd.Day())
	testutils.Equals(t, 29, tHEnd.Day())
	testutils.Equals(t, 16, tHEnd.Hour())
	testutils.Equals(t, 59, tHEnd.Minute())
	testutils.Equals(t, 59, tHEnd.Second())
	testutils.Equals(t, 999999999, tHEnd.Nanosecond())
}
