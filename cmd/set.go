package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/spf13/cobra"
)

const (
	// SecretSizeWarning is the size threshold (in bytes) at which a warning is shown.
	SecretSizeWarning = 1 << 20 // 1 MB
)

// keyRegex matches valid secret keys: alphanumeric, underscore, hyphen.
var keyRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var setCmd = &cobra.Command{
	Use:   "set <KEY> [VALUE]",
	Short: "Set a secret in the current project bundle",
	Long: `Set a secret (environment variable) in the current project's encrypted bundle.

If VALUE is omitted, you will be prompted to enter it interactively.
Use --stdin to read the value from standard input.
Use --file to bulk import from a .env file.

Examples:
  envmoat set API_KEY sk-1234567890abcdef
  envmoat set DB_PASS
  envmoat set API_KEY --stdin <<< "sk-123"
  envmoat set --file .env`,
	RunE: runSet,
}

var setStdin bool
var setFile string

func init() {
	setCmd.Flags().BoolVarP(&setStdin, "stdin", "s", false, "Read value from standard input")
	setCmd.Flags().StringVarP(&setFile, "file", "f", "", "Bulk import from .env file (KEY=VALUE per line)")
	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	if setFile == "" && len(args) == 0 {
		cmdutil.Errorf("provide a KEY argument or use --file", "usage: envmoat set <KEY> [VALUE]")
		return fmt.Errorf("no key specified")
	}

	bundle, err := resolveBundle()
	if err != nil {
		cmdutil.Errorf("", "%v", err)
		return err
	}

	// Load existing secrets.
	plaintext, err := bundle.Store.ReadBundle(bundle.BundleFile, bundle.DEK)
	if err != nil {
		cmdutil.Errorf("", "read bundle: %v", err)
		return err
	}

	secrets := make(map[string]string)
	if len(plaintext) > 0 {
		if err := json.Unmarshal(plaintext, &secrets); err != nil {
			cmdutil.Errorf("", "parse bundle: %v", err)
			return err
		}
	}
	if secrets == nil {
		secrets = make(map[string]string)
	}

	if setFile != "" {
		// Bulk import from .env file.
		pairs, err := parseEnvFile(setFile)
		if err != nil {
			cmdutil.Errorf("", "parse file %s: %v", setFile, err)
			return err
		}
		for k, v := range pairs {
			if !keyRegex.MatchString(k) {
				cmdutil.Errorf("use alphanumeric, underscore, and hyphen only", "invalid key: %q", k)
				return fmt.Errorf("invalid key: %q", k)
			}
			if len(v) > SecretSizeWarning {
				fmt.Fprintf(os.Stderr, "envmoat: warning: value for %q is %d bytes (near practical limit)\n", k, len(v))
			}
			secrets[k] = v
		}
		newPlaintext, _ := json.Marshal(secrets)
		if err := bundle.Store.WriteBundle(bundle.BundleFile, newPlaintext, bundle.DEK); err != nil {
			cmdutil.Errorf("", "write bundle: %v", err)
			return err
		}
		fmt.Println(fmt.Sprintf("Imported %d secrets from %s", len(pairs), setFile))
		return nil
	}

	// Single key/value mode.
	key := args[0]
	if !keyRegex.MatchString(key) {
		cmdutil.Errorf("use alphanumeric, underscore, and hyphen only", "invalid key: %q", key)
		return fmt.Errorf("invalid key: %q", key)
	}

	var value string
	if setStdin {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			value = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			cmdutil.Errorf("", "read stdin: %v", err)
			return err
		}
	} else if len(args) > 1 {
		value = args[1]
	} else {
		// Interactive prompt.
		fmt.Fprint(os.Stderr, "Enter value for "+key+": ")
		password, err := readPassword()
		if err != nil {
			cmdutil.Errorf("", "read password: %v", err)
			return err
		}
		fmt.Fprintln(os.Stderr)
		value = password
	}

	if len(value) > SecretSizeWarning {
		fmt.Fprintf(os.Stderr, "envmoat: warning: value for %q is %d bytes (near practical limit)\n", key, len(value))
	}

	secrets[key] = value
	newPlaintext, _ := json.Marshal(secrets)
	if err := bundle.Store.WriteBundle(bundle.BundleFile, newPlaintext, bundle.DEK); err != nil {
		cmdutil.Errorf("", "write bundle: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("Set %s", key))
	return nil
}

// parseEnvFile parses a .env file into key-value pairs.
// Supports KEY=VALUE format, skips # comments and blank lines, handles quoted values.
func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		// Remove surrounding quotes.
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		result[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
