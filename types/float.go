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
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type Found float64

// String implements flag.Value
func (p Found) String() string {
	return fmt.Sprintf("%0.2f", p)
}

// Set implements flag.Value
func (p *Found) Set(f float64) error {
	*p = Found(f)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *Found) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var f float64
	if err := unmarshal(&f); err != nil {
		return err
	}
	return p.Set(f)
}

// MarshalYAML implements yaml.Marshaler.
func (p *Found) MarshalYAML() (interface{}, error) {
	if p == nil {
		return 0, nil
	}
	return *p, nil
}

// ToFloat64 covert any type to float64
func ToFloat64(value interface{}) (float64, error) {

	if value == nil {
		return 0, nil
	}

	var val string
	switch reflect.TypeOf(value).Kind() {
	case reflect.Float64:
		f := value.(float64)
		return f, nil
	case reflect.Float32:
		f := value.(float32)
		return float64(f), nil
	case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
		val = fmt.Sprintf("%d", value)
	case reflect.String:
		switch reflect.TypeOf(value).String() {
		case "json.Number":
			return value.(json.Number).Float64()
		default:
			val = value.(string)
		}
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}

	return strconv.ParseFloat(val, 64)
}

// RoundFund round fund to int64
func RoundFund(fund float64) int64 {
	fInt, fFloat := math.Modf(fund)
	f := int64(fInt)
	if fFloat >= 0.50000000000 {
		f++
	}
	return f
}
