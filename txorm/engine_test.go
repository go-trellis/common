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
	"sync"
	"testing"
	"time"

	"github.com/go-trellis/common/config"
	"github.com/go-trellis/common/transaction"

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
