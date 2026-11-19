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
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// defaultAsyncQueueSize is the number of log lines that can be pending
	// before low-severity lines start being dropped.
	defaultAsyncQueueSize = 8192
	// asyncBatchBytes caps how many bytes are coalesced into one write syscall.
	asyncBatchBytes = 64 << 10
	// asyncBlockTimeout bounds how long a Warn/Error line waits for queue space
	// so a stalled disk can never hang a request handler indefinitely.
	asyncBlockTimeout = time.Second
	// asyncReportInterval throttles drop/error reports written to stderr.
	asyncReportInterval = time.Second
)

var errAsyncWriterClosed = errors.New("logger: async writer closed")

// writeMode controls what happens to a line when the queue is full.
type writeMode int

const (
	// writeModeDrop never blocks; the line is dropped when the queue is full.
	writeModeDrop writeMode = iota
	// writeModeBlock waits (up to asyncBlockTimeout) for queue space.
	writeModeBlock
	// writeModeSync waits for queue space and until the line is written to the
	// underlying writer. Used for Panic/Fatal where the process is about to die.
	writeModeSync
)

// writeModeForLevel picks the backpressure policy for a log level:
// Debug/Info/Trace are droppable, Warn/Error block briefly, Fatal/Panic are
// written synchronously.
func writeModeForLevel(l Level) writeMode {
	switch {
	case l <= FatalLevel:
		return writeModeSync
	case l <= WarnLevel:
		return writeModeBlock
	default:
		return writeModeDrop
	}
}

type asyncMsg struct {
	data []byte
	ack  chan struct{} // closed once data has been handed to the inner writer
}

// asyncWriter moves file I/O off the caller's goroutine. Callers copy the line
// into a bounded channel; a single goroutine drains it, coalesces whatever is
// pending into one buffer and issues one Write to the inner writer per batch.
// Under load that turns N write syscalls (and N lock acquisitions in the
// rotating writer) into roughly one per batch.
type asyncWriter struct {
	inner io.WriteCloser
	ch    chan asyncMsg

	done      chan struct{} // closed by Close
	stopped   chan struct{} // closed when the drain goroutine has exited
	closeOnce sync.Once
	closeErr  error

	dropped  atomic.Uint64
	reported uint64 // drain goroutine only
	lastRep  time.Time
	lastErr  time.Time
}

func newAsyncWriter(inner io.WriteCloser, queueSize int) *asyncWriter {
	if queueSize <= 0 {
		queueSize = defaultAsyncQueueSize
	}
	a := &asyncWriter{
		inner:   inner,
		ch:      make(chan asyncMsg, queueSize),
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	registerAsync(a)
	go a.run()
	return a
}

// Write implements io.Writer using writeModeBlock.
func (a *asyncWriter) Write(p []byte) (int, error) {
	return a.submit(p, writeModeBlock)
}

// Dropped returns how many lines have been discarded because the queue was full.
func (a *asyncWriter) Dropped() uint64 { return a.dropped.Load() }

// Flush blocks until everything queued before the call has been written.
func (a *asyncWriter) Flush() error {
	select {
	case <-a.done:
		return errAsyncWriterClosed
	default:
	}
	m := asyncMsg{ack: make(chan struct{})}
	select {
	case a.ch <- m:
	case <-a.done:
		return errAsyncWriterClosed
	}
	select {
	case <-m.ack:
	case <-a.stopped:
	}
	return nil
}

// Close drains the queue, closes the inner writer and stops the goroutine.
func (a *asyncWriter) Close() error {
	a.closeOnce.Do(func() {
		close(a.done)
		<-a.stopped
		unregisterAsync(a)
	})
	return a.closeErr
}

func (a *asyncWriter) submit(p []byte, mode writeMode) (int, error) {
	select {
	case <-a.done:
		return 0, errAsyncWriterClosed
	default:
	}

	// io.Writer must not retain p.
	m := asyncMsg{data: append([]byte(nil), p...)}
	if mode == writeModeSync {
		m.ack = make(chan struct{})
	}

	if mode == writeModeDrop {
		select {
		case a.ch <- m:
		default:
			a.dropped.Add(1)
			return len(p), nil
		}
	} else {
		select {
		case a.ch <- m:
		default:
			t := time.NewTimer(asyncBlockTimeout)
			select {
			case a.ch <- m:
				t.Stop()
			case <-a.done:
				t.Stop()
				return 0, errAsyncWriterClosed
			case <-t.C:
				a.dropped.Add(1)
				return len(p), nil
			}
		}
	}

	if m.ack != nil {
		select {
		case <-m.ack:
		case <-a.stopped:
		}
	}
	return len(p), nil
}

func (a *asyncWriter) run() {
	defer close(a.stopped)

	buf := make([]byte, 0, asyncBatchBytes)
	var acks []chan struct{}

	for {
		select {
		case m := <-a.ch:
			buf, acks = a.collect(buf[:0], acks[:0], m)
			a.flushBatch(buf, acks)
		case <-a.done:
			// Close was requested: write whatever is still queued, then stop.
			for {
				buf, acks = buf[:0], acks[:0]
				select {
				case m := <-a.ch:
					buf, acks = a.collect(buf, acks, m)
					a.flushBatch(buf, acks)
				default:
					a.closeErr = a.inner.Close()
					return
				}
			}
		}
	}
}

// collect appends m and then whatever else is immediately available, up to
// asyncBatchBytes.
func (a *asyncWriter) collect(buf []byte, acks []chan struct{}, m asyncMsg) ([]byte, []chan struct{}) {
	buf = append(buf, m.data...)
	if m.ack != nil {
		acks = append(acks, m.ack)
	}
	for len(buf) < asyncBatchBytes {
		select {
		case m = <-a.ch:
			buf = append(buf, m.data...)
			if m.ack != nil {
				acks = append(acks, m.ack)
			}
		default:
			return buf, acks
		}
	}
	return buf, acks
}

func (a *asyncWriter) flushBatch(buf []byte, acks []chan struct{}) {
	if len(buf) > 0 {
		if _, err := a.inner.Write(buf); err != nil {
			a.reportf(&a.lastErr, "logger: async write failed: %v", err)
		}
	}
	for _, ack := range acks {
		close(ack)
	}
	if d := a.dropped.Load(); d != a.reported {
		a.reportf(&a.lastRep, "logger: dropped %d log lines (queue full), %d total", d-a.reported, d)
		a.reported = d
	}
}

// reportf writes a diagnostic to stderr, at most once per asyncReportInterval
// per category, so a failing disk can't turn into a stderr flood.
func (a *asyncWriter) reportf(last *time.Time, format string, args ...any) {
	now := time.Now()
	if now.Sub(*last) < asyncReportInterval {
		return
	}
	*last = now
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// levelWriter is a per-level view of an asyncWriter carrying that level's
// backpressure policy. It lets one queue/file be shared by all levels while
// Debug/Info are dropped first.
type levelWriter struct {
	a    *asyncWriter
	mode writeMode
}

func (w *levelWriter) Write(p []byte) (int, error) { return w.a.submit(p, w.mode) }

// registry of live async writers so applications can flush on shutdown.
var (
	asyncMu      sync.Mutex
	asyncWriters = map[*asyncWriter]struct{}{}
)

func registerAsync(a *asyncWriter) {
	asyncMu.Lock()
	asyncWriters[a] = struct{}{}
	asyncMu.Unlock()
}

func unregisterAsync(a *asyncWriter) {
	asyncMu.Lock()
	delete(asyncWriters, a)
	asyncMu.Unlock()
}

func snapshotAsync() []*asyncWriter {
	asyncMu.Lock()
	defer asyncMu.Unlock()
	out := make([]*asyncWriter, 0, len(asyncWriters))
	for a := range asyncWriters {
		out = append(out, a)
	}
	return out
}

// Flush blocks until all log lines queued so far by async file writers have
// been written. Call it before the process exits (e.g. in a shutdown hook).
func Flush() {
	for _, a := range snapshotAsync() {
		_ = a.Flush()
	}
}

// Shutdown flushes and closes all async file writers. Loggers created before
// the call must not be used afterwards.
func Shutdown() {
	for _, a := range snapshotAsync() {
		_ = a.Close()
	}
}
