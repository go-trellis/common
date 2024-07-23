/*
Copyright © 2017 Henry Huang <hhh@rutcode.com>

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

package config

import (
	"os"
	"reflect"
	"strings"

	"trellis.tech/trellis/common.v2/errcode"
	"trellis.tech/trellis/common.v2/types"
)

func (p *AdapterConfig) copyDollarSymbol(key string, maps *map[string]interface{}) error {
	var tokens []string
	if key != "" {
		tokens = append(tokens, key)
	}
	for mapK, mapV := range *maps {
		value := p.checkValue(mapV)
		if value != nil {
			(*maps)[mapK] = value
		}
	}
	return nil
}

func (p *AdapterConfig) checkValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32:
		return value
	case reflect.Bool:
		return value
	case reflect.String:
		s, ok := value.(string)
		if !ok {
			return value
		}
		_, matched := types.FindStringSubmatchMap(s, includeReg)
		if !matched {
			return value
		}

		pKey := s[2 : len(s)-1]

		if p.EnvAllowed && (p.EnvPrefix == "" || strings.HasPrefix(pKey, p.EnvPrefix)) {
			if env := os.Getenv(pKey); env != "" {
				return env
			}
		}
		v, err := p.getKeyValue(pKey)
		if err != nil {
			return nil
		}

		return v
	case reflect.Slice:
		vs := value.([]interface{})
		for i, v := range vs {
			newV := p.checkValue(v)
			vs[i] = newV
		}
		return vs
	case reflect.Map:
		vs, ok := value.(map[string]interface{})
		if !ok {
			return value
		}
		for k, v := range vs {
			newV := p.checkValue(v)
			vs[k] = newV
		}
		return vs
	default:
		panic(errcode.Newf("not suppered type: %+v", t))
	}
}

func (p *AdapterConfig) getKeyValue(key string) (interface{}, error) {
	tokens := strings.Split(key, ".")
	vm := p.configs[tokens[0]]

	for i, t := range tokens {
		if i == 0 {
			continue
		}

		switch v := vm.(type) {
		case Options:
			vm = v[t]
		case map[string]interface{}:
			vm = v[t]
		case map[interface{}]interface{}:
			vm = v[t]
		default:
			return nil, ErrNotMap
		}
	}
	return vm, nil
}

// setKeyValue set key value into *configs
func (p *AdapterConfig) setKeyValue(key string, value interface{}) (err error) {
	tokens := strings.Split(key, ".")
	for i := len(tokens) - 1; i >= 0; i-- {
		if i == 0 {
			p.configs[tokens[0]] = value
			return
		}
		v, _ := p.getKeyValue(strings.Join(tokens[:i], "."))
		switch vm := v.(type) {
		case Options:
			vm[tokens[i]] = value
			value = vm
		case map[string]interface{}:
			vm[tokens[i]] = value
			value = vm
		case map[interface{}]interface{}:
			vm[tokens[i]] = value
			value = vm
		default:
			value = map[string]interface{}{tokens[i]: value}
		}
	}
	return
}
