/*
Copyright © 2018 Henry Huang <hhh@rutcode.com>

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

package builder

import (
	"strings"
	"testing"

	"trellis.tech/trellis/common.v3/utils/testutils"
)

func TestVersion(t *testing.T) {
	ProgramName = "test"
	ProgramVersion = "1.0.0"
	ProgramBranch = "main"
	ProgramRevision = "abc123"

	result := Version()
	testutils.Assert(t, len(result) > 0, "Version should return non-empty string")
	testutils.Assert(t, strings.Contains(result, "test"), "Version should contain program name")
	testutils.Assert(t, strings.Contains(result, "1.0.0"), "Version should contain version")
}

func TestBuildInfo(t *testing.T) {
	CompilerVersion = "go1.20"
	Author = "testuser"
	BuildTime = "2024-01-01"

	result := BuildInfo()
	testutils.Assert(t, len(result) > 0, "BuildInfo should return non-empty string")
	testutils.Assert(t, strings.Contains(result, "go1.20"), "BuildInfo should contain compiler version")
}

func TestShow_NoOptions(t *testing.T) {
	ProgramName = "test"
	ProgramVersion = "1.0.0"
	ProgramBranch = "main"
	ProgramRevision = "abc123"
	CompilerVersion = "go1.20"
	BuildTime = "2024-01-01"
	Author = "testuser"

	// Should not panic
	Show()
}

func TestShow_WithOptions(t *testing.T) {
	ProgramName = "test"
	ProgramVersion = "1.0.0"
	ProgramBranch = "main"
	ProgramRevision = "abc123"
	CompilerVersion = "go1.20"
	BuildTime = "2024-01-01"
	Author = "testuser"

	// Should not panic
	Show(OnShow(), OnColor(), Color("{{ .AnsiColor.Green }}"))
}

func TestColor(t *testing.T) {
	opt := Color("red")
	testutils.Assert(t, opt != nil, "Color should return option function")

	opts := &Options{}
	opt(opts)
	testutils.Equals(t, "red", opts.Color, "Color should set color option")
}

func TestOnShow(t *testing.T) {
	opt := OnShow()
	testutils.Assert(t, opt != nil, "OnShow should return option function")

	opts := &Options{}
	opt(opts)
	testutils.Assert(t, opts.OnShow, "OnShow should set OnShow option")
}

func TestOnColor(t *testing.T) {
	opt := OnColor()
	testutils.Assert(t, opt != nil, "OnColor should return option function")

	opts := &Options{}
	opt(opts)
	testutils.Assert(t, opts.OnColor, "OnColor should set OnColor option")
}
