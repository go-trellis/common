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
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-trellis/common/config"
	"github.com/go-trellis/common/logger"
	"github.com/go-trellis/common/transaction"

	"github.com/sirupsen/logrus"
	"xorm.io/xorm/log"
)

const testEnginesYAML = `
db1:
  driver: mysql
  dsn: "user:pass@tcp(127.0.0.1:1)/db1?timeout=1ms&parseTime=true"
  is_default: true
  max_idle_conns: 1
  max_open_conns: 1
db2:
  driver: mysql
  dsn: "user:pass@tcp(127.0.0.1:1)/db2?timeout=1ms&parseTime=true"
  max_idle_conns: 1
  max_open_conns: 1
`

func testEnginesConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.NewConfigOptions(config.OptionString(config.ReaderTypeYAML, testEnginesYAML))
	if err != nil {
		t.Fatalf("new config: %v", err)
	}
	return cfg
}

func closeEngines(engines map[string]transaction.Engine) {
	seen := make(map[transaction.Engine]struct{})
	for _, engine := range engines {
		if _, ok := seen[engine]; ok {
			continue
		}
		seen[engine] = struct{}{}
		_ = engine.Close()
	}
}

func TestNewEnginesWithConfig(t *testing.T) {
	cfg := testEnginesConfig(t)
	engines, err := NewEnginesWithConfig(cfg, nil)
	if err != nil {
		t.Fatalf("NewEnginesWithConfig: %v", err)
	}
	defer closeEngines(engines)

	if _, ok := engines["db1"]; !ok {
		t.Fatal("missing db1 engine")
	}
	if _, ok := engines["db2"]; !ok {
		t.Fatal("missing db2 engine")
	}
	if _, ok := engines[transaction.DefaultDatabase]; !ok {
		t.Fatal("missing default engine")
	}
}

func TestNewEnginesWithConfigConcurrent(t *testing.T) {
	cfg := testEnginesConfig(t)
	const n = 16

	var wg sync.WaitGroup
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engines, err := NewEnginesWithConfig(cfg, nil)
			if err != nil {
				errCh <- err
				return
			}
			closeEngines(engines)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out: concurrent NewEnginesWithConfig appears stuck")
	}

	close(errCh)
	for err := range errCh {
		t.Fatalf("NewEnginesWithConfig: %v", err)
	}
}

// reenterLogger triggers another NewEnginesWithConfig from SetLevel.
// With the old package-level locker this deadlocked; it must complete now.
type reenterLogger struct {
	log.DiscardLogger
	cfg  config.Config
	once sync.Once
	err  error
}

func (l *reenterLogger) SetLevel(level log.LogLevel) {
	l.once.Do(func() {
		engines, err := NewEnginesWithConfig(l.cfg, nil)
		l.err = err
		if err == nil {
			closeEngines(engines)
		}
	})
}

func TestNewEnginesWithConfigReentrantCallback(t *testing.T) {
	cfg := testEnginesConfig(t)
	l := &reenterLogger{cfg: cfg}

	done := make(chan error, 1)
	go func() {
		engines, err := NewEnginesWithConfig(cfg, l)
		if err == nil {
			closeEngines(engines)
		}
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("outer NewEnginesWithConfig: %v", err)
		}
		if l.err != nil {
			t.Fatalf("reentrant NewEnginesWithConfig: %v", l.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out: SetLevel reentered NewEnginesWithConfig and deadlocked")
	}
}

const testShowSQLYAML = `
db1:
  driver: mysql
  dsn: "user:pass@tcp(127.0.0.1:1)/db1?timeout=1ms&parseTime=true"
  show_sql: true
  log_level: 1
  max_idle_conns: 1
  max_open_conns: 1
`

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

func (l *captureXormLogger) IsShowSQL() bool { return l.showSQL }

func (l *captureXormLogger) SetLevel(level log.LogLevel) { l.level = level }

func (l *captureXormLogger) Level() log.LogLevel { return l.level }

func (l *captureXormLogger) Infof(format string, v ...any) {
	l.infos = append(l.infos, fmt.Sprintf(format, v...))
}

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
		t.Fatal("engine.Logger().IsShowSQL() = false, ShowSQL was lost after SetLogger")
	}
	if !capLog.IsShowSQL() {
		t.Fatal("injected logger IsShowSQL() = false")
	}
	if capLog.Level() != log.LOG_INFO {
		t.Fatalf("injected logger level = %v, want LOG_INFO", capLog.Level())
	}

	engine.Logger().AfterSQL(log.LogContext{
		Ctx:         context.Background(),
		SQL:         "SELECT 1",
		Args:        []any{42},
		ExecuteTime: time.Millisecond,
	})
	if len(capLog.infos) == 0 {
		t.Fatal("expected SQL Infof on injected logger")
	}
	if !strings.Contains(capLog.infos[0], "SELECT 1") {
		t.Fatalf("sql log = %q, want SELECT 1", capLog.infos[0])
	}
}

func TestConfigShowSQLAppliesToInjectedLogger(t *testing.T) {
	cfg, err := config.NewConfigOptions(config.OptionString(config.ReaderTypeYAML, testShowSQLYAML))
	if err != nil {
		t.Fatalf("new config: %v", err)
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
		t.Fatal("missing db1 engine")
	}
	if !engine.Logger().IsShowSQL() {
		t.Fatal("show_sql: true did not apply to injected logger")
	}
	if engine.Logger().Level() != log.LOG_INFO {
		t.Fatalf("log_level = %v, want LOG_INFO", engine.Logger().Level())
	}
}

func TestToXormLoggerPrintsSQLAfterShowSQL(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logrus.New()
	l.SetOutput(io.MultiWriter(buf, io.Discard))
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	xl := logger.ToXormLogger(l)
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
		t.Fatal("ToXormLogger IsShowSQL() = false after OptShowSQL")
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

func TestInitLoggerPathPrintsSQL(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg, err := config.NewConfigOptions(config.OptionString(config.ReaderTypeYAML, `
type: console
level: 4
encoding: text
std_printers: [stdout]
`))
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	l, err := logger.InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	ll, ok := l.(*logrus.Logger)
	if !ok {
		t.Fatalf("InitLogger type = %T, want *logrus.Logger", l)
	}
	ll.SetOutput(buf)
	ll.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	engine, err := NewXORMEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
		OptLogger(logger.ToXormLogger(l)),
		OptShowSQL(true),
		OptLogLevel(log.LOG_INFO),
	)
	if err != nil {
		t.Fatalf("NewXORMEngine: %v", err)
	}
	defer engine.Close()

	if !engine.Logger().IsShowSQL() {
		t.Fatal("InitLogger path: IsShowSQL() = false")
	}
	engine.Logger().AfterSQL(log.LogContext{
		Ctx:         context.Background(),
		SQL:         "SELECT 1 AS n",
		Args:        nil,
		ExecuteTime: time.Millisecond,
	})
	out := buf.String()
	if !strings.Contains(out, "[SQL]") || !strings.Contains(out, "SELECT 1 AS n") {
		t.Fatalf("InitLogger path output %q missing SQL", out)
	}
}
