package logger

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	maxCallerDepth        = 32
	logrusPackagePrefix   = "github.com/sirupsen/logrus"
)

var (
	loggerPackageOnce sync.Once
	loggerPackageName string
)

func thisLoggerPackage() string {
	loggerPackageOnce.Do(func() {
		pc, _, _, _ := runtime.Caller(0)
		loggerPackageName = callerPackageName(runtime.FuncForPC(pc).Name())
	})
	return loggerPackageName
}

// callerPackageName reduces a fully qualified function name to its package path
// (same approach as logrus.getPackageName).
func callerPackageName(fn string) string {
	for {
		lastPeriod := strings.LastIndex(fn, ".")
		lastSlash := strings.LastIndex(fn, "/")
		if lastPeriod > lastSlash {
			fn = fn[:lastPeriod]
			continue
		}
		break
	}
	return fn
}

func shouldSkipCallerPackage(pkg string) bool {
	switch {
	case pkg == "runtime":
		return true
	case pkg == thisLoggerPackage():
		return true
	case pkg == logrusPackagePrefix, strings.HasPrefix(pkg, logrusPackagePrefix+"/"):
		return true
	default:
		return false
	}
}

// findExternalCaller walks the stack past logrus and this wrapper package so
// ReportCaller points at the real application / xorm call site.
func findExternalCaller() *runtime.Frame {
	pcs := make([]uintptr, maxCallerDepth)
	n := runtime.Callers(0, pcs)
	if n == 0 {
		return nil
	}
	frames := runtime.CallersFrames(pcs[:n])
	for {
		f, more := frames.Next()
		pkg := callerPackageName(f.Function)
		if !shouldSkipCallerPackage(pkg) {
			return &f
		}
		if !more {
			break
		}
	}
	return nil
}

func skipWrapperCallerPrettyfier(frame *runtime.Frame) (function string, file string) {
	if real := findExternalCaller(); real != nil {
		frame = real
	}
	if frame == nil {
		return "", ""
	}
	return frame.Function, fmt.Sprintf("%s:%d", frame.File, frame.Line)
}

// ensureReportCallerPrettyfier makes Text/JSON formatters skip LogrusLogger
// wrappers when printing func/file for ReportCaller.
func ensureReportCallerPrettyfier(l *logrus.Logger) {
	if l == nil {
		return
	}
	switch f := l.Formatter.(type) {
	case *logrus.TextFormatter:
		if f.CallerPrettyfier == nil {
			f.CallerPrettyfier = skipWrapperCallerPrettyfier
		}
	case *logrus.JSONFormatter:
		if f.CallerPrettyfier == nil {
			f.CallerPrettyfier = skipWrapperCallerPrettyfier
		}
	}
}
