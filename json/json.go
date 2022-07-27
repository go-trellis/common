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
	eJson "encoding/json"
	"io"

	iter "github.com/json-iterator/go"
)

type JsonNumber = eJson.Number
type Number = iter.Number

func Marshal(v interface{}) ([]byte, error) {
	json := iter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return iter.MarshalIndent(v, prefix, indent)
}

func MarshalToString(v interface{}) (string, error) {
	return iter.MarshalToString(v)
}

func Unmarshal(bs []byte, v interface{}) error {
	json := iter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(bs, v)
}

func UnmarshalFromString(bs string, v interface{}) error {
	return iter.UnmarshalFromString(bs, v)
}

func NewDecoder(reader io.Reader) *iter.Decoder {
	return iter.NewDecoder(reader)
}

func NewEncoder(writer io.Writer) *iter.Encoder {
	return iter.NewEncoder(writer)
}
