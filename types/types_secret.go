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

import "flag"

var _ flag.Value = (*Secret)(nil)

const Hidden = "<hidden>"

type Secret string

// String implements flag.Value
func (p Secret) String() string {
	return string(p)
}

// Set implements flag.Value
func (p *Secret) Set(s string) error {
	*p = Secret(s)
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface for Secret.
func (p Secret) MarshalYAML() (interface{}, error) {
	if len(p) == 0 {
		return "", nil
	}
	return Hidden, nil
}

//UnmarshalYAML implements the yaml.Unmarshaler interface for Secret.
func (p *Secret) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}
	return p.Set(v)
}
