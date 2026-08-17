/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

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
	"strings"
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"

	"xorm.io/xorm/contexts"
)

var (
	/** queryDuration is a histogram that records the duration of each XORM query. It has buckets for 0.1s, 0.2s, ..., 5s. The name of this metric is "xorm_query_duration_seconds". The help text explains what this metric measures. The labels "object" and "operation" are used to differentiate between different types of queries (e.g., select, insert, update). This metric is useful for monitoring the performance of XORM queries in a system. It can be used with Prometheus to visualize query duration over time. The init function registers this metric with Prometheus so that it can be exposed and monitored. **/
	queryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "xorm_query_duration_seconds",
			Help:    "Time taken to execute XORM queries",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"object", "operation"},
	)
	/** queryCount is a counter that records the number of each XORM query. It has buckets for 0.1s, 0.2s, ..., 5s. The name of this metric is "xorm_query_count". The help text explains what this metric measures. The labels "object" and "operation" are used to differentiate between different types of queries (e.g., select, insert, update). This metric is useful for monitoring the number of XORM queries executed in a system. It can be used with Prometheus to visualize query count over time. The init function registers this metric with Prometheus so that it can be exposed and monitored. **/
	queryCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xorm_query_count",
			Help: "Number of XORM queries executed",
		},
		[]string{"object", "operation"},
	)
)

func init() {
	prometheus.MustRegister(queryDuration, queryCount)
}

// instrumentQuery is a helper function that wraps the execution of an XORM query with Prometheus metrics.
func instrumentQuery(object, operation string, queryFunc func()) {
	start := time.Now()
	queryCount.WithLabelValues(object, operation).Inc()
	queryFunc()
	queryDuration.WithLabelValues(object, operation).Observe(time.Since(start).Seconds())
}

var _ contexts.Hook = (*MetricsHook)(nil)

// MetricsHook records per-SQL operation metrics via Prometheus.
type MetricsHook struct {
	db       string
	total    *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewMetricsHook creates a contexts.Hook that reports query count and latency.
// db is used as the "db" label. Optional registerer defaults to DefaultRegisterer.
// Safe to call multiple times against the same registerer (AlreadyRegistered reuse).
func NewMetricsHook(db string, registerer ...prometheus.Registerer) *MetricsHook {
	reg := prometheus.DefaultRegisterer
	if len(registerer) > 0 && registerer[0] != nil {
		reg = registerer[0]
	}
	if db == "" {
		db = "default"
	}

	total := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trellis",
			Name:      "txorm_query_total",
			Help:      "Total number of txorm SQL operations.",
		},
		[]string{"db", "op", "result"},
	)
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "trellis",
			Name:      "txorm_query_duration_seconds",
			Help:      "Latency of txorm SQL operations in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"db", "op"},
	)

	return &MetricsHook{
		db:       db,
		total:    mustRegisterOrGet(reg, total),
		duration: mustRegisterOrGet(reg, duration),
	}
}

func mustRegisterOrGet[T prometheus.Collector](reg prometheus.Registerer, c T) T {
	if err := reg.Register(c); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := are.ExistingCollector.(T); ok {
				return existing
			}
		}
	}
	return c
}

func (h *MetricsHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	return c.Ctx, nil
}

func (h *MetricsHook) AfterProcess(c *contexts.ContextHook) error {
	op := sqlOp(c.SQL)
	result := "ok"
	if c.Err != nil {
		result = "error"
	}
	h.total.WithLabelValues(h.db, op, result).Inc()
	h.duration.WithLabelValues(h.db, op).Observe(c.ExecuteTime.Seconds())
	return nil
}

func sqlOp(sql string) string {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "other"
	}
	end := 0
	for end < len(sql) && !unicode.IsSpace(rune(sql[end])) {
		end++
	}
	op := strings.ToLower(sql[:end])
	switch op {
	case "select", "insert", "update", "delete",
		"begin", "commit", "rollback", "prepare":
		return op
	default:
		if strings.HasPrefix(op, "begin") {
			return "begin"
		}
		return "other"
	}
}
