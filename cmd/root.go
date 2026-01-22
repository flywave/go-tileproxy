package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tileproxy",
	Short: "A fast and flexible tile proxy service",
	Long:  `Go-based tile proxy supporting TMS, WMS, WMTS, Mapbox, and Cesium services with health monitoring`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
