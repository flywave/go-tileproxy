package main

import (
	"context"
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
	defaultShutdown = 30 * time.Second
)

var (
	configFile  string
	port        int
	host        string
	logLevel    string
	watchConfig bool
)

func init() {
	serveCmd.Flags().StringVarP(&configFile, "config", "c", "config.json", "Path to configuration file (JSON)")
	serveCmd.Flags().IntVarP(&port, "port", "p", 8000, "Server port")
	serveCmd.Flags().StringVarP(&host, "host", "H", "0.0.0.0", "Server host")
	serveCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	serveCmd.Flags().BoolVar(&watchConfig, "watch", true, "Watch the config file and hot-reload on change")

	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start tile proxy server",
	Long:  `Start the tile proxy server using a configuration file. The configuration is hot-reloaded when the file changes or on SIGHUP; a broken configuration never takes the server down.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runServer(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	},
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
	log.Printf("config hot-reload enabled: send SIGHUP or edit %s", configFile)

	var watcher *tileproxy.ConfigWatcher
	if watchConfig {
		watcher = tileproxy.NewConfigWatcher(proxy, configFile, nil, 2*time.Second, 1*time.Second)
		watcher.Start()
		defer watcher.Stop()
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	for sig := range sigCh {
		switch sig {
		case syscall.SIGHUP:
			log.Println("received SIGHUP, reloading configuration...")
			// a broken configuration keeps the current services running
			if err := proxy.ReloadFromFile(configFile, nil); err != nil {
				log.Printf("reload failed: %v", err)
				continue
			}
			log.Printf("reloaded %d services: %v", len(proxy.GetServiceIDs()), proxy.GetServiceIDs())

		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("shutting down...")
			signal.Stop(sigCh)

			if watcher != nil {
				watcher.Stop()
			}

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
	// the initial configuration must be valid, otherwise there is nothing
	// sensible to serve
	services, err := setting.LoadProxyServicesFromFile(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return tileproxy.NewTileProxy(&setting.GlobalsSetting{}, services, nil)
}
