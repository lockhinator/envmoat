package cmdutil

import (
	"fmt"
	"os"
)

// Debug logs a debug message to stderr only when ENVMOAT_DEBUG is set to a non-empty value.
// Never logs secret values.
func Debug(format string, args ...any) {
	if os.Getenv("ENVMOAT_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "envmoat: debug: "+format+"\n", args...)
	}
}