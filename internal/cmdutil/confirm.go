package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Confirm prompts the user for yes/no confirmation on stderr.
// Returns true if the user confirms, false otherwise.
// Always returns false unless ENVMOAT_INTERACTIVE=1 is set (for testing/CI safety).
func Confirm(prompt string) bool {
	fmt.Fprint(os.Stderr, prompt+" [y/N]: ")

	// Never prompt in non-interactive mode (default for all environments except explicit interactive).
	if os.Getenv("ENVMOAT_INTERACTIVE") != "1" {
		fmt.Fprintln(os.Stderr)
		return false
	}
	// In interactive mode, check if stdin is a terminal.
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr)
		return false
	}

	// Read user input with a short timeout to avoid blocking indefinitely.
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}
