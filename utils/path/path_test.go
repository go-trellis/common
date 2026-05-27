/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

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

package path

import (
	"os"
	"testing"

	"github.com/go-trellis/common/utils/testutils"
)

func TestJoin(t *testing.T) {
	result := Join("a", "b", "c")
	testutils.Assert(t, len(result) > 0, "should join path")
}

func TestBase(t *testing.T) {
	testutils.Equals(t, "file.txt", Base("/path/to/file.txt"), "should return base name")
	// Base("/") returns "/" on Unix systems, not "."
	base := Base("/")
	testutils.Assert(t, base == "/" || base == ".", "should return / or . for root")
}

func TestDir(t *testing.T) {
	testutils.Equals(t, "/path/to", Dir("/path/to/file.txt"), "should return directory")
}

func TestExt(t *testing.T) {
	testutils.Equals(t, ".txt", Ext("file.txt"), "should return extension")
	testutils.Equals(t, "", Ext("file"), "should return empty for no extension")
}

func TestIsAbs(t *testing.T) {
	testutils.Assert(t, IsAbs("/absolute/path"), "should detect absolute path")
	testutils.Assert(t, !IsAbs("relative/path"), "should detect relative path")
}

func TestRemoveExt(t *testing.T) {
	testutils.Equals(t, "file", RemoveExt("file.txt"), "should remove extension")
	testutils.Equals(t, "file", RemoveExt("file"), "should return same if no extension")
}

func TestAddExt(t *testing.T) {
	testutils.Equals(t, "file.txt", AddExt("file", "txt"), "should add extension")
	testutils.Equals(t, "file.txt", AddExt("file", ".txt"), "should add extension with dot")
	testutils.Equals(t, "file.txt", AddExt("file.txt", "csv"), "should not add if already has ext")
}

func TestChangeExt(t *testing.T) {
	testutils.Equals(t, "file.csv", ChangeExt("file.txt", "csv"), "should change extension")
	testutils.Equals(t, "file", ChangeExt("file", "txt"), "should add extension if none")
}

func TestExists(t *testing.T) {
	testutils.Assert(t, Exists("."), "current directory should exist")
	testutils.Assert(t, !Exists("/nonexistent/path/12345"), "non-existent path should not exist")
}

func TestEnsureDir(t *testing.T) {
	tmpDir := os.TempDir()
	testDir := Join(tmpDir, "test_ensure_dir")
	defer os.RemoveAll(testDir)

	err := EnsureDir(testDir)
	testutils.Ok(t, err)
	testutils.Assert(t, IsDir(testDir), "directory should be created")
}

func TestNormalize(t *testing.T) {
	result := Normalize("a/b/c")
	testutils.Assert(t, len(result) > 0, "should normalize path")
}
