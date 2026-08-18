/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

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
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/go-trellis/common/config"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm/log"
)

// traceIDLogField is both the context value key and the logrus field name.
// Keep in sync with middleware/tracing.TraceIDKey ("trace_id").
const traceIDLogField = "trace_id"

var (
	_ Logger            = (*LogrusLogger)(nil)
	_ log.Logger        = (*LogrusLogger)(nil)
	_ log.ContextLogger = (*LogrusLogger)(nil)
)

func NewWithLogrusLogger(l *logrus.Logger) Logger {
	if l == nil {
		return &noop{}
	}
	return &LogrusLogger{logger: l, level: logrusLevelToXorm(l.GetLevel())}
}

type LogrusLogger struct {
	logger    *logrus.Logger
	isShowSQL bool
	// level filters xorm SQL/log output only; never mutate the shared logrus level.
	// orm log_level: 4 (LOG_OFF) used to map to logrus Panic and silence the process.
	level log.LogLevel
}

func NewLogrusLogger() (*LogrusLogger, error) {
	// Create a null logger if no output is configured
	nullLogger := logrus.New()
	nullLogger.SetOutput(io.Discard)

	ll := &LogrusLogger{
		logger: nullLogger,
		// Default permissive for the xorm filter; databases.*.log_level overrides via SetLevel.
		// logger config "level" only sets the logrus sink, not this field.
		level: log.LOG_DEBUG,
	}

	return ll, nil
}

// NewLogrusLoggerWithRotate creates a new logrus logger with file rotation
func NewLogrusLoggerWithRotate(config *RotateLogsConfig) (*LogrusLogger, error) {
	logger := logrus.New()

	if config != nil {
		if err := SetupRotateLogsLogger(logger, config); err != nil {
			return nil, err
		}
	} else {
		logger.SetOutput(io.Discard)
	}

	ll := &LogrusLogger{
		logger: logger,
		// Default permissive for the xorm filter; databases.*.log_level overrides via SetLevel.
		// logger config "level" only sets the logrus sink, not this field.
		level: log.LOG_DEBUG,
	}

	return ll, nil
}

// NewLogrusLoggerWithConfig builds a LogrusLogger from config.Config.
// Uses the same rotate keys as RotateLogsConfigFromConfig, plus optional:
//
//	std_printers: [stdout|stderr] — also write to these (MultiWriter with the file)
//	level: logrus level name (debug|info|warn|error|...) — NOT xorm databases.*.log_level
//	formatter: json|text — logrus JSONFormatter or TextFormatter (default text)
//	report_caller: bool — enable logrus ReportCaller (file:line of the caller)
func NewLogrusLoggerWithConfig(cfg config.Config) (*LogrusLogger, error) {
	if cfg == nil {
		return NewLogrusLogger()
	}

	rc, err := RotateLogsConfigFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	ll, err := NewLogrusLoggerWithRotate(rc)
	if err != nil {
		return nil, err
	}

	if err := applyLogrusLoggerExtras(ll, cfg); err != nil {
		return nil, err
	}
	return ll, nil
}

func logrusLevelToXorm(lv logrus.Level) log.LogLevel {
	switch lv {
	case logrus.TraceLevel, logrus.DebugLevel:
		return log.LOG_DEBUG
	case logrus.InfoLevel:
		return log.LOG_INFO
	case logrus.WarnLevel:
		return log.LOG_WARNING
	case logrus.ErrorLevel:
		return log.LOG_ERR
	case logrus.FatalLevel, logrus.PanicLevel:
		return log.LOG_OFF
	default:
		return log.LOG_UNKNOWN
	}
}

func (p *LogrusLogger) enabled(min log.LogLevel) bool {
	return p.level <= min
}

// BeforeSQL implements log.ContextLogger (no-op).
func (p *LogrusLogger) BeforeSQL(log.LogContext) {}

// AfterSQL implements log.ContextLogger. databases.*.log_level filters SQL here only;
// it must not gate general Infof/Debugf used by application code on the same logger.
// When session.Context(reqCtx) was set, trace_id is copied from ctx (do not SetLogger per request).
func (p *LogrusLogger) AfterSQL(ctx log.LogContext) {
	if !p.enabled(log.LOG_INFO) {
		return
	}
	var sessionPart string
	if ctx.Ctx != nil {
		if v := ctx.Ctx.Value(log.SessionIDKey); v != nil {
			if key, ok := v.(string); ok {
				sessionPart = fmt.Sprintf(" [%s]", key)
			}
		}
	}
	entry := p.sqlLogEntry(ctx)
	if ctx.ExecuteTime > 0 {
		entry.Infof("[SQL]%s %s %v - %v", sessionPart, ctx.SQL, ctx.Args, ctx.ExecuteTime)
		return
	}
	entry.Infof("[SQL]%s %s %v", sessionPart, ctx.SQL, ctx.Args)
}

func (p *LogrusLogger) sqlLogEntry(ctx log.LogContext) *logrus.Entry {
	if id := traceIDFromContext(ctx.Ctx); id != "" {
		return p.logger.WithField(traceIDLogField, id)
	}
	return logrus.NewEntry(p.logger)
}

func traceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(traceIDLogField).(string)
	return id
}

func applyLogrusLoggerExtras(ll *LogrusLogger, cfg config.Config) error {
	if ll == nil || ll.logger == nil || cfg == nil {
		return nil
	}

	if printers := cfg.GetStringList("std_printers"); len(printers) > 0 {
		ws := []io.Writer{}
		if ll.logger.Out != nil && ll.logger.Out != io.Discard {
			ws = append(ws, ll.logger.Out)
		}
		for _, p := range printers {
			switch strings.TrimSpace(strings.ToLower(p)) {
			case "stdout":
				ws = append(ws, os.Stdout)
			case "stderr":
				ws = append(ws, os.Stderr)
			}
		}
		switch len(ws) {
		case 0:
			// keep current output
		case 1:
			ll.logger.SetOutput(ws[0])
		default:
			ll.logger.SetOutput(io.MultiWriter(ws...))
		}
	}

	if lvlName := strings.TrimSpace(cfg.GetString("level")); lvlName != "" {
		parsed, err := logrus.ParseLevel(lvlName)
		if err != nil {
			return err
		}
		// Only the logrus sink. xorm SQL filter is databases.*.log_level via SetLevel.
		ll.logger.SetLevel(parsed)
	}

	if name := strings.TrimSpace(cfg.GetString("formatter")); name != "" {
		if err := applyLogrusFormatter(ll.logger, name); err != nil {
			return err
		}
	}

	if cfg.GetInterface("report_caller") != nil {
		ll.SetReportCaller(cfg.GetBoolean("report_caller"))
	}

	return nil
}

func applyLogrusFormatter(l *logrus.Logger, name string) error {
	if l == nil {
		return nil
	}
	switch strings.TrimSpace(strings.ToLower(name)) {
	case "text":
		l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{})
	default:
		return fmt.Errorf("unsupported formatter %q, want json or text", name)
	}
	return nil
}

// SetReportCaller enables or disables logrus caller reporting (file:line).
// When enabled, func/file skip this wrapper package so they point at the real caller.
func (p *LogrusLogger) SetReportCaller(reportCaller bool) {
	if p == nil || p.logger == nil {
		return
	}
	p.logger.SetReportCaller(reportCaller)
	if reportCaller {
		ensureReportCallerPrettyfier(p.logger)
	}
}

// SetRotateLogs sets up file rotation for the logger
func (p *LogrusLogger) SetRotateLogs(config *RotateLogsConfig) error {
	if p.logger == nil || config == nil {
		return nil
	}
	return SetupRotateLogsLogger(p.logger, config)
}

// AddRotateLogsHook adds a file rotation hook to the logger
// This allows logs to be written to both the default output and rotated files
func (p *LogrusLogger) AddRotateLogsHook(config *RotateLogsConfig) error {
	if p.logger == nil || config == nil {
		return nil
	}
	return AddRotateLogsHook(p.logger, config)
}

// With creates a child logger with specified fields
func (p *LogrusLogger) With(kvs ...any) Logger {
	lenFields := len(kvs)
	fields := make(logrus.Fields)

	for i := 0; i < lenFields; i += 2 {
		k := kvs[i]
		var v any = errors.New("MISSING VALUE")
		if i+1 < lenFields {
			v = kvs[i+1]
		}
		fields[toString(k)] = v
	}

	// Create a wrapper logger that uses WithFields for all log calls
	return &logrusLoggerWithFields{
		logger:    p.logger,
		fields:    fields,
		isShowSQL: p.isShowSQL,
		level:     p.level,
	}
}

// logrusLoggerWithFields wraps a logger with fields applied to all log calls
type logrusLoggerWithFields struct {
	logger    *logrus.Logger
	fields    logrus.Fields
	isShowSQL bool
	level     log.LogLevel
}

func (p *logrusLoggerWithFields) With(kvs ...any) Logger {
	// Merge new fields with existing fields
	lenFields := len(kvs)
	newFields := make(logrus.Fields)
	maps.Copy(newFields, p.fields)

	for i := 0; i < lenFields; i += 2 {
		k := kvs[i]
		var v any = errors.New("MISSING VALUE")
		if i+1 < lenFields {
			v = kvs[i+1]
		}
		newFields[toString(k)] = v
	}

	return &logrusLoggerWithFields{
		logger:    p.logger,
		fields:    newFields,
		isShowSQL: p.isShowSQL,
		level:     p.level,
	}
}

func (p *logrusLoggerWithFields) Log(kvs ...any) error {
	p.logger.WithFields(p.fields).Info(kvs...)
	return nil
}

func (p *logrusLoggerWithFields) Debug(kvs ...any) {
	p.logger.WithFields(p.fields).Debug(kvs...)
}

func (p *logrusLoggerWithFields) Debugf(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Debugf(msg, kvs...)
}

func (p *logrusLoggerWithFields) Info(kvs ...any) {
	p.logger.WithFields(p.fields).Info(kvs...)
}

func (p *logrusLoggerWithFields) Infof(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Infof(msg, kvs...)
}

func (p *logrusLoggerWithFields) Warn(kvs ...any) {
	p.logger.WithFields(p.fields).Warn(kvs...)
}

func (p *logrusLoggerWithFields) Warnf(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Warnf(msg, kvs...)
}

func (p *logrusLoggerWithFields) Error(kvs ...any) {
	p.logger.WithFields(p.fields).Error(kvs...)
}

func (p *logrusLoggerWithFields) Errorf(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Errorf(msg, kvs...)
}

func (p *logrusLoggerWithFields) Level() log.LogLevel {
	return p.level
}

func (p *logrusLoggerWithFields) SetLevel(l log.LogLevel) {
	p.level = l
}

func (p *logrusLoggerWithFields) ShowSQL(show ...bool) {
	p.isShowSQL = len(show) == 0 || (len(show) > 0 && show[0])
}

func (p *logrusLoggerWithFields) IsShowSQL() bool {
	return p.isShowSQL
}

func (p *logrusLoggerWithFields) Writer() io.Writer {
	if p.logger != nil {
		return p.logger.Out
	}
	return io.Discard
}

// Log prints log with kvs
func (p *LogrusLogger) Log(kvs ...any) error {
	p.Info(kvs...)
	return nil
}

// Debug prints debug information (gated by logrus level only, not databases.*.log_level).
func (p *LogrusLogger) Debug(kvs ...any) {
	p.logger.Debug(kvs...)
}

// Debugf format prints debug information
func (p *LogrusLogger) Debugf(msg string, kvs ...any) {
	p.logger.Debugf(msg, kvs...)
}

// Info prints info information
func (p *LogrusLogger) Info(kvs ...any) {
	p.logger.Info(kvs...)
}

// Infof format prints info information
func (p *LogrusLogger) Infof(msg string, kvs ...any) {
	p.logger.Infof(msg, kvs...)
}

// Warn prints warn information
func (p *LogrusLogger) Warn(kvs ...any) {
	p.logger.Warn(kvs...)
}

// Warnf format prints warn information
func (p *LogrusLogger) Warnf(msg string, kvs ...any) {
	p.logger.Warnf(msg, kvs...)
}

// Error prints error information
func (p *LogrusLogger) Error(kvs ...any) {
	p.logger.Error(kvs...)
}

// Errorf format prints error information
func (p *LogrusLogger) Errorf(msg string, kvs ...any) {
	p.logger.Errorf(msg, kvs...)
}

// Level returns current xorm filter log level
func (p *LogrusLogger) Level() log.LogLevel {
	return p.level
}

// SetLevel only filters xorm SQL via AfterSQL. Never change the shared logrus level,
// and never gate general Infof/Debugf used by application code.
func (p *LogrusLogger) SetLevel(l log.LogLevel) {
	p.level = l
}

// ShowSQL sets whether to show SQL
func (p *LogrusLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		p.isShowSQL = show[0]
	} else {
		p.isShowSQL = true
	}
}

// IsShowSQL returns whether SQL is shown
func (p *LogrusLogger) IsShowSQL() bool {
	return p.isShowSQL
}

func (p *LogrusLogger) Writer() io.Writer {
	return p.logger.Out
}

const (
	maxCallerDepth      = 32
	logrusPackagePrefix = "github.com/sirupsen/logrus"
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

// xormLogFilter applies databases.*.log_level to xorm Infof/Debugf/Warnf/Errorf/AfterSQL
// without changing the wrapped logger's application-facing methods when called directly.
type xormLogFilter struct {
	log.ContextLogger
}

// WrapXormFilter returns a ContextLogger for engine.SetLogger that respects SetLevel
// for all xorm log calls (PING, SQL, sync, …). Pass the same underlying logger to
// application code without this wrapper so app Infof is only gated by logrus level.
func WrapXormFilter(l log.Logger) log.ContextLogger {
	if l == nil {
		return nil
	}
	var inner log.ContextLogger
	switch t := l.(type) {
	case log.ContextLogger:
		inner = t
	default:
		inner = log.NewLoggerAdapter(t)
	}
	if f, ok := inner.(*xormLogFilter); ok {
		return f
	}
	return &xormLogFilter{ContextLogger: inner}
}

func (f *xormLogFilter) Debugf(format string, v ...any) {
	if f.Level() <= log.LOG_DEBUG {
		f.ContextLogger.Debugf(format, v...)
	}
}

func (f *xormLogFilter) Infof(format string, v ...any) {
	if f.Level() <= log.LOG_INFO {
		f.ContextLogger.Infof(format, v...)
	}
}

func (f *xormLogFilter) Warnf(format string, v ...any) {
	if f.Level() <= log.LOG_WARNING {
		f.ContextLogger.Warnf(format, v...)
	}
}

func (f *xormLogFilter) Errorf(format string, v ...any) {
	if f.Level() <= log.LOG_ERR {
		f.ContextLogger.Errorf(format, v...)
	}
}

func (f *xormLogFilter) BeforeSQL(ctx log.LogContext) {
	f.ContextLogger.BeforeSQL(ctx)
}

func (f *xormLogFilter) AfterSQL(ctx log.LogContext) {
	if f.Level() <= log.LOG_INFO {
		f.ContextLogger.AfterSQL(ctx)
	}
}
