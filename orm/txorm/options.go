/*
Copyright © 2024 Henry Huang <hhh@rutcode.com>

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

package txorm

// In represents an input parameter for a query. It contains the column name and the arguments to be passed to the database.
type In struct {
	Column string
	Args   []any
}

// InOpts returns a new In with the given column and arguments.
func InOpts(column string, args ...any) *In {
	if column == "" {
		return nil
	}
	return &In{column, args}
}

// Args returns a slice of any with the given arguments.
func Args(args ...any) []any {
	var bs []any
	return append(bs, args...)
}
