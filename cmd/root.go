package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// forceReauth forces re-authentication by bypassing the session cache.
var forceReauth bool

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
	rootCmd.Flags().BoolVar(&forceReauth, "force-reauth", false,
		"Always prompt for master password, bypassing session cache")
}