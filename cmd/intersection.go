package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pavletto/altituder/cmd/ddm"
	"github.com/spf13/cobra"
)

// intersectionCmd represents the intersection command
var intersectionCmd = &cobra.Command{
	Use:   "intersection",
	Short: "Find terrain intersection via raycast",
	Long: `Perform a raycast from a camera position to find terrain intersection.

This command simulates a camera view (e.g., from a drone) and finds where
the camera's line of sight intersects with the terrain.

Examples:
  altituder intersection --cam-lat 25.001 --cam-lon 55.729 --cam-alt 177.72 --quat 0.8581,0.0776,-0.1359,0.4899
  altituder intersection --cam-lat 25.001 --cam-lon 55.729 --cam-alt 177.72 --quat 1,0,0,0 --step 2 --max-dist 1000

The quaternion represents camera orientation in w,x,y,z format (must be 4 values).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse required flags
		camLat, _ := cmd.Flags().GetFloat64("cam-lat")
		camLon, _ := cmd.Flags().GetFloat64("cam-lon")
		camAlt, _ := cmd.Flags().GetFloat64("cam-alt")
		quatStr, _ := cmd.Flags().GetString("quat")

		// Validate
		if camLat < -90 || camLat > 90 {
			log.Fatal("Camera latitude must be between -90 and 90")
		}
		if camLon < -180 || camLon > 180 {
			log.Fatal("Camera longitude must be between -180 and 180")
		}

		// Parse quaternion
		quatParts := strings.Split(quatStr, ",")
		if len(quatParts) != 4 {
			log.Fatal("Quaternion must have exactly 4 values (w,x,y,z)")
		}
		var quat [4]float64
		for i, p := range quatParts {
			val, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				log.Fatalf("Invalid quaternion value '%s': %v", p, err)
			}
			quat[i] = val
		}

		// Load configuration
		cfg := LoadConfig(cmd)

		// Create store
		store, err := cfg.CreateStore()
		if err != nil {
			log.Fatalf("Failed to create store: %v", err)
		}

		// Get optional parameters
		zoom := cfg.DefaultZoom
		if cmd.Flags().Changed("zoom") {
			zoom, _ = cmd.Flags().GetInt("zoom")
		}

		step, _ := cmd.Flags().GetFloat64("step")
		maxDist, _ := cmd.Flags().GetFloat64("max-dist")

		// Build request
		req := ddm.IntersectionRequest{
			CamLon:  camLon,
			CamLat:  camLat,
			CamAlt:  camAlt,
			Quat:    quat,
			Zoom:    zoom,
			Step:    step,
			MaxDist: maxDist,
		}

		// Execute
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ddm.SearchIntersection(ctx, store, req)
		if err != nil {
			log.Fatalf("Failed to find intersection: %v", err)
		}

		// Output results
		fmt.Printf("Camera Position: %.6f, %.6f @ %.2fm\n", camLat, camLon, camAlt)
		fmt.Printf("Quaternion: [%.4f, %.4f, %.4f, %.4f]\n", quat[0], quat[1], quat[2], quat[3])
		fmt.Println()
		
		if result.Hit {
			fmt.Println("✓ Terrain intersection found!")
			fmt.Printf("  Location: %.6f, %.6f\n", result.Lat, result.Lon)
			fmt.Printf("  Ground Elevation: %.2f meters\n", result.Ground)
		} else {
			fmt.Println("✗ No terrain intersection found within search distance")
			fmt.Printf("  Last checked: %.6f, %.6f\n", result.Lat, result.Lon)
		}
	},
}

func init() {
	rootCmd.AddCommand(intersectionCmd)

	// Required flags
	intersectionCmd.Flags().Float64("cam-lat", 0, "Camera latitude (required)")
	intersectionCmd.Flags().Float64("cam-lon", 0, "Camera longitude (required)")
	intersectionCmd.Flags().Float64("cam-alt", 0, "Camera altitude in meters (required)")
	intersectionCmd.Flags().String("quat", "", "Camera quaternion as w,x,y,z (required)")
	intersectionCmd.MarkFlagRequired("cam-lat")
	intersectionCmd.MarkFlagRequired("cam-lon")
	intersectionCmd.MarkFlagRequired("cam-alt")
	intersectionCmd.MarkFlagRequired("quat")

	// Optional flags
	intersectionCmd.Flags().IntP("zoom", "z", 0, "Zoom level (default from global config)")
	intersectionCmd.Flags().Float64("step", 1.0, "Step size for raycast in meters")
	intersectionCmd.Flags().Float64("max-dist", 5000.0, "Maximum search distance in meters")
}
