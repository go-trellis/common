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
	"flag"
	"testing"

	"github.com/go-trellis/common/utils/testutils"
	"gopkg.in/yaml.v3"
)

func TestSecret_String(t *testing.T) {
	s := Secret("my-secret")
	testutils.Equals(t, "my-secret", s.String(), "String should return actual value")
}

func TestSecret_Set(t *testing.T) {
	var s Secret
	err := s.Set("new-secret")
	testutils.Ok(t, err)
	testutils.Equals(t, Secret("new-secret"), s, "Set should set the value")
}

func TestSecret_MarshalYAML(t *testing.T) {
	s := Secret("my-secret")
	result, err := s.MarshalYAML()
	testutils.Ok(t, err)
	testutils.Equals(t, Hidden, result, "MarshalYAML should return hidden value")
}

func TestSecret_MarshalYAML_Empty(t *testing.T) {
	s := Secret("")
	result, err := s.MarshalYAML()
	testutils.Ok(t, err)
	testutils.Equals(t, "", result, "MarshalYAML should return empty for empty secret")
}

func TestSecret_UnmarshalYAML(t *testing.T) {
	var s Secret
	err := yaml.Unmarshal([]byte(`"my-secret"`), &s)
	testutils.Ok(t, err)
	testutils.Equals(t, Secret("my-secret"), s, "UnmarshalYAML should set the value")
}

func TestSecret_FlagValue(t *testing.T) {
	var s Secret
	testutils.Assert(t, flag.Value(&s) != nil, "Secret should implement flag.Value")

	err := s.Set("flag-secret")
	testutils.Ok(t, err)
	testutils.Equals(t, "flag-secret", s.String(), "String should return set value")
}

func TestSecret_MarshalJSON(t *testing.T) {
	s := Secret("my-secret")
	// Secret doesn't implement MarshalJSON directly, test through yaml
	data, err := yaml.Marshal(&s)
	testutils.Ok(t, err)
	testutils.Assert(t, len(data) > 0, "should marshal to YAML")
}
