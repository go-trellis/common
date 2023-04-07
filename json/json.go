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
	"io"

	"github.com/goccy/go-json"
)

type Number = json.Number
type RawMessage = json.RawMessage
type Delim = json.Delim
type Token = json.Token

func Unmarshal(bs []byte, v interface{}) error {
	return json.Unmarshal(bs, v)
}

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func NewDecoder(reader io.Reader) *json.Decoder {
	return json.NewDecoder(reader)
}

func NewEncoder(writer io.Writer) *json.Encoder {
	return json.NewEncoder(writer)
}
