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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// blockingWriter blocks every Write until release is closed.
type blockingWriter struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	release chan struct{}
}

func (b *blockingWriter) Write(p []byte) (int, error) {
	<-b.release
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}
func (b *blockingWriter) Close() error { return nil }

func (b *blockingWriter) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestAsyncWriter_WritesAllAndFlushes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	w, err := NewRotateLogsWithConfig(&RotateConfig{FileName: path, Async: true})
	if err != nil {
		t.Fatal(err)
	}
	aw := w.(*asyncWriter)

	const n = 1000
	for i := 0; i < n; i++ {
		if _, err := w.Write([]byte(fmt.Sprintf("line %d\n", i))); err != nil {
			t.Fatal(err)
		}
	}
	if err := aw.Flush(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), "\n"); got != n {
		t.Fatalf("lines = %d, want %d", got, n)
	}
	if err := aw.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("x\n")); err == nil {
		t.Fatal("write after Close should fail")
	}
}

func TestAsyncWriter_CloseDrainsQueue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	w, err := NewRotateLogsWithConfig(&RotateConfig{FileName: path, Async: true})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		w.Write([]byte("x\n"))
	}
	if err := w.(*asyncWriter).Close(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if got := strings.Count(string(data), "\n"); got != 100 {
		t.Fatalf("lines = %d, want 100", got)
	}
}

func TestAsyncWriter_DropsLowLevelsWhenFull(t *testing.T) {
	inner := &blockingWriter{release: make(chan struct{})}
	aw := newAsyncWriter(inner, 2)
	defer func() { close(inner.release); aw.Close() }()

	info := &levelWriter{a: aw, mode: writeModeForLevel(InfoLevel)}

	// The drain goroutine takes the first line and blocks inside inner.Write;
	// the next two fill the queue; everything after that must be dropped
	// without blocking the caller.
	info.Write([]byte("a\n"))
	deadline := time.Now().Add(time.Second)
	for len(aw.ch) != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	info.Write([]byte("b\n"))
	info.Write([]byte("c\n"))

	start := time.Now()
	for i := 0; i < 100; i++ {
		if _, err := info.Write([]byte("dropped\n")); err != nil {
			t.Fatal(err)
		}
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("Info writes blocked on a full queue")
	}
	if got := aw.Dropped(); got != 100 {
		t.Fatalf("dropped = %d, want 100", got)
	}
}

func TestAsyncWriter_ErrorBlocksUntilSpace(t *testing.T) {
	inner := &blockingWriter{release: make(chan struct{})}
	aw := newAsyncWriter(inner, 1)
	defer aw.Close()

	errW := &levelWriter{a: aw, mode: writeModeForLevel(ErrorLevel)}
	errW.Write([]byte("1\n")) // taken by drain goroutine
	deadline := time.Now().Add(time.Second)
	for len(aw.ch) != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	errW.Write([]byte("2\n")) // fills queue

	finished := make(chan struct{})
	go func() {
		errW.Write([]byte("3\n")) // must wait, not drop
		close(finished)
	}()
	select {
	case <-finished:
		t.Fatal("Error write should block while the queue is full")
	case <-time.After(50 * time.Millisecond):
	}

	close(inner.release)
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("Error write did not complete after space freed")
	}
	if aw.Dropped() != 0 {
		t.Fatalf("dropped = %d, want 0", aw.Dropped())
	}
	aw.Flush()
	if got := inner.String(); got != "1\n2\n3\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestAsyncWriter_FatalIsSynchronous(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	w, err := NewRotateLogsWithConfig(&RotateConfig{FileName: path, Async: true})
	if err != nil {
		t.Fatal(err)
	}
	defer w.(*asyncWriter).Close()

	fatal := levelWriterFor(w, FatalLevel)
	if _, err := fatal.Write([]byte("fatal\n")); err != nil {
		t.Fatal(err)
	}
	// No Flush: the line must already be on disk when Write returns.
	data, _ := os.ReadFile(path)
	if string(data) != "fatal\n" {
		t.Fatalf("file = %q, want fatal line", data)
	}
}

func TestAsyncWriter_RotatesBySize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	w, err := NewRotateLogsWithConfig(&RotateConfig{
		FileName:     path,
		Async:        true,
		RotationSize: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	aw := w.(*asyncWriter)
	for i := 0; i < 20; i++ {
		w.Write([]byte(strings.Repeat("x", 49) + "\n"))
		aw.Flush() // one batch per line so size checks happen between lines
	}
	aw.Close()

	matches, _ := filepath.Glob(path + ".*")
	if len(matches) == 0 {
		t.Fatal("expected size-based rotation to create archives")
	}
}

func TestRotatingFileWriter_SizeTrackedWithoutStat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	// Pre-existing content must count toward the size limit.
	if err := os.WriteFile(path, []byte(strings.Repeat("y", 90)), 0o644); err != nil {
		t.Fatal(err)
	}
	w, err := newRotatingFileWriterFromRotateConfig(&RotateConfig{
		FileName:     path,
		RotationSize: 100,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	w.Write([]byte(strings.Repeat("a", 20))) // 90+20 >= 100 after this write
	w.Write([]byte("b\n"))                   // triggers rotation first

	matches, _ := filepath.Glob(path + ".*")
	if len(matches) != 1 {
		t.Fatalf("archives = %v, want exactly 1", matches)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "b\n" {
		t.Fatalf("active file = %q", data)
	}
}
