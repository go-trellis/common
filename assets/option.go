/*
Copyright © 2023 Henry Huang <hhh@rutcode.com>

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

package assets

const (
	defaultSuffix = ".gz"
	defaultPath   = "./"
)

type Options struct {
	Suffix string
	Path   string
}

func (p *Options) init() {
	if p.Suffix == "" {
		p.Suffix = defaultSuffix
	}
	if p.Path == "" {
		p.Path = defaultPath
	}
	return
}

type Option func(*Options)

func OptSuffix(s string) Option {
	return func(o *Options) {
		o.Suffix = s
	}
}

func OptPath(path string) Option {
	return func(o *Options) {
		o.Path = path
	}
}
