package cmdutil

import (
	"fmt"
	"os"
)

// Errorf prints an error message with an actionable recovery hint to stderr.
// The format string and args are printed first, followed by a hint on how to recover.
func Errorf(actionableHint string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "envmoat: error: %s\n", msg)
	if actionableHint != "" {
		fmt.Fprintf(os.Stderr, "envmoat: hint: %s\n", actionableHint)
	}
}