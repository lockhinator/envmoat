package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "envmoat",
	Short: "Keep your secrets out of AI agent context",
	Long: `envmoat is a macOS+Linux secrets manager that keeps encrypted environment
variables invisible to AI coding agents.

Secrets are stored encrypted outside your project directories and auto-injected
into your shell session when you cd into a tracked project.

Quick start:
  envmoat setup    # create master password + install shell hook
  cd ~/projects/myapp
  envmoat init     # create marker + bundle
  envmoat set API_KEY  # prompts for value
`,
	Run: func(cmd *cobra.Command, args []string) {
		// No subcommand: print welcome + usage
		fmt.Println("Welcome to envmoat!")
		fmt.Println()
		fmt.Println("Keep your secrets out of AI agent context.")
		fmt.Println()
		fmt.Println("Quick start:")
		fmt.Println("  envmoat setup    # create master password + install shell hook")
		fmt.Println("  cd ~/projects/myapp")
		fmt.Println("  envmoat init     # create marker + bundle")
		fmt.Println("  envmoat set API_KEY  # prompts for value")
		fmt.Println()
		fmt.Println("Run 'envmoat --help' for all commands.")
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("envmoat version {{.Version}}\n")

	// Add subcommand stubs
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(deinitCmd)
	rootCmd.AddCommand(verifyCmd)
}

// setupCmd — create master password + install shell hook
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create master password and install shell hook",
	Long: `Create a master password for encrypting secrets and install the shell hook
into your rc file (~/.zshrc or ~/.bashrc).

Run this once after installation. Use --reset to change your master password.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// initCmd — create marker + auto-named bundle
var initCmd = &cobra.Command{
	Use:   "init [project-root]",
	Short: "Initialize envmoat in a project directory",
	Long: `Create a .envmoat marker file and auto-named secret bundle for the project.
Defaults to current directory. Auto-adds .envmoat to .gitignore.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// setCmd — add/update secret
var setCmd = &cobra.Command{
	Use:   "set KEY [VALUE]",
	Short: "Add or update a secret",
	Long: `Add or update a secret in the current project's bundle.
If VALUE is omitted, you will be prompted to enter it interactively.

Examples:
  envmoat set API_KEY sk-1234...
  envmoat set DB_PASS     # prompts for value
  echo "secret" | envmoat set --stdin API_KEY
  envmoat set --file .env # bulk import`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// getCmd — print decrypted value
var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Print a secret's value to stdout",
	Long: `Print the decrypted value of a secret to stdout.
Use --clip to copy to clipboard instead of printing.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// listCmd — list keys (values hidden)
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret keys (values hidden)",
	Long: `List all secret keys in the current project's bundle.
Secret values are never shown. Shows active profile name for context.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// loadCmd — emit shell-safe export lines
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Emit shell-safe export lines (for shell hook)",
	Long: `Resolve the current directory's marker, decrypt the bundle, and emit
shell-safe "export KEY='VALUE'" lines to stdout.

This command is called by the shell hook on cd. Errors go to stderr.
Exits 0 with no output when no bundle is found.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// removeCmd — delete a secret
var removeCmd = &cobra.Command{
	Use:   "remove KEY",
	Short: "Delete a secret",
	Long: `Delete a secret from the current project's bundle.
Prompts for confirmation unless -y/--yes is provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// deinitCmd — remove marker and bundle
var deinitCmd = &cobra.Command{
	Use:   "deinit [project-root]",
	Short: "Remove envmoat marker and bundle from a project",
	Long: `Remove the .envmoat marker file and associated secret bundle from a project.
Prompts for confirmation unless -y/--yes is provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

// verifyCmd — integrity check
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify bundle integrity and consistency",
	Long: `Check that all bundles can be decrypted and that the index is consistent.
Lists orphaned bundles and prompts for deletion.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}