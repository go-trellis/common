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

package json

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestMarshal(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	obj := TestStruct{Name: "test", Value: 42}
	data, err := Marshal(obj)
	testutils.Ok(t, err)
	testutils.Assert(t, len(data) > 0, "marshal should return data")
	testutils.Assert(t, strings.Contains(string(data), "test"), "data should contain name")
	testutils.Assert(t, strings.Contains(string(data), "42"), "data should contain value")
}

func TestUnmarshal(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := []byte(`{"name":"test","value":42}`)
	var obj TestStruct
	err := Unmarshal(data, &obj)
	testutils.Ok(t, err)
	testutils.Equals(t, obj.Name, "test")
	testutils.Equals(t, obj.Value, 42)
}

func TestMarshalIndent(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	obj := TestStruct{Name: "test", Value: 42}
	data, err := MarshalIndent(obj, "", "  ")
	testutils.Ok(t, err)
	testutils.Assert(t, len(data) > 0, "marshal indent should return data")
	testutils.Assert(t, strings.Contains(string(data), "test"), "data should contain name")
}

func TestNewDecoder(t *testing.T) {
	data := `{"name":"test","value":42}`
	reader := strings.NewReader(data)
	decoder := NewDecoder(reader)
	testutils.Assert(t, decoder != nil, "decoder should not be nil")

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var obj TestStruct
	err := decoder.Decode(&obj)
	testutils.Ok(t, err)
	testutils.Equals(t, obj.Name, "test")
	testutils.Equals(t, obj.Value, 42)
}

func TestNewEncoder(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	obj := TestStruct{Name: "test", Value: 42}
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)
	testutils.Assert(t, encoder != nil, "encoder should not be nil")

	err := encoder.Encode(obj)
	testutils.Ok(t, err)
	testutils.Assert(t, buf.Len() > 0, "encoded data should not be empty")
}

func TestMarshalUnmarshalComplex(t *testing.T) {
	type Nested struct {
		Field string `json:"field"`
	}
	type TestStruct struct {
		Name    string   `json:"name"`
		Value   int      `json:"value"`
		Items   []string `json:"items"`
		Nested  Nested   `json:"nested"`
		Enabled bool     `json:"enabled"`
	}

	obj := TestStruct{
		Name:    "test",
		Value:   42,
		Items:   []string{"item1", "item2"},
		Nested:  Nested{Field: "nested_value"},
		Enabled: true,
	}

	data, err := Marshal(obj)
	testutils.Ok(t, err)

	var result TestStruct
	err = Unmarshal(data, &result)
	testutils.Ok(t, err)
	testutils.Equals(t, result.Name, obj.Name)
	testutils.Equals(t, result.Value, obj.Value)
	testutils.Equals(t, len(result.Items), len(obj.Items))
	testutils.Equals(t, result.Nested.Field, obj.Nested.Field)
	testutils.Equals(t, result.Enabled, obj.Enabled)
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	invalidData := []byte(`{invalid json}`)
	var obj TestStruct
	err := Unmarshal(invalidData, &obj)
	testutils.NotOk(t, err, "should return error for invalid JSON")
}

func TestMarshalNil(t *testing.T) {
	data, err := Marshal(nil)
	testutils.Ok(t, err)
	testutils.Equals(t, string(data), "null")
}

func TestDecoderDecodeMultiple(t *testing.T) {
	data := `{"name":"test1"}{"name":"test2"}`
	reader := strings.NewReader(data)
	decoder := NewDecoder(reader)

	type TestStruct struct {
		Name string `json:"name"`
	}

	var obj1 TestStruct
	err := decoder.Decode(&obj1)
	testutils.Ok(t, err)
	testutils.Equals(t, obj1.Name, "test1")

	var obj2 TestStruct
	err = decoder.Decode(&obj2)
	testutils.Ok(t, err)
	testutils.Equals(t, obj2.Name, "test2")
}

func TestEncoderEncodeToWriter(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	obj := TestStruct{Name: "test"}
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	err := encoder.Encode(obj)
	testutils.Ok(t, err)

	// Verify the encoded data can be decoded
	var result TestStruct
	decoder := NewDecoder(&buf)
	err = decoder.Decode(&result)
	testutils.Ok(t, err)
	testutils.Equals(t, result.Name, obj.Name)
}

func TestNumberType(t *testing.T) {
	// Test that Number type is available
	var num Number = "123"
	testutils.Assert(t, num != "", "number should not be empty")
}

func TestRawMessageType(t *testing.T) {
	// Test that RawMessage type is available
	var raw RawMessage = []byte(`{"test":"value"}`)
	testutils.Assert(t, len(raw) > 0, "raw message should not be empty")
}
