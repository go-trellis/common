package logger

import (
	"xorm.io/xorm/log"
)

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
