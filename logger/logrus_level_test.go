package logger_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-trellis/common/logger"
	"github.com/go-trellis/common/utils/testutils"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm/log"
)

func TestLogLevelDoesNotGateAppInfof(t *testing.T) {
	var buf bytes.Buffer
	l := logrus.New()
	l.SetOutput(&buf)
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	xl := logger.NewWithLogrusLogger(l).(*logger.LogrusLogger)
	xl.SetLevel(log.LOG_ERR) // databases.log_level: error — must not silence app Infof

	xl.Infof("app message")
	out := buf.String()
	testutils.Assert(t, strings.Contains(out, "app message"), "app Infof gated by xorm log_level: %s", out)

	buf.Reset()
	xl.AfterSQL(log.LogContext{
		Ctx:         context.Background(),
		SQL:         "SELECT 1",
		Args:        []any{},
		ExecuteTime: time.Millisecond,
	})
	testutils.Assert(t, !strings.Contains(buf.String(), "[SQL]"), "SQL should be filtered by log_level=ERR: %s", buf.String())

	buf.Reset()
	xl.SetLevel(log.LOG_INFO)
	xl.AfterSQL(log.LogContext{
		Ctx:         context.Background(),
		SQL:         "SELECT 1",
		Args:        []any{},
		ExecuteTime: time.Millisecond,
	})
	testutils.Assert(t, strings.Contains(buf.String(), "[SQL]"), "SQL should print at log_level=INFO: %s", buf.String())
}

func TestWrapXormFilterGatesEngineInfof(t *testing.T) {
	var buf bytes.Buffer
	l := logrus.New()
	l.SetOutput(&buf)
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	xl := logger.NewWithLogrusLogger(l).(*logger.LogrusLogger)
	filtered := logger.WrapXormFilter(xl)
	filtered.SetLevel(log.LOG_ERR)

	filtered.Infof("PING DATABASE mysql")
	testutils.Assert(t, buf.Len() == 0, "engine Infof should be filtered at error: %s", buf.String())

	xl.Infof("app still ok")
	testutils.Assert(t, strings.Contains(buf.String(), "app still ok"), "app Infof should still print: %s", buf.String())

	buf.Reset()
	filtered.SetLevel(log.LOG_INFO)
	filtered.Infof("PING DATABASE mysql")
	testutils.Assert(t, strings.Contains(buf.String(), "PING DATABASE"), "engine Infof at info: %s", buf.String())
}
