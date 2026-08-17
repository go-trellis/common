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

package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-trellis/common/config"

	"github.com/sirupsen/logrus"
)

func testLoggerConfig(t *testing.T, yaml string) config.Config {
	t.Helper()
	cfg, err := config.NewConfigOptions(config.OptionString(config.ReaderTypeYAML, yaml))
	if err != nil {
		t.Fatalf("NewConfigOptions: %v", err)
	}
	return cfg
}

func TestInitLoggerNilConfig(t *testing.T) {
	l, err := InitLogger(nil)
	if err != nil {
		t.Fatalf("InitLogger(nil): %v", err)
	}
	if l == nil {
		t.Fatal("expected noop logger")
	}
}

func TestInitLoggerNoop(t *testing.T) {
	for _, yaml := range []string{
		`type: noop`,
		`type: ""`,
		`{}`,
	} {
		t.Run(yaml, func(t *testing.T) {
			l, err := InitLogger(testLoggerConfig(t, yaml))
			if err != nil {
				t.Fatalf("InitLogger: %v", err)
			}
			if l == nil {
				t.Fatal("expected logger")
			}
		})
	}
}

func TestInitLoggerConsole(t *testing.T) {
	l, err := InitLogger(testLoggerConfig(t, `
type: console
level: 5
encoding: text
`))
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	ll, ok := l.(*logrus.Logger)
	if !ok {
		t.Fatalf("got %T, want *logrus.Logger", l)
	}
	if ll.GetLevel() != logrus.DebugLevel {
		t.Fatalf("level = %v, want debug", ll.GetLevel())
	}
	if _, ok := ll.Formatter.(*TextFormatter); !ok {
		t.Fatalf("formatter = %T, want *TextFormatter", ll.Formatter)
	}
	if !ll.ReportCaller {
		t.Fatal("ReportCaller should default to true")
	}
}

func TestInitLoggerCallerConfig(t *testing.T) {
	l, err := InitLogger(testLoggerConfig(t, `
type: console
caller: false
`))
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	ll := l.(*logrus.Logger)
	if ll.ReportCaller {
		t.Fatal("ReportCaller should be false when caller: false")
	}
}

func TestInitLoggerConsoleDefaultLevel(t *testing.T) {
	l, err := InitLogger(testLoggerConfig(t, `
type: console
`))
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	ll := l.(*logrus.Logger)
	if ll.GetLevel() != logrus.InfoLevel {
		t.Fatalf("default level = %v, want info", ll.GetLevel())
	}
}

func TestInitLoggerConsoleJSON(t *testing.T) {
	l, err := InitLogger(testLoggerConfig(t, `
type: console
level: 4
encoding: json
std_printers:
  - stderr
`))
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	ll := l.(*logrus.Logger)
	if _, ok := ll.Formatter.(*JSONFormatter); !ok {
		t.Fatalf("formatter = %T, want *JSONFormatter", ll.Formatter)
	}
	if ll.Out != os.Stderr {
		t.Fatalf("Out = %v, want stderr", ll.Out)
	}
}

func TestInitLoggerFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "app.log")

	l, err := InitLogger(testLoggerConfig(t, `
type: file
level: 4
filename: `+filename+`
move_file_type: 3
max_length: 1048576
max_backups: 5
encoding: text
`))
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	if l == nil {
		t.Fatal("expected logger")
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("log file should exist: %v", err)
	}
}

func TestInitLoggerFileCreatesNestedPathAtInfoLevel(t *testing.T) {
	// Init message is Debug; with Info level the rotate writer never Write()s
	// before the writability check — ensureLogFileWritable must still create it.
	dir := t.TempDir()
	filename := filepath.Join(dir, "nested", "app.log")

	l, err := InitLogger(testLoggerConfig(t, `
type: file
level: 4
filename: `+filename+`
encoding: text
`))
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	if l == nil {
		t.Fatal("expected logger")
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("nested log file should exist: %v", err)
	}
}

func TestInitLoggerFileMissingFilename(t *testing.T) {
	_, err := InitLogger(testLoggerConfig(t, `
type: file
level: 4
`))
	if err == nil {
		t.Fatal("expected error for missing filename")
	}
}

func TestInitLoggerUnknownType(t *testing.T) {
	_, err := InitLogger(testLoggerConfig(t, `
type: syslog
`))
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestLogFormatter(t *testing.T) {
	jsonCfg := testLoggerConfig(t, `encoding: json`)
	if _, ok := logFormatter(jsonCfg).(*JSONFormatter); !ok {
		t.Fatalf("json encoding: got %T", logFormatter(jsonCfg))
	}

	textCfg := testLoggerConfig(t, `encoding: text`)
	if _, ok := logFormatter(textCfg).(*TextFormatter); !ok {
		t.Fatalf("text encoding: got %T", logFormatter(textCfg))
	}

	defaultCfg := testLoggerConfig(t, `{}`)
	if _, ok := logFormatter(defaultCfg).(*TextFormatter); !ok {
		t.Fatalf("default encoding: got %T", logFormatter(defaultCfg))
	}
}

func TestStdWriters(t *testing.T) {
	if stdWriters(nil) != os.Stdout {
		t.Fatal("nil printers should default to stdout")
	}
	if stdWriters([]string{"unknown"}) != os.Stdout {
		t.Fatal("unknown printers should default to stdout")
	}
	if stdWriters([]string{"stdout"}) != os.Stdout {
		t.Fatal("stdout printer")
	}
	if stdWriters([]string{"stderr"}) != os.Stderr {
		t.Fatal("stderr printer")
	}

	w := stdWriters([]string{"stdout", "stderr"})
	if w == os.Stdout || w == os.Stderr {
		t.Fatal("stdout+stderr should use MultiWriter")
	}
	if _, err := w.Write(nil); err != nil {
		t.Fatalf("MultiWriter.Write: %v", err)
	}
}
