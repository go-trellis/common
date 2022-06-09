/*
Copyright © 2022 Henry Huang <hhh@rutcode.com>

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
	"fmt"
)

var _ flag.Value = (*Strings)(nil)

// Strings array string
type Strings []string

func (x Strings) Len() int { return len(x) }

func (x Strings) Less(i, j int) bool { return x[i] < x[j] }

func (x Strings) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

// String implements flag.Value
func (x Strings) String() string {
	return fmt.Sprintf("%s", []string(x))
}

// Set implements flag.Value
func (x *Strings) Set(s string) error {
	*x = append(*x, s)
	return nil
}
