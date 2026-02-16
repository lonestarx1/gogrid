package trace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestSpanSetAttribute(t *testing.T) {
	s := &Span{}
	s.SetAttribute("key", "value")

	if s.Attributes["key"] != "value" {
		t.Errorf("Attributes[key] = %q, want %q", s.Attributes["key"], "value")
	}
}

func TestSpanSetError(t *testing.T) {
	s := &Span{}
	s.SetError(errors.New("something failed"))

	if s.Status != StatusError {
		t.Errorf("Status = %d, want %d", s.Status, StatusError)
	}
	if s.Error != "something failed" {
		t.Errorf("Error = %q, want %q", s.Error, "something failed")
	}
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()

	// No span in empty context.
	if got := SpanFromContext(ctx); got != nil {
		t.Errorf("SpanFromContext(empty) = %v, want nil", got)
	}

	// NewSpan inserts a span into the context.
	ctx, span := NewSpan(ctx, "test")
	if span.Name != "test" {
		t.Errorf("Name = %q, want %q", span.Name, "test")
	}
	if span.ID == "" {
		t.Error("ID is empty")
	}
	if span.StartTime.IsZero() {
		t.Error("StartTime is zero")
	}

	got := SpanFromContext(ctx)
	if got != span {
		t.Error("SpanFromContext did not return the expected span")
	}
}

func TestSpanParentLinking(t *testing.T) {
	ctx := context.Background()

	ctx, parent := NewSpan(ctx, "parent")
	_, child := NewSpan(ctx, "child")

	if child.ParentID != parent.ID {
		t.Errorf("child.ParentID = %q, want %q", child.ParentID, parent.ID)
	}
}

func TestNoopTracer(t *testing.T) {
	tracer := Noop{}
	ctx, span := tracer.StartSpan(context.Background(), "test")

	// Should not panic and should return valid objects.
	if ctx == nil {
		t.Error("ctx is nil")
	}
	if span == nil {
		t.Error("span is nil")
	}

	// EndSpan should not panic.
	tracer.EndSpan(span)
}

func TestInMemoryTracer(t *testing.T) {
	tracer := NewInMemory()

	ctx, span1 := tracer.StartSpan(context.Background(), "span1")
	tracer.EndSpan(span1)

	_, span2 := tracer.StartSpan(ctx, "span2")
	tracer.EndSpan(span2)

	spans := tracer.Spans()
	if len(spans) != 2 {
		t.Fatalf("Spans len = %d, want 2", len(spans))
	}
	if spans[0].Name != "span1" {
		t.Errorf("spans[0].Name = %q, want %q", spans[0].Name, "span1")
	}
	if spans[1].Name != "span2" {
		t.Errorf("spans[1].Name = %q, want %q", spans[1].Name, "span2")
	}
	if spans[1].ParentID != spans[0].ID {
		t.Errorf("span2.ParentID = %q, want %q", spans[1].ParentID, spans[0].ID)
	}
	if spans[0].EndTime.IsZero() {
		t.Error("span1 EndTime is zero after EndSpan")
	}
}

func TestInMemorySpansSliceCopy(t *testing.T) {
	tracer := NewInMemory()
	_, span := tracer.StartSpan(context.Background(), "test")
	tracer.EndSpan(span)

	spans1 := tracer.Spans()
	// Appending to the returned slice should not affect the tracer's internal slice.
	_ = append(spans1, &Span{Name: "extra"})

	spans2 := tracer.Spans()
	if len(spans2) != 1 {
		t.Errorf("Spans len = %d after append to copy, want 1", len(spans2))
	}
}

func TestStdoutTracer(t *testing.T) {
	var buf bytes.Buffer
	tracer := NewStdout(&buf)

	_, span := tracer.StartSpan(context.Background(), "test-op")
	span.SetAttribute("key", "val")
	tracer.EndSpan(span)

	// Should have written JSON.
	if buf.Len() == 0 {
		t.Fatal("Stdout tracer wrote nothing")
	}

	var decoded Span
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Name != "test-op" {
		t.Errorf("Name = %q, want %q", decoded.Name, "test-op")
	}
	if decoded.Attributes["key"] != "val" {
		t.Errorf("Attributes[key] = %q, want %q", decoded.Attributes["key"], "val")
	}
	if decoded.EndTime.IsZero() {
		t.Error("EndTime is zero")
	}
}
