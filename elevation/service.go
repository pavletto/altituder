package elevation

import (
	"context"
	"fmt"
	"time"

	"github.com/pavletto/altituder/raycast"
)

// IntersectionRequest contains parameters for raycast intersection search
type IntersectionRequest struct {
	CamLon  float64    // Camera longitude
	CamLat  float64    // Camera latitude
	CamAlt  float64    // Camera altitude (GPS/WGS84)
	Quat    [4]float64 // Quaternion [w, x, y, z]
	Zoom    int        // Tile zoom level
	Step    float64    // Step size for raycast
	MaxDist float64    // Maximum distance to search
}

// IntersectionResult contains the result of intersection search
type IntersectionResult struct {
	Lon    float64 // Intersection longitude
	Lat    float64 // Intersection latitude
	Ground float64 // Ground elevation at intersection
	Hit    bool    // Whether intersection was found
}

// HeightRequest contains parameters for height lookup
type HeightRequest struct {
	Lat  float64 // Latitude
	Lon  float64 // Longitude
	Zoom int     // Tile zoom level
}

// HeightResult contains the result of height lookup
type HeightResult struct {
	Lat    float64 // Requested latitude
	Lon    float64 // Requested longitude
	Height float64 // Height at the location
	Meta   Meta    // Metadata about the tile used
}

// SearchIntersection performs raycast to find terrain intersection
// This is a reusable business logic function that can be called from
// HTTP handlers, CLI tools, or other contexts
func SearchIntersection(ctx context.Context, store *Store, req IntersectionRequest) (IntersectionResult, error) {
	if store == nil {
		return IntersectionResult{}, fmt.Errorf("store is nil")
	}

	// Set defaults
	if req.Zoom <= 0 {
		req.Zoom = store.Config().DefaultZoom
	}
	if req.Step <= 0 {
		req.Step = 1.0
	}
	if req.MaxDist <= 0 {
		req.MaxDist = 5000.0
	}

	// Create DEM adapter for terrain raycast
	adapter := &DEMAdapter{
		Store:   store,
		Zoom:    req.Zoom,
		Timeout: 5 * time.Second,
	}

	// Perform raycast
	params := raycast.RaycastParams{
		CamLon:  req.CamLon,
		CamLat:  req.CamLat,
		CamAlt:  req.CamAlt,
		Quat:    req.Quat,
		Step:    req.Step,
		MaxDist: req.MaxDist,
		DEM:     adapter,
	}

	lon, lat, ground, hit := raycast.Raycast(params)

	return IntersectionResult{
		Lon:    lon,
		Lat:    lat,
		Ground: ground,
		Hit:    hit,
	}, nil
}

// PickHeight retrieves elevation at a specific location
// This is a reusable business logic function that can be called from
// HTTP handlers, CLI tools, or other contexts
func PickHeight(ctx context.Context, store *Store, req HeightRequest) (HeightResult, error) {
	if store == nil {
		return HeightResult{}, fmt.Errorf("store is nil")
	}

	// Set default zoom if not provided
	zoom := req.Zoom
	if zoom <= 0 {
		zoom = store.Config().DefaultZoom
	}

	// Lookup height
	h, meta, err := store.Height(ctx, req.Lat, req.Lon, zoom)
	if err != nil {
		return HeightResult{}, fmt.Errorf("height lookup failed: %w", err)
	}

	return HeightResult{
		Lat:    req.Lat,
		Lon:    req.Lon,
		Height: h,
		Meta:   meta,
	}, nil
}
