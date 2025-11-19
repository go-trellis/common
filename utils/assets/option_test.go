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

import (
	"testing"

	"trellis.tech/trellis/common.v3/utils/testutils"
)

func TestOptions_Init_Default(t *testing.T) {
	opts := &Options{}
	opts.init()

	testutils.Equals(t, defaultSuffix, opts.Suffix, "Suffix should default to .gz")
	testutils.Equals(t, defaultPath, opts.Path, "Path should default to ./")
}

func TestOptions_Init_Custom(t *testing.T) {
	opts := &Options{
		Suffix: ".zip",
		Path:   "/tmp",
	}
	opts.init()

	testutils.Equals(t, ".zip", opts.Suffix, "Suffix should not change")
	testutils.Equals(t, "/tmp", opts.Path, "Path should not change")
}

func TestOptions_Init_EmptySuffix(t *testing.T) {
	opts := &Options{
		Suffix: "",
		Path:   "/tmp",
	}
	opts.init()

	testutils.Equals(t, defaultSuffix, opts.Suffix, "Empty Suffix should default to .gz")
	testutils.Equals(t, "/tmp", opts.Path, "Path should not change")
}

func TestOptions_Init_EmptyPath(t *testing.T) {
	opts := &Options{
		Suffix: ".zip",
		Path:   "",
	}
	opts.init()

	testutils.Equals(t, ".zip", opts.Suffix, "Suffix should not change")
	testutils.Equals(t, defaultPath, opts.Path, "Empty Path should default to ./")
}

func TestOptSuffix(t *testing.T) {
	opts := &Options{}
	OptSuffix(".zip")(opts)

	testutils.Equals(t, ".zip", opts.Suffix, "OptSuffix should set suffix")
}

func TestOptPath(t *testing.T) {
	opts := &Options{}
	OptPath("/custom/path")(opts)

	testutils.Equals(t, "/custom/path", opts.Path, "OptPath should set path")
}
