/*
Copyright © 2017 Henry Huang <hhh@rutcode.com>

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

package config_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"trellis.tech/trellis/common.v0/config"
	"trellis.tech/trellis/common.v0/shell"
	"trellis.tech/trellis/common.v0/testutils"
)

const (
	jsonFile  = "example.json"
	yamlFile  = "example.yml"
	wrongFile = "wrong_file"
)

type TestSlice struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

func TestNewJSONConfig(t *testing.T) {
	c, err := config.NewConfig(wrongFile)
	testutils.NotOk(t, err)
	testutils.Equals(t, c, nil)
	c, err = config.NewConfig("")
	testutils.NotOk(t, err)
	testutils.Equals(t, c, nil)

	err = config.ReadJSONFileToModel(wrongFile, nil)
	testutils.NotOk(t, err)

	whoami := shell.Output("whoami")
	err = os.Setenv("PRE_USER", whoami)
	testutils.Ok(t, err)

	c, err = config.NewConfigOptions(
		config.OptionFile(jsonFile),
		config.OptionENVAllowed(),
		config.OptionENVPrefix("PRE"))
	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "config get nil")

	faceList := c.GetList("b.d")
	testutils.Assert(t, faceList[0] == json.Number("3"), "b.d[0] should be json.Number `3`")
	testutils.Assert(t, faceList[1] == json.Number("4"), "b.d[1] should be json.Number `4`")
	testFunc(t, c)

}

func TestNewYAMLConfig(t *testing.T) {
	c, err := config.NewConfig(wrongFile)
	testutils.NotOk(t, err)
	testutils.Equals(t, c, nil)
	c, err = config.NewConfig("")
	testutils.NotOk(t, err)
	testutils.Equals(t, c, nil)

	err = config.ParseYAMLFileToModel(wrongFile, nil)
	testutils.NotOk(t, err)
	err = os.Setenv("PRE_USER", shell.Output("whoami"))
	testutils.Ok(t, err)

	// c, err := config.NewConfig(yamlFile)
	c, err = config.NewConfigOptions(
		config.OptionFile(yamlFile),
		config.OptionENVAllowed(),
		config.OptionENVPrefix("PRE"))

	if err != nil {
		panic(err)
	}

	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "loaded config should not be nil")
	faceList := c.GetList("b.d")

	testutils.Assert(t, faceList[0] == 3, "b.d[0] should be `3`")
	testutils.Assert(t, faceList[1] == 4, "b.d[1] should be `4`")

	testFunc(t, c)
}

func testFunc(t *testing.T, c config.Config) {

	envUser := os.Getenv("PRE_USER")

	var slices []TestSlice
	err := c.ToObject("s", &slices)
	testutils.Ok(t, err)

	for _, k := range slices {
		if k.Name == "slice_user_right" {
			testutils.Assert(t, k.Value == envUser, "user should be %q", k.Value)
		} else if k.Name == "slice_user_wrong" {
			testutils.Assert(t, k.Value == "", "user should be %q", k.Value)
		}
	}

	testutils.Assert(t, c.GetString("b.u") == "", "env user should be empty")
	testutils.Assert(t, c.GetString("b.pre") == envUser, "env prefix user should be: "+envUser)
	testutils.Assert(t, c.GetString("b.xxx") == "", "b.xxx should be empty")

	testutils.Assert(t, c.GetString("a") == "Easy!", "a should be easy")
	testutils.Assert(t, c.GetMap("a") == nil, "map of a should be nil")
	testutils.Assert(t, c.GetInterface("b.c.cn.a") == "test", "b.c.cn.a should be test")

	_ = c.SetKeyValue("b.c.cn.a", "value")
	newC := c.Copy()
	_ = newC.SetKeyValue("b.c.cn.a", "joking")
	testutils.Assert(t, newC.GetInterface("b.c.cn.a") == "joking", "new b.c.cn.a should be joking")
	testutils.Assert(t, c.GetInterface("b.c.cn.a") == "value", "b.c.cn.a should be value")
	testutils.Assert(t, c.GetInterface("n.a") == "value", "n.a should be value")

	intList := c.GetIntList("b.c.cbd")
	testutils.Assert(t, intList[0] == 3, "b.c.cbd[0] should be 3")
	testutils.Assert(t, intList[1] == 4, "b.c.cbd[1] should be 4")
	testutils.Assert(t, c.GetString("a.b.c", "example") == "example", "a.b.c should be default example")
	testutils.Assert(t, c.GetString("b.c.e") == "Just Do it", "b.c.e should be `Just Do it`")
	testutils.Assert(t, c.GetFloat("h") == 1.01, "h should be  1.01")
	testutils.Assert(t, c.GetInt("b.c.f", 100) == 2, "b.c.f should be 2")
	testutils.Assert(t, c.GetInt("b.c.e") == 0, "b.c.e should be 0")
	testutils.Assert(t, c.GetBoolean("b.c.g"), "b.c.g should be true")
	testutils.Assert(t, c.GetBoolean("b.c.x", true), "b.c.x should be true")
	testutils.Assert(t, c.GetConfig("b") != nil, "b config should not be nil")

	testutils.Assert(t, c.GetTimeDuration("b.c.t") == time.Hour*24, "b.c.t should be 1d")

	_ = c.SetKeyValue("a.b.c", "Correct")
	_ = c.SetKeyValue("b.c.e", "Correct")
	_ = c.SetKeyValue("b.c.d", "d")
	_ = c.SetKeyValue("b.c.g", false)
	_ = c.SetKeyValue("b.d", []int{1, 2, 3, 4})

	testutils.Assert(t, c.GetString("a", "example") == "example", "a should be default example")
	testutils.Assert(t, c.GetInterface("a", "example") != nil, "a should not be nil")
	testutils.Assert(t, c.GetString("a.b.c", "example") == "Correct", "a.b.c should be Correct")
	testutils.Assert(t, c.GetString("b.c.e", "example") == "Correct", "b.c.e should be Correct")
	testutils.Assert(t, !c.GetBoolean("b.c.g", true), "b.c.g should be false")
	testutils.Assert(t, c.GetString("b.c.d", "example") == "d", "b.c.d should be example")

	_ = c.SetKeyValue("a", "Difficult!")
	_ = c.SetKeyValue("h.a", []bool{false, true, false})
	_ = c.SetKeyValue("h.f", []float64{1.2, 2.3, 3.4})
	_ = c.SetKeyValue("h.b", "10T")

	testutils.Assert(t, c.GetString("a.b.c", "example") == "example", "a.b.c should be default example")
	testutils.Assert(t, c.GetString("a", "example") == "Difficult!", "a should be Difficult!")
	testutils.Assert(t, c.GetList("a") == nil, "list of a should be nil")
	testutils.Assert(t, c.GetBooleanList("h.a") != nil, "list of h.a should not be nil")
	testutils.Assert(t, c.GetFloatList("h.f") != nil, "list of h.f should not be nil")
	testutils.Assert(t, c.GetFloat("h") == 0, "h should be empty")

	hb := c.GetByteSize("h.b")
	testutils.Assert(t, hb.Int64() == 10995116277760, "h.b should equals 10995116277760")
	testutils.Assert(t, c.GetString("b.d", "example") == "example", "b.d should be default example")
	testutils.Assert(t, c.GetList("b.d") != nil, "b.d should not be nil")
	testutils.Assert(t, c.GetStringList("b.d") == nil, "string list of b.d should be nil")
	testutils.Assert(t, c.GetIntList("b.d") != nil, "int list of b.d should not be nil")

	_ = c.SetKeyValue("b.d", []string{"1", "2", "3"})

	stringList := c.GetStringList("b.d")
	testutils.Assert(t, stringList[0] == "1", "b.d[0] should be 1")
	testutils.Assert(t, stringList[1] == "2", "b.d[1] should be 2")
	testutils.Assert(t, stringList[2] == "3", "b.d[2] should be 3")

	testutils.Assert(t, c.GetKeys() != nil, "keys should not be nil")
	bs, _ := c.Dump()
	testutils.Assert(t, bs != nil, "dump config should not be nil")
}
