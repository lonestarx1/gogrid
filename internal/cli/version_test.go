package cli

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestRunVersion_Format(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.runVersion()
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	out := stdout.String()

	// Should contain version string.
	if !strings.Contains(out, "gogrid") {
		t.Errorf("output %q missing 'gogrid'", out)
	}

	// Should contain OS/arch.
	if !strings.Contains(out, runtime.GOOS+"/"+runtime.GOARCH) {
		t.Errorf("output %q missing OS/ARCH", out)
	}

	// Should contain Go version.
	if !strings.Contains(out, runtime.Version()) {
		t.Errorf("output %q missing Go version", out)
	}
}

func TestRunVersion_DevDefault(t *testing.T) {
	var stdout bytes.Buffer
	app := New(&stdout, &bytes.Buffer{})

	app.runVersion()

	if !strings.Contains(stdout.String(), "dev") {
		t.Error("expected default version 'dev'")
	}
}
