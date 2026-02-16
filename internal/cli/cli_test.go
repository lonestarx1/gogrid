package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.Run(nil)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Error("expected usage message in stdout")
	}
}

func TestRun_Help(t *testing.T) {
	for _, arg := range []string{"help", "-h", "--help"} {
		t.Run(arg, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			app := New(&stdout, &stderr)

			code := app.Run([]string{arg})
			if code != 0 {
				t.Errorf("exit code = %d, want 0", code)
			}
			if !strings.Contains(stdout.String(), "Commands:") {
				t.Error("expected commands list in output")
			}
		})
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.Run([]string{"bogus"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Error("expected unknown command error in stderr")
	}
}

func TestRun_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.Run([]string{"version"})
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "gogrid") {
		t.Error("expected 'gogrid' in version output")
	}
}
