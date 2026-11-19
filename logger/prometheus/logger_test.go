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

package prometheus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-kit/log/level"
)

func TestAllowOption(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"", false},
		{"info", false},
		{"DEBUG", false},
		{"trace", false},
		{"warn", false},
		{"warning", false},
		{"error", false},
		{"fatal", false},
		{"panic", false},
		{"verbose", true},
	}
	for _, c := range cases {
		_, err := allowOption(c.in)
		if c.wantErr && err == nil {
			t.Fatalf("allowOption(%q) want error", c.in)
		}
		if !c.wantErr && err != nil {
			t.Fatalf("allowOption(%q): %v", c.in, err)
		}
	}
}

func TestNewInfoLevelDropsDebug(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "app.log")

	l, err := New(&Config{
		FileName:   filename,
		Level:      "info",
		MaxBackups: 1,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := level.Debug(l).Log("msg", "debug-line"); err != nil {
		t.Fatalf("Debug.Log: %v", err)
	}
	if err := level.Info(l).Log("msg", "info-line"); err != nil {
		t.Fatalf("Info.Log: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	out := string(data)
	if strings.Contains(out, "debug-line") {
		t.Fatalf("level=info should drop debug, got: %s", out)
	}
	if !strings.Contains(out, "info-line") {
		t.Fatalf("level=info should keep info, got: %s", out)
	}
}

func TestNewDebugLevelKeepsDebug(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "app.log")

	l, err := New(&Config{
		FileName:   filename,
		Level:      "debug",
		MaxBackups: 1,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := level.Debug(l).Log("msg", "debug-line"); err != nil {
		t.Fatalf("Debug.Log: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "debug-line") {
		t.Fatalf("level=debug should keep debug, got: %s", data)
	}
}

func TestNewCallerEnabled(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "app.log")

	l, err := New(&Config{
		FileName:   filename,
		Level:      "info",
		Caller:     true,
		MaxBackups: 1,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := level.Info(l).Log("msg", "with-caller"); err != nil {
		t.Fatalf("Info.Log: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, `"caller"`) {
		t.Fatalf("caller=true should include caller field, got: %s", out)
	}
	if !strings.Contains(out, "logger_test.go:") {
		t.Fatalf("caller should point at this test file, got: %s", out)
	}
}

func TestNewCallerDisabled(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "app.log")

	l, err := New(&Config{
		FileName:   filename,
		Level:      "info",
		Caller:     false,
		MaxBackups: 1,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := level.Info(l).Log("msg", "no-caller"); err != nil {
		t.Fatalf("Info.Log: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if strings.Contains(string(data), `"caller"`) {
		t.Fatalf("caller=false should omit caller, got: %s", data)
	}
	if !strings.Contains(string(data), `"ts":"`) {
		t.Fatalf("log line should include ts, got: %s", data)
	}
}

func TestNewBadLevel(t *testing.T) {
	dir := t.TempDir()
	_, err := New(&Config{
		FileName: filepath.Join(dir, "app.log"),
		Level:    "verbose",
	})
	if err == nil {
		t.Fatal("expected error for unsupported level")
	}
}
