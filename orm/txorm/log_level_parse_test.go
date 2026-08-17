package txorm

import (
	"testing"

	"xorm.io/xorm/log"
)

func TestParseXormLogLevel(t *testing.T) {
	cases := []struct {
		in   any
		want log.LogLevel
	}{
		{nil, log.LOG_DEBUG},
		{"", log.LOG_DEBUG},
		{"debug", log.LOG_DEBUG},
		{"info", log.LOG_INFO},
		{"WARN", log.LOG_WARNING},
		{"warning", log.LOG_WARNING},
		{"error", log.LOG_ERR},
		{"err", log.LOG_ERR},
		{"off", log.LOG_OFF},
		{3, log.LOG_ERR},
		{1, log.LOG_INFO},
		{"3", log.LOG_ERR},
		{float64(2), log.LOG_WARNING},
	}
	for _, c := range cases {
		got := parseXormLogLevel(c.in)
		if got != c.want {
			t.Fatalf("parseXormLogLevel(%v)=%v want %v", c.in, got, c.want)
		}
	}
}
