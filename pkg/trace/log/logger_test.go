package log

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/trace"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{Debug, "debug"},
		{Info, "info"},
		{Warn, "warn"},
		{Error, "error"},
		{Level(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, Warn)

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}

	var e1, e2 entry
	if err := json.Unmarshal([]byte(lines[0]), &e1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &e2); err != nil {
		t.Fatal(err)
	}

	if e1.Level != "warn" {
		t.Errorf("first line level = %q, want warn", e1.Level)
	}
	if e2.Level != "error" {
		t.Errorf("second line level = %q, want error", e2.Level)
	}
}

func TestLoggerOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, Debug)

	logger.Info("hello world", "key1", "val1", "key2", "val2")

	var e entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if e.Level != "info" {
		t.Errorf("Level = %q, want info", e.Level)
	}
	if e.Msg != "hello world" {
		t.Errorf("Msg = %q, want %q", e.Msg, "hello world")
	}
	if e.Time == "" {
		t.Error("Time is empty")
	}
	if e.Fields["key1"] != "val1" {
		t.Errorf("Fields[key1] = %q, want val1", e.Fields["key1"])
	}
	if e.Fields["key2"] != "val2" {
		t.Errorf("Fields[key2] = %q, want val2", e.Fields["key2"])
	}
}

func TestLoggerOddKVPairs(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, Debug)

	// Odd number of key-value args — last key should be dropped.
	logger.Info("test", "key1", "val1", "orphan")

	var e entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if e.Fields["key1"] != "val1" {
		t.Errorf("Fields[key1] = %q, want val1", e.Fields["key1"])
	}
	if _, ok := e.Fields["orphan"]; ok {
		t.Error("orphan key should not be in Fields")
	}
}

func TestLoggerTraceCorrelation(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, Debug)

	tracer := trace.NewInMemory()
	ctx, span := tracer.StartSpan(context.Background(), "test.op")

	logger.InfoCtx(ctx, "correlated log")

	var e entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if e.SpanID != span.ID {
		t.Errorf("SpanID = %q, want %q", e.SpanID, span.ID)
	}
	if e.TraceID == "" {
		t.Error("TraceID is empty")
	}
}

func TestLoggerCtxWithParentSpan(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, Debug)

	tracer := trace.NewInMemory()
	ctx, parent := tracer.StartSpan(context.Background(), "parent")
	ctx, child := tracer.StartSpan(ctx, "child")
	_ = parent

	logger.DebugCtx(ctx, "child log")

	var e entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatal(err)
	}

	if e.SpanID != child.ID {
		t.Errorf("SpanID = %q, want %q", e.SpanID, child.ID)
	}
	if e.TraceID != child.ParentID {
		t.Errorf("TraceID = %q, want %q (parent ID)", e.TraceID, child.ParentID)
	}
}

func TestLoggerNoCtxSpan(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, Debug)

	logger.InfoCtx(context.Background(), "no span")

	var e entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatal(err)
	}

	if e.SpanID != "" {
		t.Errorf("SpanID should be empty, got %q", e.SpanID)
	}
	if e.TraceID != "" {
		t.Errorf("TraceID should be empty, got %q", e.TraceID)
	}
}

func TestLoggerAllLevels(t *testing.T) {
	tests := []struct {
		level string
		logFn func(*Logger, string, ...string)
	}{
		{"debug", func(l *Logger, msg string, kvs ...string) { l.Debug(msg, kvs...) }},
		{"info", func(l *Logger, msg string, kvs ...string) { l.Info(msg, kvs...) }},
		{"warn", func(l *Logger, msg string, kvs ...string) { l.Warn(msg, kvs...) }},
		{"error", func(l *Logger, msg string, kvs ...string) { l.Error(msg, kvs...) }},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, Debug)
			tt.logFn(logger, "test msg")

			var e entry
			if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
				t.Fatal(err)
			}
			if e.Level != tt.level {
				t.Errorf("Level = %q, want %q", e.Level, tt.level)
			}
		})
	}
}

func TestLoggerAllCtxLevels(t *testing.T) {
	tests := []struct {
		level string
		logFn func(*Logger, context.Context, string, ...string)
	}{
		{"debug", func(l *Logger, ctx context.Context, msg string, kvs ...string) { l.DebugCtx(ctx, msg, kvs...) }},
		{"info", func(l *Logger, ctx context.Context, msg string, kvs ...string) { l.InfoCtx(ctx, msg, kvs...) }},
		{"warn", func(l *Logger, ctx context.Context, msg string, kvs ...string) { l.WarnCtx(ctx, msg, kvs...) }},
		{"error", func(l *Logger, ctx context.Context, msg string, kvs ...string) { l.ErrorCtx(ctx, msg, kvs...) }},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, Debug)
			tt.logFn(logger, context.Background(), "test msg")

			var e entry
			if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
				t.Fatal(err)
			}
			if e.Level != tt.level {
				t.Errorf("Level = %q, want %q", e.Level, tt.level)
			}
		})
	}
}

func TestLoggerConcurrency(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, Debug)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("concurrent msg")
		}()
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 50 {
		t.Errorf("expected 50 log lines, got %d", len(lines))
	}
}

func TestFileWriterCreateAndWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	fw, err := NewFileWriter(path, FileConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fw.Close() }()

	msg := []byte("hello world\n")
	n, err := fw.Write(msg)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(msg) {
		t.Errorf("Write = %d, want %d", n, len(msg))
	}

	data, _ := os.ReadFile(path)
	if string(data) != string(msg) {
		t.Errorf("file content = %q, want %q", string(data), string(msg))
	}
}

func TestFileWriterRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	fw, err := NewFileWriter(path, FileConfig{MaxSize: 20, MaxFiles: 3})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fw.Close() }()

	// Write enough to trigger rotation.
	data := []byte("12345678901234567890X") // 21 bytes > MaxSize 20
	_, err = fw.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	// Write again to the new file.
	_, _ = fw.Write([]byte("new data\n"))

	// The rotated file should exist.
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Errorf("rotated file .1 should exist: %v", err)
	}

	// The current file should have the new data.
	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "new data") {
		t.Errorf("current file should contain new data, got: %q", string(content))
	}
}

func TestFileWriterMultipleRotations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	fw, err := NewFileWriter(path, FileConfig{MaxSize: 10, MaxFiles: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fw.Close() }()

	// Three writes, each triggering rotation.
	for i := 0; i < 3; i++ {
		_, _ = fw.Write([]byte("12345678901")) // 11 bytes > 10
	}

	// MaxFiles=2, so .1 and .2 should exist, .3 should not.
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Error("rotated file .1 should exist")
	}
	if _, err := os.Stat(path + ".2"); err != nil {
		t.Error("rotated file .2 should exist")
	}
}

func TestFileWriterNoRotationWhenUnderLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	fw, err := NewFileWriter(path, FileConfig{MaxSize: 1000})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fw.Close() }()

	_, _ = fw.Write([]byte("small\n"))

	if _, err := os.Stat(path + ".1"); !os.IsNotExist(err) {
		t.Error("should not rotate under limit")
	}
}

func TestFileWriterSubdirectoryCreation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "test.log")

	fw, err := NewFileWriter(path, FileConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fw.Close() }()

	_, err = fw.Write([]byte("test\n"))
	if err != nil {
		t.Fatal(err)
	}
}
