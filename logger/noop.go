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
func (noop) Debug(msg string, args ...interface{})  {}
func (noop) Debugf(msg string, args ...interface{}) {}
func (noop) Infof(msg string, args ...interface{})  {}
func (noop) Info(msg string, args ...interface{})   {}
func (noop) Warnf(msg string, args ...interface{})  {}
func (noop) Warn(msg string, args ...interface{})   {}
func (noop) Error(msg string, args ...interface{})  {}
func (noop) Errorf(msg string, args ...interface{}) {}
func (noop) Panic(msg string, args ...interface{})  {}
func (noop) Panicf(msg string, args ...interface{}) {}
func (noop) Fatal(msg string, args ...interface{})  {}
func (noop) Fatalf(msg string, args ...interface{}) {}
func (noop) With(...interface{}) Logger {
	return &noop{}
}
