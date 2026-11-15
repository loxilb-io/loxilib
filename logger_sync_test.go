// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2026 LoxiLB Authors
package loxilib

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func testLogger(t *testing.T, output io.Writer, level LogLevelT) *Logger {
	t.Helper()
	l, err := NewLogger(output, level, false)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestDefaultLoggerConcurrentSwapAndLogging(t *testing.T) {
	l := testLogger(t, io.Discard, LogTrace)
	previous := SwapDefaultLogger(l)
	defer SwapDefaultLogger(previous)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				LogIt(LogInfo, "live %d", j)
			}
		}()
	}
	for i := 0; i < 1000; i++ {
		SwapDefaultLogger(nil)
		SwapDefaultLogger(l)
	}
	wg.Wait()
	out := &bytes.Buffer{}
	SwapDefaultLogger(testLogger(t, out, LogInfo))
	LogIt(LogInfo, "after replacement")
	SwapDefaultLogger(nil)
	LogIt(LogInfo, "must be silent")
	if !strings.Contains(out.String(), "after replacement") || strings.Contains(out.String(), "must be silent") {
		t.Fatal(out.String())
	}
}

type blockedLogWriter struct {
	entered chan struct{}
	release chan struct{}
}

func (w *blockedLogWriter) Write(p []byte) (int, error) {
	close(w.entered)
	<-w.release
	return len(p), nil
}

func TestDefaultLoggerSwapJoinsInflightCall(t *testing.T) {
	w := &blockedLogWriter{make(chan struct{}), make(chan struct{})}
	previous := SwapDefaultLogger(testLogger(t, w, LogInfo))
	defer SwapDefaultLogger(previous)
	var once sync.Once
	release := func() { once.Do(func() { close(w.release) }) }
	defer release()
	logged := make(chan struct{})
	go func() { LogIt(LogInfo, "blocked"); close(logged) }()
	select {
	case <-w.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("log did not reach writer")
	}
	swapped := make(chan struct{})
	go func() { SwapDefaultLogger(nil); close(swapped) }()
	select {
	case <-swapped:
		t.Fatal("replacement returned while old output was in use")
	case <-time.After(20 * time.Millisecond):
	}
	release()
	select {
	case <-swapped:
	case <-time.After(5 * time.Second):
		t.Fatal("replacement did not join")
	}
	<-logged
}

type logFormatter func() string

func (f logFormatter) String() string { return f() }

func TestLogFormatterCanLogDuringSwap(t *testing.T) {
	var oldOutput bytes.Buffer
	old := testLogger(t, &oldOutput, LogInfo)
	var output bytes.Buffer
	next := testLogger(t, &output, LogInfo)
	previous := SwapDefaultLogger(old)
	defer SwapDefaultLogger(previous)

	formatting := make(chan struct{})
	continueFormatting := make(chan struct{})
	var once sync.Once
	resume := func() { once.Do(func() { close(continueFormatting) }) }
	defer resume()
	logged := make(chan struct{})
	go func() {
		LogIt(LogInfo, "%s", logFormatter(func() string {
			close(formatting)
			<-continueFormatting
			LogIt(LogInfo, "nested message")
			return "original message"
		}))
		close(logged)
	}()
	select {
	case <-formatting:
	case <-time.After(5 * time.Second):
		t.Fatal("logger did not format the outer message")
	}
	swapped := make(chan *Logger, 1)
	go func() { swapped <- SwapDefaultLogger(next) }()
	deadline := time.Now().Add(5 * time.Second)
	for loadDefaultLogger() != next {
		if time.Now().After(deadline) {
			t.Fatal("replacement was not published while the old call was formatting")
		}
		time.Sleep(time.Millisecond)
	}
	select {
	case <-swapped:
		t.Fatal("swap did not wait for old call")
	default:
	}
	resume()
	select {
	case <-logged:
	case <-time.After(5 * time.Second):
		t.Fatal("nested LogIt blocked during swap")
	}
	select {
	case got := <-swapped:
		if got != old || !strings.Contains(output.String(), "nested message") {
			t.Fatalf("swap returned %p, want %p, nested output %q", got, old, output.String())
		}
		if !strings.Contains(oldOutput.String(), "original message") {
			t.Fatalf("old logger dropped original message: %q", oldOutput.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("swap failed to drain the old call")
	}
}

func TestLoggerConcurrentConfiguration(t *testing.T) {
	l := testLogger(t, io.Discard, LogTrace)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				l.Log(LogDebug, "message")
				l.LogItGetLevel()
			}
		}()
	}
	for i := 0; i < 1000; i++ {
		l.LogItSetLevel(LogInfo)
		l.LogItSetTTY(false)
		l.LogItSetLevel(LogTrace)
	}
	wg.Wait()
	if err := l.LogItSetLevel(LogTrace + 1); err == nil || l.LogItGetLevel() != LogTrace {
		t.Fatal("invalid level changed configuration")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	for level := LogEmerg; level <= LogTrace; level++ {
		out := &bytes.Buffer{}
		l := testLogger(t, out, level)
		for message := LogEmerg; message <= LogTrace; message++ {
			l.Log(message, "message-level-%d", message)
		}
		if got := strings.Count(out.String(), "message-level-"); got != int(level)+1 {
			t.Fatalf("level %d: got %d lines", level, got)
		}
	}
	if _, err := NewLogger(nil, LogInfo, false); err == nil {
		t.Fatal("nil output accepted")
	}
	if _, err := NewLogger(io.Discard, -1, false); err == nil {
		t.Fatal("invalid level accepted")
	}
}

func TestLogItInitConcurrentFirstPublication(t *testing.T) {
	previous := SwapDefaultLogger(nil)
	defer SwapDefaultLogger(previous)
	dir := t.TempDir()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			LogItInit(filepath.Join(dir, "init.log"), LogInfo, false)
			LogIt(LogInfo, "initialized")
		}()
	}
	wg.Wait()
	if SwapDefaultLogger(nil) == nil {
		t.Fatal("initialization did not publish a logger")
	}
}

func BenchmarkLogItFilteredParallel(b *testing.B) {
	logger, err := NewLogger(io.Discard, LogInfo, false)
	if err != nil {
		b.Fatal(err)
	}
	previous := SwapDefaultLogger(logger)
	b.Cleanup(func() { SwapDefaultLogger(previous) })
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			LogIt(LogDebug, "filtered request log")
		}
	})
}
