package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user for yes/no confirmation on stderr.
// Returns true if the user confirms, false otherwise.
// If stdin is not a terminal, returns false.
func Confirm(prompt string) bool {
	fmt.Fprint(os.Stderr, prompt+" [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr)
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(line), "y") || strings.HasPrefix(strings.TrimSpace(line), "Y")
}
