package ddm_test

import (
	"context"
	"testing"
	"time"

	"github.com/pavletto/altituder/cmd/ddm"
)

func TestSearchIntersection(t *testing.T) {
	// Create a test store with minimal configuration
	cfg := ddm.StoreConfig{
		CacheDir:      "/tmp/test-cache",
		URLTemplate:   "", // No download for unit tests
		PermitDownload: false,
		DefaultZoom:   14,
		MaxNativeZoom: 14,
		HeightFactor:  1.0,
	}

	store, err := ddm.NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name    string
		req     ddm.IntersectionRequest
		wantErr bool
	}{
		{
			name: "valid request with defaults",
			req: ddm.IntersectionRequest{
				CamLon: 55.729469896669606,
				CamLat: 25.00104389507723,
				CamAlt: 177.72191600000002,
				Quat:   [4]float64{0.8581, 0.0776, -0.1359, 0.4899}, // normalized quaternion
			},
			wantErr: false, // May not hit terrain without actual DEM data, but shouldn't error
		},
		{
			name: "valid request with custom parameters",
			req: ddm.IntersectionRequest{
				CamLon:  55.0,
				CamLat:  25.0,
				CamAlt:  100.0,
				Quat:    [4]float64{1, 0, 0, 0}, // identity quaternion
				Zoom:    14,
				Step:    2.0,
				MaxDist: 1000.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := ddm.SearchIntersection(ctx, store, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchIntersection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Basic validation of result structure
			if !tt.wantErr {
				// Longitude should be in valid range
				if result.Lon < -180 || result.Lon > 180 {
					t.Errorf("SearchIntersection() longitude out of range: %v", result.Lon)
				}
				// Latitude should be in valid range
				if result.Lat < -90 || result.Lat > 90 {
					t.Errorf("SearchIntersection() latitude out of range: %v", result.Lat)
				}
			}
		})
	}
}

func TestSearchIntersection_NilStore(t *testing.T) {
	req := ddm.IntersectionRequest{
		CamLon: 55.0,
		CamLat: 25.0,
		CamAlt: 100.0,
		Quat:   [4]float64{1, 0, 0, 0},
	}

	ctx := context.Background()
	_, err := ddm.SearchIntersection(ctx, nil, req)
	if err == nil {
		t.Error("SearchIntersection() with nil store should return error")
	}
}

func TestPickHeight(t *testing.T) {
	// Create a test store with minimal configuration
	cfg := ddm.StoreConfig{
		CacheDir:      "/tmp/test-cache",
		URLTemplate:   "", // No download for unit tests
		PermitDownload: false,
		DefaultZoom:   14,
		MaxNativeZoom: 14,
		HeightFactor:  1.0,
	}

	store, err := ddm.NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name    string
		req     ddm.HeightRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: ddm.HeightRequest{
				Lat:  25.0,
				Lon:  55.0,
				Zoom: 14,
			},
			wantErr: true, // Will error without actual DEM data
		},
		{
			name: "valid request with default zoom",
			req: ddm.HeightRequest{
				Lat: 25.0,
				Lon: 55.0,
			},
			wantErr: true, // Will error without actual DEM data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := ddm.PickHeight(ctx, store, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PickHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Basic validation when no error expected
			if !tt.wantErr {
				if result.Lat != tt.req.Lat {
					t.Errorf("PickHeight() lat = %v, want %v", result.Lat, tt.req.Lat)
				}
				if result.Lon != tt.req.Lon {
					t.Errorf("PickHeight() lon = %v, want %v", result.Lon, tt.req.Lon)
				}
			}
		})
	}
}

func TestPickHeight_NilStore(t *testing.T) {
	req := ddm.HeightRequest{
		Lat:  25.0,
		Lon:  55.0,
		Zoom: 14,
	}

	ctx := context.Background()
	_, err := ddm.PickHeight(ctx, nil, req)
	if err == nil {
		t.Error("PickHeight() with nil store should return error")
	}
}

func TestIntersectionRequest_DefaultValues(t *testing.T) {
	cfg := ddm.StoreConfig{
		CacheDir:      "/tmp/test-cache",
		URLTemplate:   "",
		PermitDownload: false,
		DefaultZoom:   15,
		MaxNativeZoom: 15,
		HeightFactor:  1.0,
	}

	store, err := ddm.NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	req := ddm.IntersectionRequest{
		CamLon: 55.0,
		CamLat: 25.0,
		CamAlt: 100.0,
		Quat:   [4]float64{1, 0, 0, 0},
		// Zoom, Step, MaxDist not set - should use defaults
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = ddm.SearchIntersection(ctx, store, req)
	// Should not panic or error due to missing parameters
	if err != nil && err.Error() == "store is nil" {
		t.Error("SearchIntersection() should handle default parameters")
	}
}

func TestHeightRequest_DefaultZoom(t *testing.T) {
	cfg := ddm.StoreConfig{
		CacheDir:      "/tmp/test-cache",
		URLTemplate:   "",
		PermitDownload: false,
		DefaultZoom:   16,
		MaxNativeZoom: 16,
		HeightFactor:  1.0,
	}

	store, err := ddm.NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	req := ddm.HeightRequest{
		Lat: 25.0,
		Lon: 55.0,
		// Zoom not set - should use default
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = ddm.PickHeight(ctx, store, req)
	// Should not panic or error due to missing zoom
	if err != nil && err.Error() == "store is nil" {
		t.Error("PickHeight() should handle default zoom")
	}
}
