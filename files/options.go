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

package files

type Option func(*Options)
type Options struct {
	ReadBufferLength int64
}

func ReadBufferLength(len int64) Option {
	return func(o *Options) {
		o.ReadBufferLength = len
	}
}

// DefaultReadBufferLength default reader buffer length
const (
	DefaultReadBufferLength = 1024
)
