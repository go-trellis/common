/*
Copyright © 2016 Henry Huang <hhh@rutcode.com>

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

package flagext

import (
	"flag"
	"fmt"
)

// Parser is a thing that can ParseFlags
type Parser interface {
	ParseFlags(*flag.FlagSet)
}

// ParseFlags registers flags with the provided Parsers
func ParseFlags(rs ...Parser) {
	for _, r := range rs {
		r.ParseFlags(flag.CommandLine)
	}
}

// DefaultValues initiates a set of configs (Parsers) with their defaults.
func DefaultValues(rs ...Parser) {
	fs := flag.NewFlagSet("", flag.PanicOnError)
	for _, r := range rs {
		r.ParseFlags(fs)
	}
	_ = fs.Parse([]string{})
}

// StringSlice is a slice of strings that implements flag.Value
type StringSlice []string

// String implements flag.Value
func (v StringSlice) String() string {
	return fmt.Sprintf("%s", []string(v))
}

// Set implements flag.Value
func (v *StringSlice) Set(s string) error {
	*v = append(*v, s)
	return nil
}
