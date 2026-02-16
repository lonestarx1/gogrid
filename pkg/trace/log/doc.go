// Package log provides structured JSON logging with trace correlation
// for GoGrid.
//
// The Logger writes JSON log lines with level, timestamp, message, and
// optional fields. When a trace span exists in the context, the logger
// automatically includes trace_id and span_id for correlation.
//
// Usage:
//
//	logger := log.New(os.Stdout, log.Info)
//	logger.InfoCtx(ctx, "agent started", "agent", "researcher", "model", "gpt-4o")
//
// For file logging with rotation:
//
//	fw, err := log.NewFileWriter("/var/log/gogrid.log", log.FileConfig{
//	    MaxSize:  10 * 1024 * 1024, // 10 MB
//	    MaxFiles: 5,
//	})
//	logger := log.New(fw, log.Debug)
package log
