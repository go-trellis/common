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

import "trellis.tech/trellis/common.v0/json"

type JSON struct {
	id      string
	factory func() interface{}
}

func NewJSONCodec(id string, factory func() interface{}) *JSON {
	return &JSON{id: id, factory: factory}
}

func (*JSON) String() string {
	return "json"
}

// Unmarshal implements Codec.
func (p *JSON) Unmarshal(msg []byte) (interface{}, error) {
	out := p.factory()
	if err := json.Unmarshal(msg, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Marshal implements Codec.
func (p *JSON) Marshal(msg interface{}) ([]byte, error) {
	return json.Marshal(msg)
}
