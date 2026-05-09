/*
Copyright © 2015 Prometheus Team
Licensed under the Apache License, Version 2.0 (the "License");

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
	"encoding/json"
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
	"gopkg.in/yaml.v3"
)

func TestHostPort_UnmarshalYAML(t *testing.T) {
	var hp HostPort
	err := yaml.Unmarshal([]byte(`"localhost:8080"`), &hp)
	testutils.Ok(t, err)
	testutils.Equals(t, "localhost", hp.Host, "Host should match")
	testutils.Equals(t, "8080", hp.Port, "Port should match")
}

func TestHostPort_UnmarshalYAML_Empty(t *testing.T) {
	var hp HostPort
	err := yaml.Unmarshal([]byte(`""`), &hp)
	testutils.Ok(t, err)
	testutils.Equals(t, "", hp.Host, "Host should be empty")
	testutils.Equals(t, "", hp.Port, "Port should be empty")
}

func TestHostPort_UnmarshalYAML_Invalid(t *testing.T) {
	var hp HostPort
	err := yaml.Unmarshal([]byte(`"invalid"`), &hp)
	testutils.NotOk(t, err, "should return error for invalid address")
}

func TestHostPort_UnmarshalYAML_NoPort(t *testing.T) {
	var hp HostPort
	err := yaml.Unmarshal([]byte(`"localhost:"`), &hp)
	testutils.NotOk(t, err, "should return error for missing port")
}

func TestHostPort_UnmarshalJSON(t *testing.T) {
	var hp HostPort
	err := json.Unmarshal([]byte(`"localhost:8080"`), &hp)
	testutils.Ok(t, err)
	testutils.Equals(t, "localhost", hp.Host, "Host should match")
	testutils.Equals(t, "8080", hp.Port, "Port should match")
}

func TestHostPort_UnmarshalJSON_Empty(t *testing.T) {
	var hp HostPort
	err := json.Unmarshal([]byte(`""`), &hp)
	testutils.Ok(t, err)
	testutils.Equals(t, "", hp.Host, "Host should be empty")
	testutils.Equals(t, "", hp.Port, "Port should be empty")
}

func TestHostPort_UnmarshalJSON_Invalid(t *testing.T) {
	var hp HostPort
	err := json.Unmarshal([]byte(`"invalid"`), &hp)
	testutils.NotOk(t, err, "should return error for invalid address")
}

func TestHostPort_MarshalYAML(t *testing.T) {
	hp := HostPort{Host: "localhost", Port: "8080"}
	data, err := yaml.Marshal(&hp)
	testutils.Ok(t, err)
	testutils.Assert(t, len(data) > 0, "should marshal to YAML")
}

func TestHostPort_MarshalJSON(t *testing.T) {
	hp := HostPort{Host: "localhost", Port: "8080"}
	data, err := json.Marshal(&hp)
	testutils.Ok(t, err)
	testutils.Assert(t, len(data) > 0, "should marshal to JSON")
	testutils.Assert(t, string(data) == `"localhost:8080"`, "should marshal correctly")
}

func TestHostPort_String(t *testing.T) {
	hp := HostPort{Host: "localhost", Port: "8080"}
	testutils.Equals(t, "localhost:8080", hp.String(), "String should format correctly")
}

func TestHostPort_String_Empty(t *testing.T) {
	hp := HostPort{}
	testutils.Equals(t, "", hp.String(), "Empty HostPort should return empty string")
}

func TestHostPort_String_OnlyHost(t *testing.T) {
	hp := HostPort{Host: "localhost", Port: ""}
	testutils.Equals(t, "localhost:", hp.String(), "Should format with empty port")
}

func TestHostPort_String_OnlyPort(t *testing.T) {
	hp := HostPort{Host: "", Port: "8080"}
	testutils.Equals(t, ":8080", hp.String(), "Should format with empty host")
}
