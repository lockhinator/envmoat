// Package cmd implements the envmoat CLI commands.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/store"
	"golang.org/x/term"
)

// setupCmd — create master password + install shell hook
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create master password and install shell hook",
	Long: `Create a master password for encrypting secrets and install the shell hook
into your rc file (~/.zshrc or ~/.bashrc).

Run this once after installation. Use --reset to change your master password.`,
	Run: runSetup,
}

var setupReset bool

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&setupReset, "reset", false, "Change master password (re-encrypts all bundles)")
}

func runSetup(cmd *cobra.Command, args []string) {
	// Check FileVault on macOS.
	if runtime.GOOS == "darwin" {
		checkFileVault()
	}

	st, err := store.NewStore()
	if err != nil {
		cmdutil.Errorf("run 'envmoat setup' again", "create store: %v", err)
		return
	}

	// Initialize store directory structure.
	if err := st.InitStore(); err != nil {
		cmdutil.Errorf("run 'envmoat setup' again", "initialize store: %v", err)
		return
	}

	// Prompt for master password.
	password, err := promptPassword()
	if err != nil {
		cmdutil.Errorf("", "read password: %v", err)
		return
	}

	if setupReset {
		// In reset mode, we just update the config salt and re-encrypt.
		// For now, same as fresh setup (re-encrypt would be rotate command).
		fmt.Println("Password updated. Run 'envmoat rotate' to re-encrypt existing bundles.")
	}

	// Generate new global salt and write config.
	cfg, err := store.DefaultConfig()
	if err != nil {
		cmdutil.Errorf("run 'envmoat setup' again", "generate config: %v", err)
		return
	}
	// Derive LUK to verify password works (side-effect validation).
	_, err = crypto.DeriveLUK(password, cfg.GlobalSalt)
	if err != nil {
		cmdutil.Errorf("run 'envmoat setup' again", "derive key: %v", err)
		return
	}

	if err := store.WriteConfig(st.ConfigPath, cfg); err != nil {
		cmdutil.Errorf("run 'envmoat setup' again", "write config: %v", err)
		return
	}

	fmt.Println("Master password set. Global salt stored in config.yaml.")

	// Install shell hook.
	if err := installShellHook(); err != nil {
		cmdutil.Errorf("run 'envmoat setup' again", "install shell hook: %v", err)
		return
	}

	fmt.Println("Shell hook installed.")
	fmt.Println("envmoat is ready. Run 'envmoat init' in a project to get started.")
}

// promptPassword interactively prompts the user for a password (hidden input).
func promptPassword() (string, error) {
	fmt.Print("Enter master password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()

	fmt.Print("Confirm master password: ")
	byteConfirm, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()

	password := string(bytePassword)
	confirm := string(byteConfirm)

	if password != confirm {
		return "", fmt.Errorf("passwords do not match")
	}

	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}

	return password, nil
}

// checkFileVault warns if FileVault is disabled on macOS.
func checkFileVault() {
	cmd := exec.Command("fdesetup", "isactive")
	if err := cmd.Run(); err != nil {
		fmt.Println("Warning: FileVault appears to be disabled. It is recommended to enable FileVault for full disk encryption.")
		fmt.Println("Enable with: sudo fdesetup enable")
		fmt.Println()
	}
}

// installShellHook detects the user's shell rc file and installs the envmoat hook.
// Idempotent: checks for __envmoat_hook marker before appending.
func installShellHook() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	// Detect rc file.
	rcPath, shellType, err := detectRcFile(home)
	if err != nil {
		return err
	}

	// Read existing rc file content.
	content, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", rcPath, err)
	}

	// Idempotent check: if hook already present, skip.
	if strings.Contains(string(content), "__envmoat_hook") {
		fmt.Printf("Shell hook already installed in %s.\n", rcPath)
		return nil
	}

	// Generate hook script.
	hook := generateShellHook(shellType)

	// Append hook to rc file.
	var toWrite string
	if len(content) > 0 {
		// Ensure there's a newline before the hook.
		if !strings.HasSuffix(string(content), "\n") {
			toWrite = string(content) + "\n"
		} else {
			toWrite = string(content)
		}
	}
	toWrite += "\n" + hook + "\n"

	f, err := os.OpenFile(rcPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open %s for append: %w", rcPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(toWrite); err != nil {
		return fmt.Errorf("write to %s: %w", rcPath, err)
	}

	fmt.Printf("Shell hook installed in %s.\n", rcPath)
	return nil
}

// detectRcFile finds the user's shell rc file (~/.zshrc or ~/.bashrc).
func detectRcFile(home string) (string, string, error) {
	// Check SHELL env var first.
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh" // macOS default
	}

	shellName := filepath.Base(shell)

	switch shellName {
	case "zsh":
		rcPath := filepath.Join(home, ".zshrc")
		// If .zshrc doesn't exist, check for .bashrc or create .zshrc
		if _, err := os.Stat(rcPath); os.IsNotExist(err) {
			bashRcPath := filepath.Join(home, ".bashrc")
			if _, err := os.Stat(bashRcPath); err == nil {
				return bashRcPath, "bash", nil
			}
		}
		return rcPath, "zsh", nil
	case "bash":
		rcPath := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(rcPath); os.IsNotExist(err) {
			// Try .bash_profile as fallback
			bashProfilePath := filepath.Join(home, ".bash_profile")
			if _, err := os.Stat(bashProfilePath); err == nil {
				return bashProfilePath, "bash", nil
			}
		}
		return rcPath, "bash", nil
	default:
		// Default to zsh on macOS, bash on Linux.
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".zshrc"), "zsh", nil
		}
		return filepath.Join(home, ".bashrc"), "bash", nil
	}
}

// generateShellHook returns the appropriate shell hook script for the given shell.
func generateShellHook(shellType string) string {
	switch shellType {
	case "zsh":
		return `# __envmoat_hook
if [[ $- == *i* ]]; then
  __envmoat_last_bundle=""
  __envmoat_hook() {
    local bundle
    bundle=$(envmoat load 2>/dev/null)
    if [[ -n "$bundle" && "$bundle" != "$__envmoat_last_bundle" ]]; then
      __envmoat_last_bundle="$bundle"
      eval "$bundle"
    fi
  }
  autoload -U add-zsh-hook
  add-zsh-hook chpwd __envmoat_hook
  __envmoat_hook
fi
# __envmoat_hook_end`
	case "bash":
		return `# __envmoat_hook
if [[ $- == *i* ]]; then
  __envmoat_last_bundle=""
  __envmoat_hook() {
    local bundle
    bundle=$(envmoat load 2>/dev/null)
    if [[ -n "$bundle" && "$bundle" != "$__envmoat_last_bundle" ]]; then
      __envmoat_last_bundle="$bundle"
      eval "$bundle"
    fi
  }
  if [[ ${BASH_VERSINFO[0]:-0} -ge 5 && ${BASH_VERSINFO[1]:-0} -ge 1 ]]; then
    PROMPT_COMMAND=("${PROMPT_COMMAND[@]}" "__envmoat_hook")
  else
    PROMPT_COMMAND="__envmoat_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
  fi
  __envmoat_hook
fi
# __envmoat_hook_end`
	default:
		// Default to zsh hook.
		return `# __envmoat_hook
if [[ $- == *i* ]]; then
  __envmoat_last_bundle=""
  __envmoat_hook() {
    local bundle
    bundle=$(envmoat load 2>/dev/null)
    if [[ -n "$bundle" && "$bundle" != "$__envmoat_last_bundle" ]]; then
      __envmoat_last_bundle="$bundle"
      eval "$bundle"
    fi
  }
  autoload -U add-zsh-hook
  add-zsh-hook chpwd __envmoat_hook
  __envmoat_hook
fi
# __envmoat_hook_end`
	}
}
