package logger

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"xorm.io/xorm/log"
)

// TestLogrusIntegration tests if logrus functionality is properly integrated
func TestLogrusIntegration(t *testing.T) {
	// Test creating a logrus logger
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test basic logging functions
	logger.Debug("Debug message", "key", "value")
	logger.Info("Info message", "key", "value")
	logger.Warn("Warn message", "key", "value")
	logger.Error("Error message", "key", "value")

	// Test formatted logging functions
	logger.Debugf("Debug formatted: %s", "test")
	logger.Infof("Info formatted: %s", "test")
	logger.Warnf("Warn formatted: %s", "test")
	logger.Errorf("Error formatted: %s", "test")

	t.Log("Logrus integration test passed")
}

// TestWithOperation tests the With operation for creating child loggers with fields
func TestWithOperation(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test With with single key-value pair
	childLogger := logger.With("key1", "value1")
	childLogger.Info("Message with field")
	childLogger.Debug("Debug message with field")
	childLogger.Warn("Warn message with field")
	childLogger.Error("Error message with field")

	// Test With with multiple key-value pairs
	multiLogger := logger.With("key1", "value1", "key2", "value2", "key3", "value3")
	multiLogger.Info("Message with multiple fields")
	multiLogger.Debugf("Debug formatted: %s", "test")
	multiLogger.Infof("Info formatted: %s", "test")
	multiLogger.Warnf("Warn formatted: %s", "test")
	multiLogger.Errorf("Error formatted: %s", "test")

	// Test nested With operations (chaining)
	nestedLogger := logger.With("parent", "value").With("child", "value2")
	nestedLogger.Info("Nested logger message")
	nestedLogger.Debug("Nested debug message")

	// Test With with odd number of arguments (missing value)
	oddLogger := logger.With("key1", "value1", "key2")
	oddLogger.Info("Message with missing value")

	// Test With with empty arguments
	emptyLogger := logger.With()
	emptyLogger.Info("Message with no fields")
}

// TestWithFieldsLogger tests logrusLoggerWithFields methods
func TestWithFieldsLogger(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	withFieldsLogger := logger.With("test", "value")

	// Test all logging methods
	// Note: Log method is from xorm log.Logger interface which is embedded in Logger
	withFieldsLogger.Debug("debug message")
	withFieldsLogger.Debugf("debug formatted: %s", "test")
	withFieldsLogger.Info("info message")
	withFieldsLogger.Infof("info formatted: %s", "test")
	withFieldsLogger.Warn("warn message")
	withFieldsLogger.Warnf("warn formatted: %s", "test")
	withFieldsLogger.Error("error message")
	withFieldsLogger.Errorf("error formatted: %s", "test")

	// Test Level and SetLevel
	level := withFieldsLogger.Level()
	if level == log.LOG_DEBUG {
		t.Log("Level returned LOG_DEBUG")
	}

	withFieldsLogger.SetLevel(log.LOG_INFO)
	newLevel := withFieldsLogger.Level()
	if newLevel != log.LOG_INFO {
		t.Errorf("Expected LOG_INFO, got %v", newLevel)
	}

	// Test ShowSQL and IsShowSQL
	withFieldsLogger.ShowSQL(true)
	if !withFieldsLogger.IsShowSQL() {
		t.Error("Expected IsShowSQL to return true")
	}

	withFieldsLogger.ShowSQL(false)
	if withFieldsLogger.IsShowSQL() {
		t.Error("Expected IsShowSQL to return false")
	}

	withFieldsLogger.ShowSQL() // no args, should default to true
	if !withFieldsLogger.IsShowSQL() {
		t.Error("Expected IsShowSQL to return true when ShowSQL called with no args")
	}

	// Test Writer
	writer := withFieldsLogger.Writer()
	if writer == nil {
		t.Error("Writer should not be nil")
	}
}

// TestLevelOperations tests Level and SetLevel methods
func TestLevelOperations(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test all log levels
	levels := []struct {
		xormLevel log.LogLevel
		name      string
	}{
		{log.LOG_DEBUG, "DEBUG"},
		{log.LOG_INFO, "INFO"},
		{log.LOG_WARNING, "WARNING"},
		{log.LOG_ERR, "ERROR"},
		{log.LOG_OFF, "OFF"},
	}

	for _, lvl := range levels {
		logger.SetLevel(lvl.xormLevel)
		gotLevel := logger.Level()
		if gotLevel != lvl.xormLevel {
			t.Errorf("SetLevel(%s): expected %v, got %v", lvl.name, lvl.xormLevel, gotLevel)
		}
	}
}

// TestShowSQLOperations tests ShowSQL and IsShowSQL methods
func TestShowSQLOperations(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test default value (should be false)
	if logger.IsShowSQL() {
		t.Error("Expected IsShowSQL to return false by default")
	}

	// Test ShowSQL(true)
	logger.ShowSQL(true)
	if !logger.IsShowSQL() {
		t.Error("Expected IsShowSQL to return true after ShowSQL(true)")
	}

	// Test ShowSQL(false)
	logger.ShowSQL(false)
	if logger.IsShowSQL() {
		t.Error("Expected IsShowSQL to return false after ShowSQL(false)")
	}

	// Test ShowSQL() with no args (should default to true)
	logger.ShowSQL()
	if !logger.IsShowSQL() {
		t.Error("Expected IsShowSQL to return true after ShowSQL() with no args")
	}
}

// TestWriter tests the Writer method
func TestWriter(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	writer := logger.Writer()
	if writer == nil {
		t.Error("Writer should not be nil")
	}

	// Writer should be io.Discard for the default logger
	if writer != io.Discard {
		t.Logf("Writer is not io.Discard, got: %v", writer)
	}
}

// TestNewWithLogrusLogger tests NewWithLogrusLogger function
func TestNewWithLogrusLogger(t *testing.T) {
	// Test with valid logrus logger
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(io.Discard)
	logger := NewWithLogrusLogger(logrusLogger)
	if logger == nil {
		t.Error("Expected non-nil logger")
	}

	// Test logging with the logger
	logger.Info("Test message")
	logger.Debug("Debug message")

	// Test with nil logger (should return noop)
	nilLogger := NewWithLogrusLogger(nil)
	if nilLogger == nil {
		t.Error("Expected non-nil logger (noop)")
	}

	// Verify it's a noop by checking it doesn't panic
	nilLogger.Info("Noop message")
	nilLogger.Debug("Noop debug")
}

// TestNoopLogger tests the noop logger implementation
func TestNoopLogger(t *testing.T) {
	noopLogger := Noop()

	// Test all logging methods (should not panic)
	noopLogger.Debug("test")
	noopLogger.Debugf("test: %s", "value")
	noopLogger.Info("test")
	noopLogger.Infof("test: %s", "value")
	noopLogger.Warn("test")
	noopLogger.Warnf("test: %s", "value")
	noopLogger.Error("test")
	noopLogger.Errorf("test: %s", "value")

	// Test With operation
	childNoop := noopLogger.With("key", "value")
	if childNoop == nil {
		t.Error("With should return a logger")
	}

	// Test Level and SetLevel
	noopLogger.SetLevel(log.LOG_INFO)
	newLevel := noopLogger.Level()
	if newLevel != log.LOG_INFO {
		t.Errorf("Expected level to be set, got %v", newLevel)
	}

	// Test ShowSQL and IsShowSQL
	noopLogger.ShowSQL(true)
	if noopLogger.IsShowSQL() {
		t.Error("Noop IsShowSQL should always return false")
	}

	// Test Writer
	writer := noopLogger.Writer()
	if writer != io.Discard {
		t.Errorf("Expected Writer to return io.Discard, got %v", writer)
	}
}

// TestWithChaining tests chaining multiple With calls
func TestWithChaining(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Chain multiple With calls
	chainedLogger := logger.
		With("level1", "value1").
		With("level2", "value2").
		With("level3", "value3")

	chainedLogger.Info("Chained logger message")
	chainedLogger.Debug("Chained debug message")
	chainedLogger.Warn("Chained warn message")
	chainedLogger.Error("Chained error message")
}

// TestWithFieldTypes tests With with different value types
func TestWithFieldTypes(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test with different types
	logger.With("string", "value").
		With("int", 42).
		With("bool", true).
		With("float", 3.14).
		Info("Message with different types")
}

// TestWithFieldsEdgeCases tests edge cases for With operation
func TestWithFieldsEdgeCases(t *testing.T) {
	// Test With with empty fields
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test chaining With with empty fields
	emptyFieldsLogger := logger.With()
	emptyFieldsLogger.Info("Message with empty fields")

	// Test With with only one argument (key without value)
	singleArgLogger := logger.With("key")
	singleArgLogger.Info("Message with single arg")

	// Test toString with different types
	logger.With("string", "value").
		With("int", 42).
		With("bool", true).
		With("float", 3.14).
		With("struct", struct{ A int }{A: 1}).
		With("map", map[string]int{"key": 1}).
		Info("Testing toString with different types")
}

// TestLevelEdgeCases tests edge cases for Level method
func TestLevelEdgeCases(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test all logrus levels to ensure all branches are covered
	logger.SetLevel(log.LOG_DEBUG)
	_ = logger.Level()

	// Test with nil logger in logrusLoggerWithFields
	withFieldsLogger := logger.With("test", "value")
	// Create a scenario where logger might be nil (though it shouldn't happen in practice)
	// We'll test the Level method with different log levels
	withFieldsLogger.SetLevel(log.LOG_WARNING)
	_ = withFieldsLogger.Level()

	// Test default case in Level() - set an invalid level by manipulating the underlying logger
	// This is hard to test directly, but we can test all valid levels
	levels := []log.LogLevel{
		log.LOG_DEBUG,
		log.LOG_INFO,
		log.LOG_WARNING,
		log.LOG_ERR,
		log.LOG_OFF,
	}

	for _, lvl := range levels {
		logger.SetLevel(lvl)
		gotLevel := logger.Level()
		if gotLevel != lvl {
			t.Errorf("SetLevel(%v): expected %v, got %v", lvl, lvl, gotLevel)
		}
	}
}

// TestWriterEdgeCases tests edge cases for Writer method
func TestWriterEdgeCases(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test Writer with regular logger
	writer := logger.Writer()
	if writer == nil {
		t.Error("Writer should not be nil")
	}

	// Test Writer with With fields logger
	withFieldsLogger := logger.With("test", "value")
	writer2 := withFieldsLogger.Writer()
	if writer2 == nil {
		t.Error("Writer should not be nil for withFieldsLogger")
	}
}

// TestLogMethod tests the Log method (from xorm log.Logger interface)
func TestLogMethod(t *testing.T) {
	logger, err := NewLogrusLogger()
	if err != nil {
		t.Fatalf("Failed to create logrus logger: %v", err)
	}

	// Test Log method (implements xorm log.Logger interface)
	// Note: LogrusLogger implements xorm log.Logger interface through Logger interface
	// The Log method is available through the embedded log.Logger interface
	err = logger.Log("test message")
	if err != nil {
		t.Errorf("Log should return nil, got %v", err)
	}

	// Test Log with With fields - logrusLoggerWithFields also implements log.Logger
	withLogger := logger.With("key", "value")
	// Since Logger embeds log.Logger, we can call Log directly if it's available
	// But logrusLoggerWithFields implements it, so we need to check
	if logImpl, ok := withLogger.(interface{ Log(...any) error }); ok {
		err = logImpl.Log("test message with fields")
		if err != nil {
			t.Errorf("Log should return nil, got %v", err)
		}
	}
}
