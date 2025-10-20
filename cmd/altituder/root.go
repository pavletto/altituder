/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "altituder",
	Short: "Terrain elevation and raycast intersection service",
	Long: `Altituder is a Go-based terrain elevation and raycast intersection service.

It provides both CLI commands and HTTP API endpoints for:
- Height Lookup: Get terrain elevation at any geographic coordinate
- Intersection Search: Perform raycast to find terrain intersections

Configuration can be set via environment variables or command-line flags.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("cache-dir", "c", "./cache", "Cache directory for DEM tiles")
	rootCmd.PersistentFlags().String("url-template", "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm", "URL template for downloading tiles")
	rootCmd.PersistentFlags().String("subdomains", "a,b,c", "Comma-separated list of subdomains")
	rootCmd.PersistentFlags().IntP("zoom", "z", 14, "Default zoom level")
	rootCmd.PersistentFlags().Int("max-native-zoom", 14, "Maximum native zoom level")
	rootCmd.PersistentFlags().Float64("height-factor", 1.0, "Height multiplication factor")
	rootCmd.PersistentFlags().String("nodata-values", "", "Comma-separated list of no-data values (e.g., '-32768,3.4028235e+38')")
}
