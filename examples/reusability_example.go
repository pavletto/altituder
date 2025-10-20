package examples

// Example: How to use the refactored business logic functions
// This file demonstrates the reusability of SearchIntersection and PickHeight
// outside of HTTP handlers (e.g., in a CLI tool, batch processor, or other service)

import (
	"context"
	"fmt"
	"log"
	"time"

	elevation2 "github.com/pavletto/altituder/elevation"
)

// ExampleHeightLookup demonstrates using PickHeight in a non-HTTP context
func ExampleHeightLookup() {
	// Setup store configuration
	cfg := elevation2.StoreConfig{
		CacheDir:       "./cache",
		URLTemplate:    "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm",
		Subdomains:     []string{"a", "b", "c"},
		PermitDownload: true,
		DefaultZoom:    14,
		MaxNativeZoom:  14,
		HeightFactor:   1.0,
	}

	store, err := elevation2.NewStore(cfg)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Create a height request
	req := elevation2.HeightRequest{
		Lat:  25.0,
		Lon:  55.0,
		Zoom: 14,
	}

	// Call the business logic function directly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := elevation2.PickHeight(ctx, store, req)
	if err != nil {
		log.Printf("Height lookup failed: %v", err)
		return
	}

	fmt.Printf("Height at (%.6f, %.6f): %.2f meters\n", result.Lat, result.Lon, result.Height)
	fmt.Printf("Tile: z=%d x=%d y=%d\n", result.Meta.Z, result.Meta.X, result.Meta.Y)
	fmt.Printf("Source: %s\n", result.Meta.Source)
}

// ExampleBatchHeightLookup demonstrates processing multiple locations
func ExampleBatchHeightLookup() {
	cfg := elevation2.StoreConfig{
		CacheDir:       "./cache",
		URLTemplate:    "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm",
		Subdomains:     []string{"a", "b", "c"},
		PermitDownload: true,
		DefaultZoom:    14,
		MaxNativeZoom:  14,
		HeightFactor:   1.0,
	}

	store, err := elevation2.NewStore(cfg)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Process multiple locations
	locations := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Dubai Downtown", 25.197525, 55.274288},
		{"Burj Khalifa", 25.197197, 55.274376},
		{"Dubai Marina", 25.080524, 55.139373},
	}

	for _, loc := range locations {
		req := elevation2.HeightRequest{
			Lat:  loc.lat,
			Lon:  loc.lon,
			Zoom: 14,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		result, err := elevation2.PickHeight(ctx, store, req)
		cancel()

		if err != nil {
			log.Printf("%s: Height lookup failed: %v", loc.name, err)
			continue
		}

		fmt.Printf("%s (%.6f, %.6f): %.2f meters\n", loc.name, result.Lat, result.Lon, result.Height)
	}
}

// ExampleIntersectionSearch demonstrates using SearchIntersection in a non-HTTP context
func ExampleIntersectionSearch() {
	cfg := elevation2.StoreConfig{
		CacheDir:       "./cache",
		URLTemplate:    "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm",
		Subdomains:     []string{"a", "b", "c"},
		PermitDownload: true,
		DefaultZoom:    14,
		MaxNativeZoom:  14,
		HeightFactor:   1.0,
	}

	store, err := elevation2.NewStore(cfg)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Simulate a drone camera looking at terrain
	req := elevation2.IntersectionRequest{
		CamLon:  55.729469896669606,
		CamLat:  25.00104389507723,
		CamAlt:  177.72191600000002,
		Quat:    [4]float64{0.8581, 0.0776, -0.1359, 0.4899}, // normalized quaternion
		Zoom:    14,
		Step:    1.0,
		MaxDist: 5000.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := elevation2.SearchIntersection(ctx, store, req)
	if err != nil {
		log.Printf("Intersection search failed: %v", err)
		return
	}

	if result.Hit {
		fmt.Printf("Terrain intersection found at (%.6f, %.6f)\n", result.Lat, result.Lon)
		fmt.Printf("Ground elevation: %.2f meters\n", result.Ground)
	} else {
		fmt.Println("No terrain intersection found within search distance")
	}
}

// ExampleCLITool demonstrates how these functions could be used in a CLI tool
func ExampleCLITool(args []string) {
	// Example CLI command: altituder height --lat 25.0 --lon 55.0 --zoom 14
	// This would parse args and call PickHeight

	// Example CLI command: altituder intersect --cam-lat 25.0 --cam-lon 55.0 --cam-alt 100 --quat "1,0,0,0"
	// This would parse args and call SearchIntersection

	fmt.Println("See ARCHITECTURE.md for API usage examples")
}

// Note: This file is for documentation purposes and demonstrates reusability
// To use these examples, you would create appropriate CLI commands or services that call these functions
