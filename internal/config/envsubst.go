package config

import (
	"os"
	"strings"
)

// Substitute replaces ${VAR} and ${VAR:-default} patterns in s with
// environment variable values. If a variable is unset or empty and no
// default is provided, the pattern is replaced with an empty string.
func Substitute(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	i := 0
	for i < len(s) {
		// Look for "${".
		idx := strings.Index(s[i:], "${")
		if idx < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i : i+idx])
		i += idx + 2 // skip past "${"

		// Find closing "}".
		end := strings.IndexByte(s[i:], '}')
		if end < 0 {
			// No closing brace — write the literal "${" and continue.
			b.WriteString("${")
			continue
		}

		expr := s[i : i+end]
		i += end + 1 // skip past "}"

		// Check for ":-" default separator.
		if sep := strings.Index(expr, ":-"); sep >= 0 {
			name := expr[:sep]
			def := expr[sep+2:]
			if val := os.Getenv(name); val != "" {
				b.WriteString(val)
			} else {
				b.WriteString(def)
			}
		} else {
			b.WriteString(os.Getenv(expr))
		}
	}

	return b.String()
}
