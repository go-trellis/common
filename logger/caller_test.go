package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-trellis/common/logger"
	"github.com/go-trellis/common/utils/testutils"
	"github.com/sirupsen/logrus"
)

func TestReportCallerSkipsWrapper(t *testing.T) {
	var buf bytes.Buffer
	l := logrus.New()
	l.SetOutput(&buf)
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	ll := logger.NewWithLogrusLogger(l).(*logger.LogrusLogger)
	ll.SetReportCaller(true)
	ll.Infof("caller check")

	out := buf.String()
	testutils.Assert(t, strings.Contains(out, "caller check"), "log message missing: %s", out)
	testutils.Assert(t, !strings.Contains(out, "logger_logrus.go"), "should not report wrapper file: %s", out)
	testutils.Assert(t, strings.Contains(out, "caller_test.go"), "should report test call site: %s", out)
}
