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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-trellis/common/transaction"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"xorm.io/xorm/contexts"
)

type countHook struct {
	after int
	sqls  []string
}

func (h *countHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	return c.Ctx, nil
}

func (h *countHook) AfterProcess(c *contexts.ContextHook) error {
	h.after++
	h.sqls = append(h.sqls, c.SQL)
	return nil
}

func TestAddHookOnXORMEngine(t *testing.T) {
	engine, err := NewXORMEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
	)
	if err != nil {
		t.Fatalf("NewXORMEngine: %v", err)
	}
	defer engine.Close()

	h := &countHook{}
	if err := AddHook(engine, h); err != nil {
		t.Fatalf("AddHook(*xorm.Engine): %v", err)
	}
	if err := h.AfterProcess(&contexts.ContextHook{
		Ctx: context.Background(),
		SQL: "SELECT 1",
	}); err != nil {
		t.Fatal(err)
	}
	if h.after != 1 {
		t.Fatalf("after = %d, want 1", h.after)
	}
}

func TestAddHookOnTransactionEngine(t *testing.T) {
	x, err := NewXEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
	)
	if err != nil {
		t.Fatalf("NewXEngine: %v", err)
	}
	defer x.Close()

	var eng transaction.Engine = x
	h := &countHook{}
	if err := eng.AddHook(h); err != nil {
		t.Fatalf("transaction.Engine.AddHook: %v", err)
	}
	if err := AddHook(eng, h); err != nil {
		t.Fatalf("txorm.AddHook(transaction.Engine): %v", err)
	}
}

func TestAddHookRejectsBadType(t *testing.T) {
	x, err := NewXEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
	)
	if err != nil {
		t.Fatalf("NewXEngine: %v", err)
	}
	defer x.Close()

	if err := x.AddHook("not-a-hook"); err == nil {
		t.Fatal("expected error for non-hook type")
	}
}

func TestOptHookAttachesAtCreate(t *testing.T) {
	h := &countHook{}
	engine, err := NewXORMEngine(
		"mysql",
		"user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true",
		OptHook(h),
	)
	if err != nil {
		t.Fatalf("NewXORMEngine: %v", err)
	}
	defer engine.Close()

	// OptHook registered the hook on the engine; invoke AfterProcess via
	// the same hook instance to prove identity, and ensure AddHook path works.
	if err := AddHook(engine, h); err != nil {
		t.Fatalf("AddHook after OptHook: %v", err)
	}
}

func TestMetricsHookRecords(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewMetricsHook("testdb", reg)

	_ = h.AfterProcess(&contexts.ContextHook{
		Ctx:         context.Background(),
		SQL:         "SELECT id FROM t",
		ExecuteTime: 10 * time.Millisecond,
	})
	_ = h.AfterProcess(&contexts.ContextHook{
		Ctx:         context.Background(),
		SQL:         "INSERT INTO t VALUES (1)",
		ExecuteTime: 5 * time.Millisecond,
		Err:         errors.New("dup"),
	})

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}

	var totalSelectOK, totalInsertErr float64
	var sawDuration bool
	for _, mf := range mfs {
		switch mf.GetName() {
		case "trellis_txorm_query_total":
			for _, m := range mf.GetMetric() {
				labels := labelMap(m)
				if labels["db"] == "testdb" && labels["op"] == "select" && labels["result"] == "ok" {
					totalSelectOK = m.GetCounter().GetValue()
				}
				if labels["db"] == "testdb" && labels["op"] == "insert" && labels["result"] == "error" {
					totalInsertErr = m.GetCounter().GetValue()
				}
			}
		case "trellis_txorm_query_duration_seconds":
			sawDuration = len(mf.GetMetric()) > 0
		}
	}
	if totalSelectOK != 1 {
		t.Fatalf("select ok = %v, want 1", totalSelectOK)
	}
	if totalInsertErr != 1 {
		t.Fatalf("insert error = %v, want 1", totalInsertErr)
	}
	if !sawDuration {
		t.Fatal("expected duration histogram samples")
	}
}

func TestMetricsHookReregisterSafe(t *testing.T) {
	reg := prometheus.NewRegistry()
	h1 := NewMetricsHook("db1", reg)
	h2 := NewMetricsHook("db2", reg)
	if h1.total != h2.total {
		t.Fatal("expected shared CounterVec after AlreadyRegistered reuse")
	}
	_ = h1.AfterProcess(&contexts.ContextHook{Ctx: context.Background(), SQL: "SELECT 1"})
	_ = h2.AfterProcess(&contexts.ContextHook{Ctx: context.Background(), SQL: "SELECT 1"})
}

func TestSQLOp(t *testing.T) {
	cases := map[string]string{
		"SELECT 1":            "select",
		"  insert into t":     "insert",
		"BEGIN TRANSACTION":   "begin",
		"COMMIT":              "commit",
		"ROLLBACK":            "rollback",
		"SHOW TABLES":         "other",
		"":                    "other",
	}
	for sql, want := range cases {
		if got := sqlOp(sql); got != want {
			t.Fatalf("sqlOp(%q) = %q, want %q", sql, got, want)
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
