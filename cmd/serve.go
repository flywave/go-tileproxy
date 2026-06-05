package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/setting"
	"github.com/spf13/cobra"
)

const (
	maxConfigSize    = 10 * 1024 * 1024
	defaultShutdown  = 30 * time.Second
	configReloadFile = "config.json"
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
	Long:  `Start the tile proxy server using a configuration file. Supports JSON format.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runServer(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	},
}

type multiServiceConfig struct {
	Services []*setting.ProxyService `json:"services"`
}

func runServer() error {
	proxy := loadAndBuildProxy(configFile)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: proxy,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("starting tileproxy server on %s", srv.Addr)
	log.Printf("services: %v", proxy.GetServiceIDs())
	log.Printf("send SIGHUP to reload config, SIGINT/SIGTERM to shutdown")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	for sig := range sigCh {
		switch sig {
		case syscall.SIGHUP:
			log.Println("received SIGHUP, reloading configuration...")
			newProxy := loadAndBuildProxy(configFile)
			proxy.Shutdown(context.Background())
			proxy = newProxy
			srv.Handler = proxy
			log.Printf("reloaded %d services: %v", len(proxy.GetServiceIDs()), proxy.GetServiceIDs())

		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("shutting down...")
			signal.Stop(sigCh)

			ctx, cancel := context.WithTimeout(context.Background(), defaultShutdown)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("HTTP server shutdown error: %v", err)
			}
			if err := proxy.Shutdown(ctx); err != nil {
				log.Printf("service shutdown error: %v", err)
			}
			log.Println("server stopped")
			return nil
		}
	}

	return nil
}

func loadAndBuildProxy(path string) *tileproxy.TileProxy {
	proxyServices, err := loadConfig(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	proxy := tileproxy.NewTileProxy(&setting.GlobalsSetting{}, proxyServices, nil)
	return proxy
}

func loadConfig(path string) ([]*setting.ProxyService, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	if info.Size() > maxConfigSize {
		return nil, errors.New("config file too large (max 10MB)")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if isJSON(data) {
		return parseConfigJSON(data)
	}

	return nil, fmt.Errorf("unsupported config format, expected JSON")
}

func parseConfigJSON(data []byte) ([]*setting.ProxyService, error) {
	var multi multiServiceConfig
	if err := json.Unmarshal(data, &multi); err == nil && len(multi.Services) > 0 {
		return multi.Services, nil
	}

	var single setting.ProxyService
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&single); err == nil {
		single.Service = setting.CovertService(single.Service)
		return []*setting.ProxyService{&single}, nil
	}

	return nil, errors.New("invalid config: expected {\"services\":[...]} or single service object")
}

func isJSON(data []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(data, &js) == nil
}
