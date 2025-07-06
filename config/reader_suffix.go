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
	"strings"

	"trellis.tech/trellis/common.v2/files"
)

type defSuffixReader struct {
	opts   ReaderOptions
	reader Reader
}

// NewSuffixReader return a suffix reader
// supported: .json, .xml, .yaml, .yml
func NewSuffixReader(opts ...ReaderOptionFunc) (reader Reader, err error) {
	r := &defSuffixReader{}

	for _, o := range opts {
		o(&r.opts)
	}

	if r.opts.filename == "" {
		return nil, ErrInvalidFilePath
	}

	r.reader, err = fileToReader(r.opts.filename)
	if err != nil {
		return
	}
	return r, nil
}

func (p *defSuffixReader) Read(model any) (err error) {
	return p.reader.Read(model)
}

func (p *defSuffixReader) Dump(v any) ([]byte, error) {
	return p.reader.Dump(v)
}

func (p *defSuffixReader) ParseData(data []byte, model any) error {
	return p.reader.ParseData(data, model)
}

func fileToReader(filename string) (Reader, error) {
	switch {
	case strings.HasSuffix(filename, ".json"):
		return NewJSONReader(ReaderOptionFilename(filename)), nil
	case strings.HasSuffix(filename, ".xml"):
		return NewXMLReader(ReaderOptionFilename(filename)), nil
	case strings.HasSuffix(filename, ".yml"),
		strings.HasSuffix(filename, ".yaml"):
		return NewYAMLReader(ReaderOptionFilename(filename)), nil
	default:
		return nil, ErrUnknownSuffixes
	}
}

func fileToReaderType(name string) ReaderType {
	switch {
	case strings.HasSuffix(name, ".json"):
		return ReaderTypeJSON
	case strings.HasSuffix(name, ".xml"):
		return ReaderTypeXML
	case strings.HasSuffix(name, ".yml"),
		strings.HasSuffix(name, ".yaml"):
		return ReaderTypeYAML
	default:
		return ReaderTypeSuffix
	}
}

func ReadFile(name string) ([]byte, error) {
	data, _, err := files.Read(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}
