package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

// ============================================================================
// TMS DEMO TEMPLATE
// ============================================================================
// This template provides a basic structure for creating a new tile proxy demo
// that proxies TMS (Tile Map Service) or XYZ tile sources.
//
// Usage:
//   1. Copy this template to a new directory
//   2. Update the constants and configuration variables
//   3. Customize the service name, port, and endpoints
//   4. Build and run: go run main.go
//
// Example URL: http://127.0.0.1:8009/layer/0/0/0.png
// ============================================================================

const (
	// API_URL is the base URL for the tile service
	// Replace with your tile service URL
	API_URL = "https://example.com/tiles"

	// PORT is the port number for this proxy service
	PORT = ":8009"

	// SERVICE_NAME is the identifier for this proxy service
	SERVICE_NAME = "template"
)

var (
	// tileSource defines the upstream tile source configuration
	tileSource = setting.TileSource{
		// URLTemplate defines the tile URL pattern
		// Supported placeholders: {x}, {y}, {z}, {quadkey}, {tms_path}
		URLTemplate: API_URL + "/{z}/{x}/{y}.png",
		// Grid defines the tile grid system (use demo.GridMap keys)
		// Common values: "global_webmercator", "global_geodetic"
		Grid: "global_webmercator",
		// Subdomains enables server rotation for load balancing
		Subdomains: []string{"0", "1", "2", "3"},
		// Options specifies tile format and encoding options
		Options: &setting.ImageOpts{
			Format: "png",
		},
	}

	// tileCache defines the cache configuration for this service
	tileCache = setting.CacheSource{
		// Sources lists the upstream source names to cache
		Sources: []string{"tile_source"},
		// Name is the unique identifier for this cache
		Name: "template_cache",
		// Grid should match the source grid
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		// CacheInfo specifies cache storage configuration
		CacheInfo: &setting.CacheInfo{
			// Directory: where cached tiles are stored
			Directory: "./cache_data/template",
			// DirectoryLayout: "tms" (standard TMS layout)
			DirectoryLayout: "tms",
		},
		// TileOptions specifies output tile format
		TileOptions: &setting.ImageOpts{
			Format: "png",
		},
	}

	// tileService defines the service configuration
	tileService = setting.TMSService{
		// Layers defines available layers in this service
		Layers: []setting.TileLayer{
			{
				// Source references the cache name
				Source: "template_cache",
				// Name is the layer identifier used in URLs
				Name: "layer",
			},
		},
	}
)

// getProxyService creates and returns the proxy service configuration
// This function sets up all sources, caches, and the main service
func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService(SERVICE_NAME)

	// Add grids from demo package (global_webmercator, global_geodetic, etc.)
	pd.Grids = demo.GridMap

	// Register sources with unique names
	pd.Sources["tile_source"] = &tileSource

	// Register caches that will store tiles from sources
	pd.Caches["template_cache"] = &tileCache

	// Set the service configuration
	pd.Service = &tileService

	return pd
}

// getService creates and initializes the tileproxy service
// This function builds the actual service from the configuration
func getService() *tileproxy.Service {
	return tileproxy.NewService(getProxyService(), &demo.Globals, nil)
}

// dataset holds the initialized service instance
var dataset *tileproxy.Service

// ProxyServer is the HTTP request handler
// It initializes the service on first request and delegates to the service
func ProxyServer(w http.ResponseWriter, req *http.Request) {
	if dataset == nil {
		dataset = getService()
	}
	dataset.Service.ServeHTTP(w, req)
}

// main starts the HTTP server
// It registers the handler and begins listening on the configured port
func main() {
	// Register the proxy handler for all requests
	http.HandleFunc("/", ProxyServer)

	// Start the server and log any errors
	log.Printf("Starting %s proxy on port %s", SERVICE_NAME, PORT)
	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
