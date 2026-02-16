package config

import (
	"testing"
)

func TestSubstitute(t *testing.T) {
	tests := []struct {
		name  string
		input string
		env   map[string]string
		want  string
	}{
		{
			name:  "no patterns",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "simple var set",
			input: "key: ${MY_VAR}",
			env:   map[string]string{"MY_VAR": "value"},
			want:  "key: value",
		},
		{
			name:  "simple var unset",
			input: "key: ${UNSET_VAR}",
			want:  "key: ",
		},
		{
			name:  "default when unset",
			input: "key: ${UNSET_VAR:-fallback}",
			want:  "key: fallback",
		},
		{
			name:  "default not used when set",
			input: "key: ${MY_VAR:-fallback}",
			env:   map[string]string{"MY_VAR": "actual"},
			want:  "key: actual",
		},
		{
			name:  "multiple patterns",
			input: "${A} and ${B:-two}",
			env:   map[string]string{"A": "one"},
			want:  "one and two",
		},
		{
			name:  "adjacent patterns",
			input: "${X}${Y}",
			env:   map[string]string{"X": "a", "Y": "b"},
			want:  "ab",
		},
		{
			name:  "unclosed brace",
			input: "key: ${BROKEN",
			want:  "key: ${BROKEN",
		},
		{
			name:  "empty default",
			input: "${VAR:-}",
			want:  "",
		},
		{
			name:  "default with special chars",
			input: "${VAR:-http://localhost:8080}",
			want:  "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars for this test.
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got := Substitute(tt.input)
			if got != tt.want {
				t.Errorf("Substitute(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
