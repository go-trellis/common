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
	"testing"

	"trellis.tech/trellis/common.v3/testutils"
	"xorm.io/xorm/log"
)

func TestNoop(t *testing.T) {
	n := Noop()
	testutils.Assert(t, n != nil, "Noop should return non-nil logger")
}

func TestNoop_Log(t *testing.T) {
	n := Noop().(*noop)
	err := n.Log("test")
	testutils.Ok(t, err)
}

func TestNoop_Debug(t *testing.T) {
	n := Noop().(*noop)
	// Should not panic
	n.Debug("test")
	n.Debugf("test %s", "message")
}

func TestNoop_Info(t *testing.T) {
	n := Noop().(*noop)
	// Should not panic
	n.Info("test")
	n.Infof("test %s", "message")
}

func TestNoop_Warn(t *testing.T) {
	n := Noop().(*noop)
	// Should not panic
	n.Warn("test")
	n.Warnf("test %s", "message")
}

func TestNoop_Error(t *testing.T) {
	n := Noop().(*noop)
	// Should not panic
	n.Error("test")
	n.Errorf("test %s", "message")
}

func TestNoop_Writer(t *testing.T) {
	n := Noop().(*noop)
	writer := n.Writer()
	testutils.Assert(t, writer == io.Discard, "Writer should return io.Discard")
}

func TestNoop_With(t *testing.T) {
	n := Noop().(*noop)
	newLogger := n.With("key", "value")
	testutils.Assert(t, newLogger != nil, "With should return non-nil logger")
	testutils.Assert(t, newLogger != n, "With should return a new logger")
}

func TestNoop_Level(t *testing.T) {
	n := Noop().(*noop)
	level := n.Level()
	// Default level is zero value (0)
	testutils.Equals(t, log.LogLevel(0), level, "Level should return zero value by default")
}

func TestNoop_SetLevel(t *testing.T) {
	n := Noop().(*noop)
	n.SetLevel(log.LOG_DEBUG)
	level := n.Level()
	testutils.Equals(t, log.LOG_DEBUG, level, "Level should return set level")
}

func TestNoop_ShowSQL(t *testing.T) {
	n := Noop().(*noop)
	// Should not panic
	n.ShowSQL(true)
	n.ShowSQL(false)
	n.ShowSQL()
}

func TestNoop_IsShowSQL(t *testing.T) {
	n := Noop().(*noop)
	result := n.IsShowSQL()
	testutils.Assert(t, !result, "IsShowSQL should return false")
}

