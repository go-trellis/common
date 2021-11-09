/*
Copyright © 2021 Henry Huang <hhh@rutcode.com>

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

package codec

// String is a code for strings.
type String struct{}

func (String) String() string {
	return "string"
}

// Unmarshal implements Codec.
func (String) Unmarshal(bytes []byte) (interface{}, error) {
	return string(bytes), nil
}

// Marshal implements Codec.
func (String) Marshal(msg interface{}) ([]byte, error) {
	return []byte(msg.(string)), nil
}
