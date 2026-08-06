/*
Copyright © 2026 Henry Huang <hhh@rutcode.com>

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
	"testing"

	"xorm.io/xorm/log"

	"github.com/sirupsen/logrus"
)

func TestToXormLoggerNil(t *testing.T) {
	xl := ToXormLogger(nil)
	if xl == nil {
		t.Fatal("expected non-nil XormLogger")
	}
	if xl.Level() != log.LOG_OFF {
		t.Fatalf("level = %v, want LOG_OFF", xl.Level())
	}
	// Must not panic.
	xl.Info("noop")
	xl.Debugf("noop %d", 1)
}

func TestToXormLoggerLevelMapping(t *testing.T) {
	l := logrus.New()
	l.SetLevel(logrus.WarnLevel)
	xl := ToXormLogger(l).(*XormLogrus)
	if xl.Level() != log.LOG_WARNING {
		t.Fatalf("level = %v, want LOG_WARNING", xl.Level())
	}
	if xl.enabled(log.LOG_INFO) {
		t.Fatal("info should be filtered when xorm level is warning")
	}
	if !xl.enabled(log.LOG_WARNING) {
		t.Fatal("warning should be enabled")
	}
}

func TestXormLogrusSetLevelDoesNotTouchLogrus(t *testing.T) {
	l := logrus.New()
	l.SetLevel(logrus.InfoLevel)
	xl := ToXormLogger(l).(*XormLogrus)
	xl.SetLevel(log.LOG_OFF)
	if l.GetLevel() != logrus.InfoLevel {
		t.Fatalf("shared logrus level changed to %v", l.GetLevel())
	}
	if xl.Level() != log.LOG_OFF {
		t.Fatalf("xorm filter level = %v, want LOG_OFF", xl.Level())
	}
}
