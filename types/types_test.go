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
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	"trellis.tech/trellis/common.v0.1/testutils"
	"trellis.tech/trellis/common.v0.1/types"
)

func TestFlags(t *testing.T) {

	timeNow := time.Now()
	type tTime struct {
		Time     types.Time     `yaml:"time" json:"time"`
		Secret   types.Secret   `yaml:"secret" json:"secret"`
		Duration types.Duration `yaml:"duration" json:"duration"`
	}

	tt := types.Time(timeNow)

	var testTime = &tTime{Time: tt, Secret: "haha", Duration: types.Duration(time.Second * 5)}

	out, err := yaml.Marshal(testTime)
	if err != nil {
		testutils.NotOk(t, err, "failed marshal yaml of time struct")
	}
	testutils.Assert(t, strings.Contains(string(out), types.FormatRFC3339(timeNow)),
		"out of the time: %q-%q", string(out), types.FormatRFC3339(timeNow))
	testutils.Assert(t, strings.Contains(string(out), "<hidden>"), "not hide data : %s", string(out))
	testutils.Assert(t, strings.Contains(string(out), "5s"), "not contains 5s: %s", string(out))

	newTime := &tTime{}
	if err := yaml.Unmarshal(out, newTime); err != nil {
		testutils.NotOk(t, err, "failed unmarshal yaml of time")
	}
	testutils.Assert(t, types.FormatRFC3339(time.Time(newTime.Time)) == newTime.Time.String(),
		"out of the time: %+v - %+v", types.FormatRFC3339(time.Time(newTime.Time)), newTime.Time.String())
	testutils.Assert(t, newTime.Secret == "<hidden>", "is not hidden: %+v", newTime.Secret)
	testutils.Assert(t, testTime.Secret == "haha", "hidden value: %+v", testTime.Secret)
	testutils.Assert(t, testTime.Duration == types.Duration(time.Second*5), "time is not 5s: %+v", testTime.Duration)
}
