package cmdutil

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr redirects os.Stderr to a pipe. Call close() to finalize and
// read the captured output. Do not read buf until close() returns.
func captureStderr(t *testing.T) (*strings.Builder, func()) {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	buf := new(strings.Builder)
	return buf, func() {
		w.Close() // signal EOF to reader
		io.Copy(buf, r) //nolint:errcheck
		r.Close()
		os.Stderr = old
	}
}

// --- Debug ---

func TestDebugNoOutputWhenEnvUnset(t *testing.T) {
	os.Unsetenv("ENVMOAT_DEBUG")
	buf, close := captureStderr(t)

	Debug("hello %s", "world")
	close()

	if buf.Len() > 0 {
		t.Fatalf("expected no output, got: %q", buf.String())
	}
}

func TestDebugOutputWhenEnvSet(t *testing.T) {
	os.Setenv("ENVMOAT_DEBUG", "1")
	defer os.Unsetenv("ENVMOAT_DEBUG")

	buf, close := captureStderr(t)

	Debug("hello %s", "world")
	close()

	got := buf.String()
	if !strings.Contains(got, "envmoat: debug: hello world") {
		t.Fatalf("expected debug prefix and message, got: %q", got)
	}
}

func TestDebugWithEmptyEnvValue(t *testing.T) {
	os.Setenv("ENVMOAT_DEBUG", "")
	defer os.Unsetenv("ENVMOAT_DEBUG")

	buf, close := captureStderr(t)

	Debug("should not appear")
	close()

	if buf.Len() > 0 {
		t.Fatalf("expected no output for empty ENVMOAT_DEBUG, got: %q", buf.String())
	}
}

// --- Errorf ---

func TestErrorfWithHint(t *testing.T) {
	buf, close := captureStderr(t)

	Errorf("try again", "something went wrong: %d", 42)
	close()

	got := buf.String()
	if !strings.Contains(got, "envmoat: error: something went wrong: 42") {
		t.Fatalf("expected error message, got: %q", got)
	}
	if !strings.Contains(got, "envmoat: hint: try again") {
		t.Fatalf("expected hint, got: %q", got)
	}
}

func TestErrorfWithoutHint(t *testing.T) {
	buf, close := captureStderr(t)

	Errorf("", "simple error")
	close()

	got := buf.String()
	if !strings.Contains(got, "envmoat: error: simple error") {
		t.Fatalf("expected error message, got: %q", got)
	}
	if strings.Contains(got, "hint:") {
		t.Fatalf("expected no hint line, got: %q", got)
	}
}

func TestErrorfFormatsArgs(t *testing.T) {
	buf, close := captureStderr(t)

	Errorf("check logs", "%s failed with code %d", "request", 500)
	close()

	got := buf.String()
	if !strings.Contains(got, "request failed with code 500") {
		t.Fatalf("expected formatted args in error, got: %q", got)
	}
}
