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

package txorm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/go-trellis/common/config"
	"github.com/go-trellis/common/logger"
	"github.com/go-trellis/common/orm/transaction"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm/contexts"
	"xorm.io/xorm/log"
)

type captureXormLogger struct {
	log.DiscardLogger
	showSQL bool
	level   log.LogLevel
	infos   []string
}

func (l *captureXormLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		l.showSQL = show[0]
		return
	}
	l.showSQL = true
}
func (l *captureXormLogger) IsShowSQL() bool          { return l.showSQL }
func (l *captureXormLogger) SetLevel(lv log.LogLevel) { l.level = lv }
func (l *captureXormLogger) Level() log.LogLevel      { return l.level }
func (l *captureXormLogger) Infof(format string, v ...any) {
	l.infos = append(l.infos, fmt.Sprintf(format, v...))
}
func (l *captureXormLogger) With(...any) logger.Logger { return l }
func (l *captureXormLogger) Writer() io.Writer         { return io.Discard }

func TestOptShowSQLAppliesToInjectedLogger(t *testing.T) {
	capLog := &captureXormLogger{}
	engine, err := NewXORMEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
		OptLogger(capLog),
		OptShowSQL(true),
		OptLogLevel(log.LOG_INFO),
	)
	if err != nil {
		t.Fatalf("NewXORMEngine: %v", err)
	}
	defer engine.Close()

	if !engine.Logger().IsShowSQL() {
		t.Fatal("IsShowSQL() = false after OptShowSQL; ShowSQL was lost after SetLogger")
	}
	if capLog.Level() != log.LOG_INFO {
		t.Fatalf("level = %v, want LOG_INFO", capLog.Level())
	}
	engine.Logger().AfterSQL(log.LogContext{
		Ctx:         context.Background(),
		SQL:         "SELECT 1",
		Args:        []any{42},
		ExecuteTime: time.Millisecond,
	})
	if len(capLog.infos) == 0 || !strings.Contains(capLog.infos[0], "SELECT 1") {
		t.Fatalf("expected SQL Infof, got %v", capLog.infos)
	}
}

func TestConfigShowSQLAppliesToInjectedLogger(t *testing.T) {
	cfg, err := config.NewConfigOptions(config.OptionString(config.ReaderTypeYAML, `
db1:
  driver: mysql
  dsn: "user:pass@tcp(127.0.0.1:1)/db1?timeout=1ms&parseTime=true"
  show_sql: true
  log_level: info
  max_idle_conns: 1
  max_open_conns: 1
`))
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	capLog := &captureXormLogger{}
	engines, err := NewXORMEngineWithConfig(cfg, capLog)
	if err != nil {
		t.Fatalf("NewXORMEngineWithConfig: %v", err)
	}
	defer func() {
		for _, e := range engines {
			_ = e.Close()
		}
	}()
	engine := engines["db1"]
	if engine == nil {
		t.Fatal("missing db1")
	}
	if !engine.Logger().IsShowSQL() {
		t.Fatal("show_sql: true did not apply to injected logger")
	}
}

func TestLogrusLoggerShowSQLPath(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logrus.New()
	l.SetOutput(buf)
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	xl := logger.NewWithLogrusLogger(l)
	engine, err := NewXORMEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
		OptLogger(xl),
		OptShowSQL(true),
		OptLogLevel(log.LOG_INFO),
	)
	if err != nil {
		t.Fatalf("NewXORMEngine: %v", err)
	}
	defer engine.Close()

	if !engine.Logger().IsShowSQL() {
		t.Fatal("LogrusLogger IsShowSQL() = false")
	}
	if l.GetLevel() != logrus.InfoLevel {
		t.Fatalf("shared logrus level changed to %v", l.GetLevel())
	}

	engine.Logger().AfterSQL(log.LogContext{
		Ctx:         context.Background(),
		SQL:         "SELECT id FROM t WHERE n = ?",
		Args:        []any{7},
		ExecuteTime: 2 * time.Millisecond,
	})
	out := buf.String()
	if !strings.Contains(out, "[SQL]") || !strings.Contains(out, "SELECT id FROM t WHERE n = ?") {
		t.Fatalf("logrus output %q missing SQL", out)
	}
}

type countHook struct {
	after int
}

func (h *countHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	return c.Ctx, nil
}
func (h *countHook) AfterProcess(c *contexts.ContextHook) error {
	h.after++
	return nil
}

func TestAddHookOnXORMAndTransactionEngine(t *testing.T) {
	x, err := NewXEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
	)
	if err != nil {
		t.Fatalf("NewXEngine: %v", err)
	}
	defer x.Close()

	h := &countHook{}
	if err := AddHook(x.Engine, h); err != nil {
		t.Fatalf("AddHook(*xorm.Engine): %v", err)
	}
	var eng transaction.Engine = x
	if err := eng.AddHook(h); err != nil {
		t.Fatalf("transaction.Engine.AddHook: %v", err)
	}
	if err := x.AddHook("bad"); err == nil {
		t.Fatal("expected error for non-hook")
	}
}

func TestOptHookAndMetricsHook(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewMetricsHook("testdb", reg)
	engine, err := NewXORMEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
		OptHook(h),
	)
	if err != nil {
		t.Fatalf("NewXORMEngine: %v", err)
	}
	defer engine.Close()

	_ = h.AfterProcess(&contexts.ContextHook{
		Ctx:         context.Background(),
		SQL:         "SELECT 1",
		ExecuteTime: time.Millisecond,
	})
	_ = h.AfterProcess(&contexts.ContextHook{
		Ctx:         context.Background(),
		SQL:         "INSERT INTO t VALUES (1)",
		ExecuteTime: time.Millisecond,
		Err:         errors.New("dup"),
	})

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	var selectOK, insertErr float64
	for _, mf := range mfs {
		if mf.GetName() != "trellis_txorm_query_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			labels := labelMap(m)
			if labels["op"] == "select" && labels["result"] == "ok" {
				selectOK = m.GetCounter().GetValue()
			}
			if labels["op"] == "insert" && labels["result"] == "error" {
				insertErr = m.GetCounter().GetValue()
			}
		}
	}
	if selectOK != 1 || insertErr != 1 {
		t.Fatalf("selectOK=%v insertErr=%v", selectOK, insertErr)
	}
}

func TestSQLOp(t *testing.T) {
	cases := map[string]string{
		"SELECT 1":          "select",
		"BEGIN TRANSACTION": "begin",
		"SHOW TABLES":       "other",
		"":                  "other",
	}
	for sql, want := range cases {
		if got := sqlOp(sql); got != want {
			t.Fatalf("sqlOp(%q)=%q want %q", sql, got, want)
		}
	}
}

func labelMap(m *dto.Metric) map[string]string {
	out := make(map[string]string, len(m.GetLabel()))
	for _, l := range m.GetLabel() {
		out[l.GetName()] = l.GetValue()
	}
	return out
}
