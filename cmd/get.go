package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <KEY>",
	Short: "Get a secret value from the current project bundle",
	Long: `Get a secret (environment variable) from the current project's encrypted bundle.
Prints the decrypted value to stdout.

Example:
  envmoat get API_KEY
  envmoat get API_KEY --json`,
	RunE: runGet,
}

var getJSON bool

func init() {
	getCmd.Flags().BoolVar(&getJSON, "json", false, "Output as JSON {\"key\": \"...\", \"value\": \"...\"}")
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmdutil.Errorf("provide a KEY argument", "usage: envmoat get <KEY>")
		return fmt.Errorf("no key specified")
	}

	key := args[0]

	bundle, err := resolveBundle()
	if err != nil {
		cmdutil.Errorf("", "%v", err)
		return err
	}

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

	value, ok := secrets[key]
	if !ok {
		cmdutil.Errorf("", "key %q not found in bundle", key)
		return fmt.Errorf("key %q not found", key)
	}

	if getJSON {
		output := struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{key, value}
		data, err := json.Marshal(output)
		if err != nil {
			cmdutil.Errorf("", "marshal JSON: %v", err)
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Println(value)
	}
	return nil
}
