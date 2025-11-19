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
	"path/filepath"
	"strings"
)

// Join joins path elements
func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Clean cleans path
func Clean(path string) string {
	return filepath.Clean(path)
}

// Base returns last element of path
func Base(path string) string {
	return filepath.Base(path)
}

// Dir returns directory part of path
func Dir(path string) string {
	return filepath.Dir(path)
}

// Ext returns file extension
func Ext(path string) string {
	return filepath.Ext(path)
}

// IsAbs checks if path is absolute
func IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// Abs returns absolute path
func Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Rel returns relative path
func Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

// Exists checks if path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsFile checks if path is a file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDir checks if path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Split splits path into directory and file
func Split(path string) (dir, file string) {
	return filepath.Split(path)
}

// Glob finds files matching pattern
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// WalkDir walks directory tree
func WalkDir(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// EnsureDir ensures directory exists, creates if not
func EnsureDir(path string) error {
	if IsDir(path) {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

// RemoveExt removes file extension
func RemoveExt(path string) string {
	ext := Ext(path)
	if ext == "" {
		return path
	}
	return strings.TrimSuffix(path, ext)
}

// AddExt adds extension to path if not present
func AddExt(path, ext string) string {
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	currentExt := Ext(path)
	if currentExt != "" {
		return path
	}
	return path + "." + ext
}

// ChangeExt changes file extension
func ChangeExt(path, newExt string) string {
	if strings.HasPrefix(newExt, ".") {
		newExt = newExt[1:]
	}
	withoutExt := RemoveExt(path)
	if withoutExt == path {
		return path
	}
	return withoutExt + "." + newExt
}

// Normalize normalizes path separators (useful for cross-platform)
func Normalize(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

// Match checks if name matches shell pattern
func Match(pattern, name string) (bool, error) {
	return filepath.Match(pattern, name)
}

// VolumeName returns volume name (Windows only)
func VolumeName(path string) string {
	return filepath.VolumeName(path)
}
