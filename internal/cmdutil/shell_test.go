package cmdutil

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// --- ShellEscapeSingleQuote ---

func TestShellEscapeSingleQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain value",
			input: "sk-1234567890abcdef",
			want:  "'sk-1234567890abcdef'",
		},
		{
			name:  "empty value",
			input: "",
			want:  "''",
		},
		{
			name:  "value with spaces",
			input: "hello world",
			want:  "'hello world'",
		},
		{
			name:  "value with single quote",
			input: "it's",
			want:  "'it'\\''s'",
		},
		{
			name:  "value with multiple single quotes",
			input: "a'b'c",
			want:  "'a'\\''b'\\''c'",
		},
		{
			name:  "value with double quotes",
			input: `he said "hi"`,
			want:  `'he said "hi"'`,
		},
		{
			name:  "value with dollar sign",
			input: "$HOME",
			want:  `'$HOME'`,
		},
		{
			name:  "value with backticks",
			input: "`whoami`",
			want:  "'`whoami`'",
		},
		{
			name:  "value with semicolon",
			input: "cmd; rm -rf /",
			want:  "'cmd; rm -rf /'",
		},
		{
			name:  "value with command substitution",
			input: "$(cat /etc/passwd)",
			want:  "'$(cat /etc/passwd)'",
		},
		{
			name:  "value with newline",
			input: "line1\nline2",
			want:  "'line1\nline2'",
		},
		{
			name:  "value with backslash",
			input: "path\\to\\file",
			want:  "'path\\to\\file'",
		},
		{
			name:  "value with mixed special chars",
			input: `p@$$w"rd; with $pecial chars`,
			want:  `'p@$$w"rd; with $pecial chars'`,
		},

	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ShellEscapeSingleQuote(tc.input)
			if got != tc.want {
				t.Errorf("ShellEscapeSingleQuote(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// --- FormatExportLine ---

func TestFormatExportLine(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{
			name:  "simple key-value",
			key:   "API_KEY",
			value: "sk-123",
			want:  "export API_KEY='sk-123'",
		},
		{
			name:  "value with single quote",
			key:   "DB_PASS",
			value: "p@ss'wrd",
			want:  "export DB_PASS='p@ss'\\''wrd'",
		},
		{
			name:  "empty value",
			key:   "EMPTY",
			value: "",
			want:  "export EMPTY=''",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatExportLine(tc.key, tc.value)
			if got != tc.want {
				t.Errorf("FormatExportLine(%q, %q) = %q, want %q", tc.key, tc.value, got, tc.want)
			}
		})
	}
}

// --- EmitLoadOutput ---

func TestEmitLoadOutputBasic(t *testing.T) {
	secrets := map[string]string{
		"API_KEY": "sk-1234567890abcdef",
		"DB_PASS": "p@ssw0rd",
	}
	var stdout, stderr bytes.Buffer

	err := EmitLoadOutput(&stdout, &stderr, "abc123", secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Header line
	if !strings.HasPrefix(output, "#bundle_hash:sha256:abc123\n") {
		t.Errorf("expected bundle hash header, got: %q", output)
	}
	// Export lines (sorted: API_KEY before DB_PASS)
	if !strings.Contains(output, "export API_KEY='sk-1234567890abcdef'\n") {
		t.Errorf("missing API_KEY export, got: %q", output)
	}
	if !strings.Contains(output, "export DB_PASS='p@ssw0rd'\n") {
		t.Errorf("missing DB_PASS export, got: %q", output)
	}
	// No errors
	if stderr.Len() > 0 {
		t.Errorf("unexpected stderr: %q", stderr.String())
	}
}

func TestEmitLoadOutputNoHash(t *testing.T) {
	secrets := map[string]string{
		"KEY": "value",
	}
	var stdout, stderr bytes.Buffer

	err := EmitLoadOutput(&stdout, &stderr, "", secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "bundle_hash") {
		t.Errorf("expected no hash header, got: %q", output)
	}
	if !strings.Contains(output, "export KEY='value'\n") {
		t.Errorf("missing export line, got: %q", output)
	}
}

func TestEmitLoadOutputEmptySecrets(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := EmitLoadOutput(&stdout, &stderr, "hash123", map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should have header but no export lines
	if !strings.HasPrefix(output, "#bundle_hash:sha256:hash123\n") {
		t.Errorf("expected hash header, got: %q", output)
	}
	// No export lines
	if strings.Contains(output, "export") {
		t.Errorf("expected no export lines, got: %q", output)
	}
}

func TestEmitLoadOutputWithSpecialChars(t *testing.T) {
	secrets := map[string]string{
		"COMPLEX": "value with 'quotes' and $dollars and `backticks`",
	}
	var stdout, stderr bytes.Buffer

	err := EmitLoadOutput(&stdout, &stderr, "", secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Verify the line starts and ends correctly
	if !strings.HasPrefix(output, "export COMPLEX=") {
		t.Errorf("expected export COMPLEX= prefix, got: %q", output)
	}
	// Single quotes escaped as '\''
	if !strings.Contains(output, "'\\''quotes'\\''") {
		t.Errorf("expected escaped single quotes, got: %q", output)
	}
	// Dollar signs and backticks preserved inside single quotes
	if !strings.Contains(output, "$dollars") {
		t.Errorf("expected $dollars preserved, got: %q", output)
	}
	if !strings.Contains(output, "`backticks`") {
		t.Errorf("expected backticks preserved, got: %q", output)
	}
}

func TestEmitLoadOutputDeterministicOrder(t *testing.T) {
	// Insert keys in random order; output should be sorted.
	secrets := map[string]string{
		"ZEBRA": "z",
		"ALPHA": "a",
		"MIDDLE": "m",
	}
	var stdout, stderr bytes.Buffer

	err := EmitLoadOutput(&stdout, &stderr, "", secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "export ALPHA='a'" {
		t.Errorf("first line = %q, want export ALPHA='a'", lines[0])
	}
	if lines[1] != "export MIDDLE='m'" {
		t.Errorf("second line = %q, want export MIDDLE='m'", lines[1])
	}
	if lines[2] != "export ZEBRA='z'" {
		t.Errorf("third line = %q, want export ZEBRA='z'", lines[2])
	}
}

func TestEmitLoadOutputErrorsGoToStderr(t *testing.T) {
	// Use a broken writer to force an error.
	brokenWriter := &brokenWriter{}
	var stderr bytes.Buffer

	err := EmitLoadOutput(brokenWriter, &stderr, "hash", map[string]string{"KEY": "val"})
	if err == nil {
		t.Fatal("expected error from broken writer, got nil")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "envmoat: error:") {
		t.Errorf("expected error on stderr, got: %q", stderrStr)
	}
}

// brokenWriter always fails on Write.
type brokenWriter struct{}

func (b *brokenWriter) Write(p []byte) (int, error) {
	return 0, errors.New("broken writer")
}
