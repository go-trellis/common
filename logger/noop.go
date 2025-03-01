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

import (
	"io"

	"xorm.io/xorm/log"
)

// Noop logger.
func Noop() Logger {
	return &noop{}
}

type noop struct {
	level log.LogLevel
}

func (noop) Log(...any) error {
	return nil
}
func (noop) Debug(args ...any)              {}
func (noop) Debugf(msg string, args ...any) {}
func (noop) DebugM(msg string, args ...any) {}
func (noop) Info(args ...any)               {}
func (noop) Infof(msg string, args ...any)  {}
func (noop) InfoM(msg string, args ...any)  {}
func (noop) Warn(args ...any)               {}
func (noop) Warnf(msg string, args ...any)  {}
func (noop) WarnM(msg string, args ...any)  {}
func (noop) Error(args ...any)              {}
func (noop) Errorf(msg string, args ...any) {}
func (noop) ErrorM(msg string, args ...any) {}
func (noop) Panic(args ...any)              {}
func (noop) Panicf(msg string, args ...any) {}
func (noop) PanicM(msg string, args ...any) {}
func (noop) Fatal(args ...any)              {}
func (noop) Fatalf(msg string, args ...any) {}
func (noop) FatalM(msg string, args ...any) {}
func (noop) Writer() io.Writer {
	return io.Discard
}
func (noop) With(...any) Logger {
	return &noop{}
}
func (p *noop) Level() log.LogLevel {
	return p.level
}

func (p *noop) SetLevel(l log.LogLevel) {
	p.level = l
}

func (p *noop) ShowSQL(show ...bool) {}
func (p *noop) IsShowSQL() bool      { return false }
