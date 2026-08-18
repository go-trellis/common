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
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-trellis/common/config"

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

func TestToXormLoggerFromInitLogger(t *testing.T) {
	cfg, err := config.NewConfigOptions(config.OptionString(config.ReaderTypeYAML, `
type: console
level: 4
encoding: text
std_printers: [stdout]
`))
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	l, err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	xl := ToXormLogger(l).(*XormLogrus)
	if xl.Logger == nil {
		t.Fatal("expected underlying *logrus.Logger from InitLogger")
	}
	if xl.Level() != log.LOG_INFO {
		t.Fatalf("level = %v, want LOG_INFO", xl.Level())
	}
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

func TestAfterSQLIncludesTraceID(t *testing.T) {
	var buf bytes.Buffer
	l := logrus.New()
	l.SetOutput(&buf)
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	xl := ToXormLogger(l).(*XormLogrus)
	xl.SetLevel(log.LOG_INFO)

	ctx := context.WithValue(context.Background(), traceIDLogField, "trc-abc")
	xl.AfterSQL(log.LogContext{
		Ctx:         ctx,
		SQL:         "SELECT 1",
		Args:        []any{},
		ExecuteTime: time.Millisecond,
	})
	out := buf.String()
	if !strings.Contains(out, "[SQL]") {
		t.Fatalf("missing SQL: %s", out)
	}
	if !strings.Contains(out, "trace_id=trc-abc") && !strings.Contains(out, `trace_id="trc-abc"`) {
		t.Fatalf("missing trace_id: %s", out)
	}

	buf.Reset()
	xl.AfterSQL(log.LogContext{
		Ctx:         context.Background(),
		SQL:         "SELECT 2",
		Args:        []any{},
		ExecuteTime: time.Millisecond,
	})
	out = buf.String()
	if !strings.Contains(out, "[SQL]") {
		t.Fatalf("missing SQL without trace: %s", out)
	}
	if strings.Contains(out, "trace_id=") {
		t.Fatalf("unexpected trace_id: %s", out)
	}
}
