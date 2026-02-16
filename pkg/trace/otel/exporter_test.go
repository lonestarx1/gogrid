package otel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/lonestarx1/gogrid/pkg/trace"
)

func TestNewExporterDefaults(t *testing.T) {
	e := NewExporter()
	defer func() { _ = e.Shutdown() }()

	if e.endpoint != "http://localhost:4318/v1/traces" {
		t.Errorf("endpoint = %q, want default", e.endpoint)
	}
	if e.serviceName != "gogrid" {
		t.Errorf("serviceName = %q, want gogrid", e.serviceName)
	}
	if e.batchSize != 100 {
		t.Errorf("batchSize = %d, want 100", e.batchSize)
	}
}

func TestOptions(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	e := NewExporter(
		WithEndpoint("http://custom:4318/v1/traces"),
		WithServiceName("myservice"),
		WithServiceVersion("1.0.0"),
		WithBatchSize(50),
		WithFlushInterval(10*time.Second),
		WithHTTPClient(client),
	)
	defer func() { _ = e.Shutdown() }()

	if e.endpoint != "http://custom:4318/v1/traces" {
		t.Errorf("endpoint = %q", e.endpoint)
	}
	if e.serviceName != "myservice" {
		t.Errorf("serviceName = %q", e.serviceName)
	}
	if e.serviceVer != "1.0.0" {
		t.Errorf("serviceVer = %q", e.serviceVer)
	}
	if e.batchSize != 50 {
		t.Errorf("batchSize = %d", e.batchSize)
	}
	if e.client != client {
		t.Error("client not set")
	}
}

func TestStartSpanCreatesValidSpan(t *testing.T) {
	e := NewExporter(WithFlushInterval(time.Hour))
	defer func() { _ = e.Shutdown() }()

	ctx, span := e.StartSpan(context.Background(), "test.op")
	if span == nil {
		t.Fatal("span is nil")
	}
	if span.Name != "test.op" {
		t.Errorf("Name = %q, want test.op", span.Name)
	}
	if span.ID == "" {
		t.Error("ID is empty")
	}
	if ctx == nil {
		t.Fatal("ctx is nil")
	}
}

func TestStartSpanPropagatesTraceID(t *testing.T) {
	e := NewExporter(WithFlushInterval(time.Hour))
	defer func() { _ = e.Shutdown() }()

	ctx, parent := e.StartSpan(context.Background(), "parent")
	_, child := e.StartSpan(ctx, "child")

	e.mu.Lock()
	parentTID := e.traceID[parent.ID]
	childTID := e.traceID[child.ID]
	e.mu.Unlock()

	if parentTID == "" {
		t.Error("parent trace ID is empty")
	}
	if childTID == "" {
		t.Error("child trace ID is empty")
	}
	if parentTID != childTID {
		t.Errorf("trace IDs differ: parent=%q child=%q", parentTID, childTID)
	}
}

func TestEndSpanAddsToBatch(t *testing.T) {
	e := NewExporter(WithFlushInterval(time.Hour))
	defer func() { _ = e.Shutdown() }()

	_, span := e.StartSpan(context.Background(), "test.op")
	e.EndSpan(span)

	if got := e.BatchLen(); got != 1 {
		t.Errorf("BatchLen = %d, want 1", got)
	}
}

func TestEndSpanSetsEndTime(t *testing.T) {
	e := NewExporter(WithFlushInterval(time.Hour))
	defer func() { _ = e.Shutdown() }()

	_, span := e.StartSpan(context.Background(), "test.op")
	e.EndSpan(span)

	if span.EndTime.IsZero() {
		t.Error("EndTime is zero after EndSpan")
	}
}

func TestFlushSendsToEndpoint(t *testing.T) {
	var mu sync.Mutex
	var received []byte
	var contentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		contentType = r.Header.Get("Content-Type")
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	e := NewExporter(
		WithEndpoint(srv.URL),
		WithFlushInterval(time.Hour),
	)
	defer func() { _ = e.Shutdown() }()

	_, span := e.StartSpan(context.Background(), "test.op")
	span.SetAttribute("key", "value")
	e.EndSpan(span)

	if err := e.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	var payload otlpPayload
	if err := json.Unmarshal(received, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(payload.ResourceSpans) != 1 {
		t.Fatalf("ResourceSpans len = %d, want 1", len(payload.ResourceSpans))
	}
	rs := payload.ResourceSpans[0]

	// Check service.name attribute.
	found := false
	for _, attr := range rs.Resource.Attributes {
		if attr.Key == "service.name" && attr.Value.StringValue == "gogrid" {
			found = true
		}
	}
	if !found {
		t.Error("service.name attribute not found")
	}

	if len(rs.ScopeSpans) != 1 {
		t.Fatalf("ScopeSpans len = %d, want 1", len(rs.ScopeSpans))
	}

	spans := rs.ScopeSpans[0].Spans
	if len(spans) != 1 {
		t.Fatalf("spans len = %d, want 1", len(spans))
	}

	os := spans[0]
	if os.Name != "test.op" {
		t.Errorf("span name = %q, want test.op", os.Name)
	}
	if os.TraceID == "" {
		t.Error("traceId is empty")
	}
	if os.SpanID == "" {
		t.Error("spanId is empty")
	}
	if os.Kind != 1 {
		t.Errorf("kind = %d, want 1", os.Kind)
	}
}

func TestFlushWithServiceVersion(t *testing.T) {
	var mu sync.Mutex
	var received []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	e := NewExporter(
		WithEndpoint(srv.URL),
		WithServiceVersion("2.0.0"),
		WithFlushInterval(time.Hour),
	)
	defer func() { _ = e.Shutdown() }()

	_, span := e.StartSpan(context.Background(), "test")
	e.EndSpan(span)
	_ = e.Flush()

	mu.Lock()
	defer mu.Unlock()

	var payload otlpPayload
	_ = json.Unmarshal(received, &payload)

	found := false
	for _, attr := range payload.ResourceSpans[0].Resource.Attributes {
		if attr.Key == "service.version" && attr.Value.StringValue == "2.0.0" {
			found = true
		}
	}
	if !found {
		t.Error("service.version attribute not found")
	}
}

func TestFlushEmptyBatch(t *testing.T) {
	e := NewExporter(WithFlushInterval(time.Hour))
	defer func() { _ = e.Shutdown() }()

	if err := e.Flush(); err != nil {
		t.Errorf("Flush empty batch should not error: %v", err)
	}
}

func TestBatchSizeAutoFlush(t *testing.T) {
	flushed := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		select {
		case flushed <- struct{}{}:
		default:
		}
	}))
	defer srv.Close()

	e := NewExporter(
		WithEndpoint(srv.URL),
		WithBatchSize(3),
		WithFlushInterval(time.Hour),
	)
	defer func() { _ = e.Shutdown() }()

	for i := 0; i < 3; i++ {
		_, span := e.StartSpan(context.Background(), "test")
		e.EndSpan(span)
	}

	select {
	case <-flushed:
		// ok
	case <-time.After(2 * time.Second):
		t.Error("auto-flush did not trigger on batch size")
	}

	if got := e.BatchLen(); got != 0 {
		t.Errorf("BatchLen after auto-flush = %d, want 0", got)
	}
}

func TestFlushHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	e := NewExporter(
		WithEndpoint(srv.URL),
		WithFlushInterval(time.Hour),
	)
	defer func() { _ = e.Shutdown() }()

	_, span := e.StartSpan(context.Background(), "test")
	e.EndSpan(span)

	err := e.Flush()
	if err == nil {
		t.Error("expected error on HTTP 500")
	}
}

func TestShutdownFlushes(t *testing.T) {
	var mu sync.Mutex
	var received []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	e := NewExporter(
		WithEndpoint(srv.URL),
		WithFlushInterval(time.Hour),
	)

	_, span := e.StartSpan(context.Background(), "test")
	e.EndSpan(span)

	if err := e.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(received) == 0 {
		t.Error("Shutdown should have flushed remaining spans")
	}
}

func TestShutdownIdempotent(t *testing.T) {
	e := NewExporter(WithFlushInterval(time.Hour))

	if err := e.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if err := e.Shutdown(); err != nil {
		t.Fatal("double Shutdown should not error")
	}
}

func TestConvertSpan(t *testing.T) {
	s := &trace.Span{
		ID:        "test-id",
		ParentID:  "parent-id",
		Name:      "test.span",
		StartTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC),
		Attributes: map[string]string{
			"key1": "val1",
		},
		Status: trace.StatusError,
		Error:  "something failed",
	}

	os := convertSpan(s, "abc123")

	if os.TraceID != "abc123" {
		t.Errorf("TraceID = %q, want abc123", os.TraceID)
	}
	if os.Name != "test.span" {
		t.Errorf("Name = %q", os.Name)
	}
	if os.ParentSpanID == "" {
		t.Error("ParentSpanID should be set")
	}
	if os.Kind != 1 {
		t.Errorf("Kind = %d, want 1", os.Kind)
	}
	if os.Status.Code != 2 {
		t.Errorf("Status.Code = %d, want 2 (error)", os.Status.Code)
	}

	// Check gogrid-prefixed attribute.
	foundAttr := false
	foundErr := false
	for _, attr := range os.Attributes {
		if attr.Key == "gogrid.key1" && attr.Value.StringValue == "val1" {
			foundAttr = true
		}
		if attr.Key == "exception.message" && attr.Value.StringValue == "something failed" {
			foundErr = true
		}
	}
	if !foundAttr {
		t.Error("gogrid.key1 attribute not found")
	}
	if !foundErr {
		t.Error("exception.message attribute not found")
	}
}

func TestConvertSpanOKStatus(t *testing.T) {
	s := &trace.Span{
		ID:        "test-id",
		Name:      "test.span",
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Status:    trace.StatusOK,
	}

	os := convertSpan(s, "tid")
	if os.Status.Code != 1 {
		t.Errorf("Status.Code = %d, want 1 (ok)", os.Status.Code)
	}
	if os.ParentSpanID != "" {
		t.Errorf("ParentSpanID = %q, want empty", os.ParentSpanID)
	}
}

func TestDeriveID(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		hexLen int
	}{
		{"trace ID length", "test-span-id", 32},
		{"span ID length", "test-span-id", 16},
		{"short", "x", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveID(tt.src, tt.hexLen)
			if len(got) != tt.hexLen {
				t.Errorf("deriveID(%q, %d) len = %d, want %d", tt.src, tt.hexLen, len(got), tt.hexLen)
			}
		})
	}
}

func TestDeriveIDDeterministic(t *testing.T) {
	a := deriveID("same-input", 16)
	b := deriveID("same-input", 16)
	if a != b {
		t.Errorf("deriveID not deterministic: %q != %q", a, b)
	}
}

func TestDeriveIDUnique(t *testing.T) {
	a := deriveID("input-a", 16)
	b := deriveID("input-b", 16)
	if a == b {
		t.Error("deriveID should produce different IDs for different inputs")
	}
}

func TestExporterConcurrency(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	e := NewExporter(
		WithEndpoint(srv.URL),
		WithBatchSize(1000),
		WithFlushInterval(time.Hour),
	)
	defer func() { _ = e.Shutdown() }()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, span := e.StartSpan(context.Background(), "concurrent.span")
			_ = ctx
			e.EndSpan(span)
		}()
	}
	wg.Wait()

	if got := e.BatchLen(); got != 50 {
		t.Errorf("BatchLen = %d, want 50", got)
	}
}
