package otel

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/lonestarx1/gogrid/pkg/trace"
)

// traceIDKey is a context key for propagating the OTLP trace ID.
type traceIDKey struct{}

// Exporter implements trace.Tracer and exports spans as OTLP JSON
// over HTTP. Spans are batched and flushed periodically or when the
// batch size is reached.
type Exporter struct {
	endpoint      string
	serviceName   string
	serviceVer    string
	batchSize     int
	flushInterval time.Duration
	client        *http.Client

	mu      sync.Mutex
	batch   []otlpSpan
	traceID map[string]string // span ID -> trace ID
	done    chan struct{}
	stopped bool
}

// Option configures an Exporter.
type Option func(*Exporter)

// WithEndpoint sets the OTLP HTTP endpoint (e.g., "http://localhost:4318/v1/traces").
func WithEndpoint(url string) Option {
	return func(e *Exporter) { e.endpoint = url }
}

// WithServiceName sets the service.name resource attribute.
func WithServiceName(name string) Option {
	return func(e *Exporter) { e.serviceName = name }
}

// WithServiceVersion sets the service.version resource attribute.
func WithServiceVersion(ver string) Option {
	return func(e *Exporter) { e.serviceVer = ver }
}

// WithBatchSize sets the maximum number of spans per flush.
func WithBatchSize(n int) Option {
	return func(e *Exporter) { e.batchSize = n }
}

// WithFlushInterval sets the time between automatic flushes.
func WithFlushInterval(d time.Duration) Option {
	return func(e *Exporter) { e.flushInterval = d }
}

// WithHTTPClient sets a custom HTTP client for exporting.
func WithHTTPClient(c *http.Client) Option {
	return func(e *Exporter) { e.client = c }
}

// NewExporter creates an OTLP JSON exporter.
func NewExporter(opts ...Option) *Exporter {
	e := &Exporter{
		endpoint:      "http://localhost:4318/v1/traces",
		serviceName:   "gogrid",
		batchSize:     100,
		flushInterval: 5 * time.Second,
		client:        &http.Client{Timeout: 10 * time.Second},
		traceID:       make(map[string]string),
		done:          make(chan struct{}),
	}
	for _, opt := range opts {
		opt(e)
	}

	go e.flushLoop()
	return e
}

// StartSpan begins a new span and assigns an OTLP trace ID.
func (e *Exporter) StartSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	ctx, span := trace.NewSpan(ctx, name)

	// Propagate or generate a trace ID.
	tid, _ := ctx.Value(traceIDKey{}).(string)
	if tid == "" {
		tid = deriveID(span.ID, 32)
	}
	ctx = context.WithValue(ctx, traceIDKey{}, tid)

	e.mu.Lock()
	e.traceID[span.ID] = tid
	e.mu.Unlock()

	return ctx, span
}

// EndSpan records the span end time and adds it to the export batch.
func (e *Exporter) EndSpan(span *trace.Span) {
	span.EndTime = time.Now()

	e.mu.Lock()
	defer e.mu.Unlock()

	tid := e.traceID[span.ID]
	delete(e.traceID, span.ID)

	os := convertSpan(span, tid)
	e.batch = append(e.batch, os)

	if len(e.batch) >= e.batchSize {
		_ = e.flushLocked()
	}
}

// Flush sends all buffered spans to the endpoint.
func (e *Exporter) Flush() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.flushLocked()
}

// Shutdown stops the background flush loop and sends remaining spans.
func (e *Exporter) Shutdown() error {
	e.mu.Lock()
	if !e.stopped {
		e.stopped = true
		close(e.done)
	}
	e.mu.Unlock()
	return e.Flush()
}

// BatchLen returns the number of buffered spans (for testing).
func (e *Exporter) BatchLen() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.batch)
}

func (e *Exporter) flushLoop() {
	ticker := time.NewTicker(e.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = e.Flush()
		case <-e.done:
			return
		}
	}
}

func (e *Exporter) flushLocked() error {
	if len(e.batch) == 0 {
		return nil
	}

	spans := make([]otlpSpan, len(e.batch))
	copy(spans, e.batch)
	e.batch = e.batch[:0]

	payload := otlpPayload{
		ResourceSpans: []resourceSpans{{
			Resource: resource{
				Attributes: []attribute{
					{Key: "service.name", Value: attrValue{StringValue: e.serviceName}},
				},
			},
			ScopeSpans: []scopeSpans{{
				Scope: scope{Name: "gogrid"},
				Spans: spans,
			}},
		}},
	}

	if e.serviceVer != "" {
		payload.ResourceSpans[0].Resource.Attributes = append(
			payload.ResourceSpans[0].Resource.Attributes,
			attribute{Key: "service.version", Value: attrValue{StringValue: e.serviceVer}},
		)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("otel: marshal: %w", err)
	}

	req, err := http.NewRequest("POST", e.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("otel: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("otel: export: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("otel: export: HTTP %d", resp.StatusCode)
	}
	return nil
}

// --- OTLP JSON structures ---

type otlpPayload struct {
	ResourceSpans []resourceSpans `json:"resourceSpans"`
}

type resourceSpans struct {
	Resource   resource     `json:"resource"`
	ScopeSpans []scopeSpans `json:"scopeSpans"`
}

type resource struct {
	Attributes []attribute `json:"attributes"`
}

type scopeSpans struct {
	Scope scope      `json:"scope"`
	Spans []otlpSpan `json:"spans"`
}

type scope struct {
	Name string `json:"name"`
}

type otlpSpan struct {
	TraceID           string      `json:"traceId"`
	SpanID            string      `json:"spanId"`
	ParentSpanID      string      `json:"parentSpanId,omitempty"`
	Name              string      `json:"name"`
	Kind              int         `json:"kind"`
	StartTimeUnixNano string      `json:"startTimeUnixNano"`
	EndTimeUnixNano   string      `json:"endTimeUnixNano"`
	Attributes        []attribute `json:"attributes,omitempty"`
	Status            otlpStatus  `json:"status"`
}

type attribute struct {
	Key   string    `json:"key"`
	Value attrValue `json:"value"`
}

type attrValue struct {
	StringValue string `json:"stringValue,omitempty"`
}

type otlpStatus struct {
	Code int `json:"code"`
}

func convertSpan(s *trace.Span, traceID string) otlpSpan {
	os := otlpSpan{
		TraceID:           traceID,
		SpanID:            deriveID(s.ID, 16),
		Name:              s.Name,
		Kind:              1, // SPAN_KIND_INTERNAL
		StartTimeUnixNano: strconv.FormatInt(s.StartTime.UnixNano(), 10),
		EndTimeUnixNano:   strconv.FormatInt(s.EndTime.UnixNano(), 10),
	}

	if s.ParentID != "" {
		os.ParentSpanID = deriveID(s.ParentID, 16)
	}

	if s.Status == trace.StatusError {
		os.Status = otlpStatus{Code: 2} // STATUS_CODE_ERROR
	} else {
		os.Status = otlpStatus{Code: 1} // STATUS_CODE_OK
	}

	for k, v := range s.Attributes {
		os.Attributes = append(os.Attributes, attribute{
			Key:   "gogrid." + k,
			Value: attrValue{StringValue: v},
		})
	}

	if s.Error != "" {
		os.Attributes = append(os.Attributes, attribute{
			Key:   "exception.message",
			Value: attrValue{StringValue: s.Error},
		})
	}

	return os
}

// deriveID produces a hex string of the given length from a source ID
// using SHA-256 hashing.
func deriveID(src string, hexLen int) string {
	h := sha256.Sum256([]byte(src))
	full := hex.EncodeToString(h[:])
	if len(full) > hexLen {
		return full[:hexLen]
	}
	return full
}
