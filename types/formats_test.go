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

	"trellis.tech/common.v2/testutils"
	"trellis.tech/common.v2/types"
)

func Test_ParseStringByteSize(t *testing.T) {

	s := "10b"
	var expI int64 = 10
	bInt := types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10kb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10mb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10gb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10tb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10pb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10eb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10zb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())
	s = "10yb"
	expI *= 1000
	bInt = types.ParseStringByteSize(s)
	testutils.Equals(t, expI, bInt.Int64())

	si := "10k"
	var expIi int64 = 10240
	biInt := types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
	si = "10m"
	expIi *= 1024
	biInt = types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
	si = "10g"
	expIi *= 1024
	biInt = types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
	si = "10t"
	expIi *= 1024
	biInt = types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
	si = "10p"
	expIi *= 1024
	biInt = types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
	si = "10e"
	expIi *= 1024
	biInt = types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
	si = "10z"
	expIi *= 1024
	biInt = types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
	si = "10y"
	expIi *= 1024
	biInt = types.ParseStringByteSize(si)
	testutils.Equals(t, expIi, biInt.Int64())
}

func Test_ParseStringTime(t *testing.T) {
	s := "1s"
	dt := types.ParseStringTime(s)
	expt := time.Second
	testutils.Equals(t, expt, dt)

	s = "10m"
	dt = types.ParseStringTime(s)
	expt = time.Minute * 10
	testutils.Equals(t, expt, dt)
	exptDurtion := types.Duration(expt)
	testutils.Equals(t, "10m", exptDurtion.String())

	s = "3h"
	dt = types.ParseStringTime(s)
	expt = time.Hour * 3
	testutils.Equals(t, expt, dt)
	exptDurtion = types.Duration(expt)
	testutils.Equals(t, "3h", exptDurtion.String())

	s = "1d"
	dt = types.ParseStringTime(s)
	expt = time.Hour * 24
	testutils.Equals(t, expt, dt)
	exptDurtion = types.Duration(expt)
	testutils.Equals(t, "1d", exptDurtion.String())

	s = "5ns"
	dt = types.ParseStringTime(s)
	expt = time.Nanosecond * 5
	testutils.Equals(t, expt, dt)
	exptDurtion = types.Duration(expt)
	testutils.Equals(t, "5ns", exptDurtion.String())

	s = "10us"
	dt = types.ParseStringTime(s)
	expt = time.Microsecond * 10
	testutils.Equals(t, expt, dt)
	exptDurtion = types.Duration(expt)
	testutils.Equals(t, "10us", exptDurtion.String())

	s = "1ms"
	dt = types.ParseStringTime(s)
	expt = time.Millisecond
	testutils.Equals(t, expt, dt)
	exptDurtion = types.Duration(expt)
	testutils.Equals(t, "1ms", exptDurtion.String())

	s = "2weeks"
	dt = types.ParseStringTime(s)
	expt = 24 * 2 * 7 * time.Hour
	testutils.Equals(t, expt, dt)
	exptDurtion = types.Duration(expt)
	testutils.Equals(t, "2w", exptDurtion.String())

	s = "2years"
	dt = types.ParseStringTime(s)
	expt = 24 * time.Hour * 365 * 2
	testutils.Equals(t, expt, dt)
	exptDurtion = types.Duration(expt)
	testutils.Equals(t, "2y", exptDurtion.String())
}
