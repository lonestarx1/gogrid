package tool

import (
	"context"
	"encoding/json"
)

// Tool is the interface that all GoGrid tools must implement.
// Tools extend agent capabilities by providing access to external systems,
// computations, or data sources.
type Tool interface {
	// Name returns the tool's unique identifier.
	Name() string
	// Description returns a human-readable description of what the tool does.
	// This is sent to the LLM to help it decide when to use the tool.
	Description() string
	// Schema returns the JSON Schema describing the tool's input parameters.
	Schema() Schema
	// Execute runs the tool with the given JSON input and returns the output.
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

// Schema represents a JSON Schema for tool parameters.
// It covers the most common schema constructs used in tool definitions.
type Schema struct {
	// Type is the JSON Schema type (e.g. "object", "string", "number", "array", "boolean").
	Type string `json:"type"`
	// Description explains what this schema element represents.
	Description string `json:"description,omitempty"`
	// Properties maps property names to their schemas. Used when Type is "object".
	Properties map[string]*Schema `json:"properties,omitempty"`
	// Required lists property names that must be present. Used when Type is "object".
	Required []string `json:"required,omitempty"`
	// Items describes array element schemas. Used when Type is "array".
	Items *Schema `json:"items,omitempty"`
	// Enum restricts the value to a fixed set.
	Enum []string `json:"enum,omitempty"`
}

// MarshalJSON implements json.Marshaler for Schema.
// This allows Schema to be used directly where json.RawMessage is expected.
func (s Schema) MarshalJSON() ([]byte, error) {
	type schemaAlias Schema
	return json.Marshal(schemaAlias(s))
}

// ToRawJSON converts the Schema to a json.RawMessage.
func (s Schema) ToRawJSON() (json.RawMessage, error) {
	return json.Marshal(s)
}
