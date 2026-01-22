package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/setting"
	"github.com/spf13/cobra"
)

var (
	configFile string
	port       int
	host       string
	logLevel   string
)

func init() {
	serveCmd.Flags().StringVarP(&configFile, "config", "c", "config.json", "Path to configuration file (JSON or YAML)")
	serveCmd.Flags().IntVarP(&port, "port", "p", 8000, "Server port")
	serveCmd.Flags().StringVarP(&host, "host", "H", "0.0.0.0", "Server host")
	serveCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")

	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start tile proxy server",
	Long:  `Start the tile proxy server using a configuration file. Supports JSON and YAML formats.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runServer(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runServer() error {
	proxyService, err := loadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	service := tileproxy.NewService(proxyService, &setting.GlobalsSetting{}, nil)

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting tileproxy server on %s\n", addr)
	fmt.Printf("Configuration loaded from: %s\n", configFile)
	fmt.Printf("Health check: http://%s/health\n", addr)

	err = http.ListenAndServe(addr, service.Service)
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func loadConfig(path string) (*setting.ProxyService, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var proxyService *setting.ProxyService

	if isJSON(data) {
		proxyService, err = setting.CreateProxyServiceFromJSON(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported config format, expected JSON or YAML")
	}

	if proxyService == nil {
		return nil, fmt.Errorf("empty configuration")
	}

	return proxyService, nil
}

func isJSON(data []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(data, &js) == nil
}
