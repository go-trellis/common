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
	"gopkg.in/yaml.v3"
)

type defYamlReader struct {
	opts ReaderOptions
}

// NewYAMLReader creates a new YAML reader with the given options.
func NewYAMLReader(opts ...ReaderOptionFunc) Reader {
	r := &defYamlReader{}
	for _, o := range opts {
		o(&r.opts)
	}
	return r
}

func (p *defYamlReader) Read(model any) error {
	return ParseYAMLFileToModel(p.opts.filename, model)
}

func (*defYamlReader) Dump(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (*defYamlReader) ParseData(data []byte, model any) error {
	return ParseYAMLData(data, model)
}

// ParseYAMLData parses YAML data into a given model.
func ParseYAMLData(data []byte, model any) error {
	return yaml.Unmarshal(data, model)
}

// ParseYAMLFileToModel parses YAML data from a file into a given model.
func ParseYAMLFileToModel(name string, model any) error {
	data, err := ReadFile(name)
	if err != nil {
		return err
	}
	return ParseYAMLData(data, model)
}
