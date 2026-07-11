// Package cmdutil provides CLI utility functions for envmoat.
package cmdutil

import (
	"fmt"
	"io"
	"strings"
)

// ShellEscapeSingleQuote escapes a value for safe use inside single quotes.
// Single quotes inside the value are escaped as '\'' (close-quote, escaped-quote, open-quote).
// All other characters are safe inside single quotes — no expansion, no interpretation.
func ShellEscapeSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// FormatExportLine produces a shell-safe "export KEY='VALUE'" line.
func FormatExportLine(key, value string) string {
	return fmt.Sprintf("export %s=%s", key, ShellEscapeSingleQuote(value))
}

// EmitLoadOutput writes shell-safe export lines to w, one per key-value pair.
// If bundleHash is non-empty, a comment header "#bundle_hash:sha256:<hash>" is written first.
// Errors are written to errW (typically os.Stderr). Returns nil on success.
func EmitLoadOutput(w io.Writer, errW io.Writer, bundleHash string, secrets map[string]string) error {
	if bundleHash != "" {
		if _, err := fmt.Fprintf(w, "#bundle_hash:sha256:%s\n", bundleHash); err != nil {
			fmt.Fprintf(errW, "envmoat: error: failed to write output: %v\n", err)
			return err
		}
	}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sortStrings(keys)

	for _, k := range keys {
		line := FormatExportLine(k, secrets[k])
		if _, err := fmt.Fprintln(w, line); err != nil {
			fmt.Fprintf(errW, "envmoat: error: failed to write output: %v\n", err)
			return err
		}
	}

	return nil
}

// sortStrings sorts a string slice in place (simple insertion sort, fine for small N).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
