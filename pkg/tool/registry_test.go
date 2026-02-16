package tool

import (
	"context"
	"encoding/json"
	"testing"
)

// stubTool is a minimal Tool implementation for testing.
type stubTool struct {
	name string
	desc string
}

func (s *stubTool) Name() string                                                 { return s.name }
func (s *stubTool) Description() string                                          { return s.desc }
func (s *stubTool) Schema() Schema                                               { return Schema{Type: "object"} }
func (s *stubTool) Execute(_ context.Context, _ json.RawMessage) (string, error) { return "", nil }

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool1 := &stubTool{name: "search", desc: "search the web"}

	if err := r.Register(tool1); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := r.Get("search")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name() != "search" {
		t.Errorf("Name = %q, want %q", got.Name(), "search")
	}
}

func TestRegistryDuplicateRegistration(t *testing.T) {
	r := NewRegistry()
	tool1 := &stubTool{name: "search", desc: "v1"}
	tool2 := &stubTool{name: "search", desc: "v2"}

	if err := r.Register(tool1); err != nil {
		t.Fatalf("Register first: %v", err)
	}
	if err := r.Register(tool2); err == nil {
		t.Fatal("Register duplicate: expected error, got nil")
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("Get nonexistent: expected error, got nil")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	tools := []*stubTool{
		{name: "search", desc: "search"},
		{name: "calculate", desc: "calc"},
		{name: "read_file", desc: "read"},
	}
	for _, tool := range tools {
		if err := r.Register(tool); err != nil {
			t.Fatalf("Register %q: %v", tool.name, err)
		}
	}

	names := r.List()
	if len(names) != 3 {
		t.Fatalf("List len = %d, want 3", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, tool := range tools {
		if !nameSet[tool.name] {
			t.Errorf("List missing %q", tool.name)
		}
	}
}

func TestRegistryLen(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("Len = %d, want 0", r.Len())
	}

	_ = r.Register(&stubTool{name: "a"})
	_ = r.Register(&stubTool{name: "b"})

	if r.Len() != 2 {
		t.Errorf("Len = %d, want 2", r.Len())
	}
}

func TestSchemaToRawJSON(t *testing.T) {
	s := Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"query": {Type: "string", Description: "search query"},
		},
		Required: []string{"query"},
	}

	raw, err := s.ToRawJSON()
	if err != nil {
		t.Fatalf("ToRawJSON: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded["type"] != "object" {
		t.Errorf("type = %v, want %q", decoded["type"], "object")
	}
}
