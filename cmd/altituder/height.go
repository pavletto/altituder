package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pavletto/altituder/elevation"
	"github.com/spf13/cobra"
)

// heightCmd represents the height command
var heightCmd = &cobra.Command{
	Use:   "height",
	Short: "Get terrain elevation at a location",
	Long: `Get terrain elevation at a specific geographic coordinate.

Examples:
  altituder height --lat 25.0 --lon 55.0
  altituder height --lat 25.0 --lon 55.0 --zoom 15
  altituder height --lat 25.197525 --lon 55.274288 --cache-dir /tmp/cache

The command will output the elevation in meters along with tile information.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse required flags
		lat, _ := cmd.Flags().GetFloat64("lat")
		lon, _ := cmd.Flags().GetFloat64("lon")

		// Validate
		if lat < -90 || lat > 90 {
			log.Fatal("Latitude must be between -90 and 90")
		}
		if lon < -180 || lon > 180 {
			log.Fatal("Longitude must be between -180 and 180")
		}

		// Load configuration
		cfg := LoadConfig(cmd)

		// Create store
		store, err := cfg.CreateStore()
		if err != nil {
			log.Fatalf("Failed to create store: %v", err)
		}

		// Get zoom level
		zoom := cfg.DefaultZoom
		if cmd.Flags().Changed("zoom") {
			zoom, _ = cmd.Flags().GetInt("zoom")
		}

		// Build request
		req := elevation.HeightRequest{
			Lat:  lat,
			Lon:  lon,
			Zoom: zoom,
		}

		// Execute
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := elevation.PickHeight(ctx, store, req)
		if err != nil {
			log.Fatalf("Failed to get height: %v", err)
		}

		// Output results
		fmt.Printf("Location: %.6f, %.6f\n", result.Lat, result.Lon)
		fmt.Printf("Elevation: %.2f meters\n", result.Height)
		fmt.Printf("Tile: z=%d x=%d y=%d\n", result.Meta.Z, result.Meta.X, result.Meta.Y)
		fmt.Printf("Source: %s\n", result.Meta.Source)
		fmt.Printf("Grid Size: %d\n", result.Meta.GridSize)
	},
}

func init() {
	rootCmd.AddCommand(heightCmd)

	// Required flags
	heightCmd.Flags().Float64("lat", 0, "Latitude (required)")
	heightCmd.Flags().Float64("lon", 0, "Longitude (required)")
	heightCmd.MarkFlagRequired("lat")
	heightCmd.MarkFlagRequired("lon")

	// Optional flags - inherit from parent but can be overridden
	heightCmd.Flags().IntP("zoom", "z", 0, "Zoom level (default from global config)")
}
