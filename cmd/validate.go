package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flywave/go-tileproxy/setting"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a proxy configuration file",
	Long: `Check a configuration file for missing or incorrect settings without
starting a server. The command prints a structured report (the same JSON a
validation endpoint would return) and exits non-zero when the configuration
is invalid.`,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read config: %v\n", err)
			os.Exit(2)
		}

		report := setting.ValidateConfigJSON(data)

		out, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode report: %v\n", err)
			os.Exit(2)
		}
		fmt.Println(string(out))

		if !report.Valid {
			os.Exit(1)
		}
	},
}

func init() {
	validateCmd.Flags().StringVarP(&configFile, "config", "c", "config.json", "Path to configuration file (JSON)")
	rootCmd.AddCommand(validateCmd)
}
