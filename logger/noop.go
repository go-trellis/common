/*
Copyright © 2021 Henry Huang <hhh@rutcode.com>

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

package logger

// Noop logger.
func Noop() Logger {
	return noop{}
}

type noop struct{}

func (noop) Log(...interface{}) error {
	return nil
}
func (noop) Debug(args ...interface{})              {}
func (noop) Debugf(msg string, args ...interface{}) {}
func (noop) DebugM(msg string, args ...interface{}) {}
func (noop) Info(args ...interface{})               {}
func (noop) Infof(msg string, args ...interface{})  {}
func (noop) InfoM(msg string, args ...interface{})  {}
func (noop) Warn(args ...interface{})               {}
func (noop) Warnf(msg string, args ...interface{})  {}
func (noop) WarnM(msg string, args ...interface{})  {}
func (noop) Error(args ...interface{})              {}
func (noop) Errorf(msg string, args ...interface{}) {}
func (noop) ErrorM(msg string, args ...interface{}) {}
func (noop) Panic(args ...interface{})              {}
func (noop) Panicf(msg string, args ...interface{}) {}
func (noop) PanicM(msg string, args ...interface{}) {}
func (noop) Fatal(args ...interface{})              {}
func (noop) Fatalf(msg string, args ...interface{}) {}
func (noop) FatalM(msg string, args ...interface{}) {}
func (noop) With(...interface{}) Logger {
	return &noop{}
}
