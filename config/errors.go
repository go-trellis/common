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

import "trellis.tech/trellis/common.v2/errcode"

// Errors
var (
	ErrNotMap                 = errcode.New("interface is not a map")
	ErrInvalidKey             = errcode.New("invalid key")
	ErrInvalidFilePath        = errcode.New("invalid file path")
	ErrUnknownSuffixes        = errcode.New("unknown file with suffix")
	ErrNotSupportedReaderType = errcode.New("not supported reader type")
)
