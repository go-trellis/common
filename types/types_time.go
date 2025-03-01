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
	"time"
)

var _ flag.Value = (*Time)(nil)

// Time usable as flag or in YAML config.
type Time time.Time

// String implements flag.Value
func (p *Time) String() string {
	if p == nil {
		return "0"
	}
	if time.Time(*p).IsZero() {
		return "0"
	}
	return time.Time(*p).Format(time.RFC3339)
}

// Set implements flag.Value
func (p *Time) Set(s string) error {
	t, err := Parse(s)
	if err != nil {
		return err
	}
	*p = Time(*t)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *Time) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return p.Set(s)
}

// MarshalYAML implements yaml.Marshaler.
func (p Time) MarshalYAML() (any, error) {
	return p.String(), nil
}
