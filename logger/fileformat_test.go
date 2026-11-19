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
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Even when the logger's own formatter would colorize (terminal stderr), the
// file must get plain text with a real timestamp.
func TestFileHookTextHasRealTimestampAndNoColors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")

	l, err := NewLogger(&LogrusConfig{
		Level:         InfoLevel,
		Formatter:     &TextFormatter{ForceColors: true},
		DefaultWriter: io.Discard,
		Configs: []any{&LugrusRotateConfig{
			Levels:       allLevels,
			RotateConfig: &RotateConfig{FileName: path},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	l.Info("hello")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("file output must not contain ANSI codes: %q", out)
	}
	if !strings.Contains(out, `time="`+strconv.Itoa(time.Now().Year())) {
		t.Fatalf("file output should contain a full timestamp: %q", out)
	}
}

func TestFileFormatterKeepsUserOptions(t *testing.T) {
	f := fileFormatter(&TextFormatter{DisableTimestamp: true, TimestampFormat: "15:04"}).(*TextFormatter)
	if !f.DisableTimestamp || f.TimestampFormat != "15:04" {
		t.Fatalf("user options lost: %+v", f)
	}
	if !f.DisableColors || !f.FullTimestamp {
		t.Fatalf("colors should be off and full timestamp on: %+v", f)
	}

	j := &JSONFormatter{}
	if fileFormatter(j) != j {
		t.Fatal("JSON formatter should be returned unchanged")
	}
}
