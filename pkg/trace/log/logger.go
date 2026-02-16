package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lonestarx1/gogrid/pkg/trace"
)

// Level represents log severity.
type Level int

const (
	// Debug is the most verbose level.
	Debug Level = iota
	// Info is the default level.
	Info
	// Warn indicates a potential issue.
	Warn
	// Error indicates a failure.
	Error
)

// String returns the level name.
func (l Level) String() string {
	switch l {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// entry is the JSON structure for a log line.
type entry struct {
	Level   string            `json:"level"`
	Time    string            `json:"time"`
	Msg     string            `json:"msg"`
	TraceID string            `json:"trace_id,omitempty"`
	SpanID  string            `json:"span_id,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Logger writes structured JSON log lines with optional trace correlation.
type Logger struct {
	mu    sync.Mutex
	out   io.Writer
	level Level
}

// New creates a Logger that writes to out at the given minimum level.
func New(out io.Writer, level Level) *Logger {
	return &Logger{out: out, level: level}
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, kvs ...string) {
	l.log(context.Background(), Debug, msg, kvs)
}

// Info logs at info level.
func (l *Logger) Info(msg string, kvs ...string) {
	l.log(context.Background(), Info, msg, kvs)
}

// Warn logs at warn level.
func (l *Logger) Warn(msg string, kvs ...string) {
	l.log(context.Background(), Warn, msg, kvs)
}

// Error logs at error level.
func (l *Logger) Error(msg string, kvs ...string) {
	l.log(context.Background(), Error, msg, kvs)
}

// DebugCtx logs at debug level with trace correlation from ctx.
func (l *Logger) DebugCtx(ctx context.Context, msg string, kvs ...string) {
	l.log(ctx, Debug, msg, kvs)
}

// InfoCtx logs at info level with trace correlation from ctx.
func (l *Logger) InfoCtx(ctx context.Context, msg string, kvs ...string) {
	l.log(ctx, Info, msg, kvs)
}

// WarnCtx logs at warn level with trace correlation from ctx.
func (l *Logger) WarnCtx(ctx context.Context, msg string, kvs ...string) {
	l.log(ctx, Warn, msg, kvs)
}

// ErrorCtx logs at error level with trace correlation from ctx.
func (l *Logger) ErrorCtx(ctx context.Context, msg string, kvs ...string) {
	l.log(ctx, Error, msg, kvs)
}

func (l *Logger) log(ctx context.Context, level Level, msg string, kvs []string) {
	if level < l.level {
		return
	}

	e := entry{
		Level: level.String(),
		Time:  time.Now().UTC().Format(time.RFC3339Nano),
		Msg:   msg,
	}

	// Trace correlation.
	if span := trace.SpanFromContext(ctx); span != nil {
		e.SpanID = span.ID
		if span.ParentID != "" {
			e.TraceID = span.ParentID
		} else {
			e.TraceID = span.ID
		}
	}

	// Key-value pairs.
	if len(kvs) > 0 {
		e.Fields = make(map[string]string, len(kvs)/2)
		for i := 0; i+1 < len(kvs); i += 2 {
			e.Fields[kvs[i]] = kvs[i+1]
		}
	}

	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	data = append(data, '\n')

	l.mu.Lock()
	_, _ = l.out.Write(data)
	l.mu.Unlock()
}

// FileConfig controls file-based log rotation.
type FileConfig struct {
	// MaxSize is the maximum size in bytes before rotation. 0 means no limit.
	MaxSize int64
	// MaxFiles is the maximum number of rotated files to keep. 0 means keep all.
	MaxFiles int
}

// FileWriter is a log writer that rotates files by size.
type FileWriter struct {
	mu     sync.Mutex
	path   string
	config FileConfig
	file   *os.File
	size   int64
}

// NewFileWriter creates a file-based log writer with rotation.
func NewFileWriter(path string, config FileConfig) (*FileWriter, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("log: create directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("log: open file: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("log: stat file: %w", err)
	}

	return &FileWriter{
		path:   path,
		config: config,
		file:   f,
		size:   info.Size(),
	}, nil
}

// Write implements io.Writer. Rotates the file if MaxSize is exceeded.
func (fw *FileWriter) Write(p []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.config.MaxSize > 0 && fw.size+int64(len(p)) > fw.config.MaxSize {
		if err := fw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := fw.file.Write(p)
	fw.size += int64(n)
	return n, err
}

// Close closes the underlying file.
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.file.Close()
}

func (fw *FileWriter) rotate() error {
	_ = fw.file.Close()

	// Shift existing rotated files.
	if fw.config.MaxFiles > 0 {
		// Remove the oldest if at limit.
		oldest := fmt.Sprintf("%s.%d", fw.path, fw.config.MaxFiles)
		_ = os.Remove(oldest)

		// Shift .N-1 -> .N, .N-2 -> .N-1, ...
		for i := fw.config.MaxFiles - 1; i >= 1; i-- {
			src := fmt.Sprintf("%s.%d", fw.path, i)
			dst := fmt.Sprintf("%s.%d", fw.path, i+1)
			_ = os.Rename(src, dst)
		}
	}

	// Rename current to .1
	_ = os.Rename(fw.path, fw.path+".1")

	// Open a fresh file.
	f, err := os.OpenFile(fw.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("log: rotate: %w", err)
	}
	fw.file = f
	fw.size = 0
	return nil
}
