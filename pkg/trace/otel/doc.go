// Package otel provides an OTLP-compatible span exporter for GoGrid.
//
// The Exporter implements trace.Tracer and batches completed spans,
// flushing them as OTLP JSON over HTTP to any OTLP-compatible backend
// (Jaeger, Zipkin, Grafana Tempo, etc.). No external dependencies are
// required — the exporter uses only the Go standard library.
//
// Usage:
//
//	exporter := otel.NewExporter(
//	    otel.WithEndpoint("http://localhost:4318/v1/traces"),
//	    otel.WithServiceName("my-agent-service"),
//	    otel.WithBatchSize(100),
//	    otel.WithFlushInterval(5 * time.Second),
//	)
//	defer exporter.Shutdown()
//
//	a := agent.New("assistant", agent.WithTracer(exporter))
package otel
