package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret keys in the current project bundle",
	Long: `List all secret keys in the current project's encrypted bundle.
Values are not shown.

Example:
  envmoat list
  envmoat list --json`,
	RunE: runList,
}

var listJSON bool

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON {\"keys\": [\"KEY1\", \"KEY2\"]}")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	bundle, err := resolveBundle()
	if err != nil {
		cmdutil.Errorf("", "%v", err)
		return err
	}

	// Show active profile/bundle name.
	if bundle.ProfileName != "" {
		fmt.Fprintf(os.Stderr, "envmoat: profile: %s\n", bundle.ProfileName)
	} else {
		fmt.Fprintf(os.Stderr, "envmoat: bundle: %s\n", bundle.BundleFile)
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

	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if listJSON {
		output := struct {
			Keys []string `json:"keys"`
		}{keys}
		data, err := json.Marshal(output)
		if err != nil {
			cmdutil.Errorf("", "marshal JSON: %v", err)
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		for _, k := range keys {
			fmt.Println(k)
		}
	}
	return nil
}
