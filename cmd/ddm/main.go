package ddm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pavletto/altituder/cmd/terrain"
)

type Server struct {
	Store *Store
}

func (s *Server) HandleIntersection(w http.ResponseWriter) {
	adapter := &DEMAdapter{
		Store:   s.Store, // твой *ddm.Store
		Zoom:    14,
		Timeout: 5 * time.Second,
	}
	q := [4]float64{28105, 2541, -4451, 16046}
	for i := range q {
		q[i] /= 32767.0
	}

	params := terrain.RaycastParams{
		CamLon:  55.729469896669606,
		CamLat:  25.00104389507723,
		CamAlt:  177.72191600000002,
		Quat:    q,
		Step:    1.0,
		MaxDist: 5000.0,
		DEM:     adapter,
	}

	lon, lat, ground, hit := terrain.Raycast(params)

	fmt.Printf("hit=%v\nlon=%.10f\nlat=%.10f\nground=%.3f\n", hit, lon, lat, ground)

	resp := map[string]any{
		"lat":    lat,
		"lon":    lon,
		"ground": ground,
		"hit":    hit,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleHeight(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	lat, err := strconv.ParseFloat(q.Get("lat"), 64)
	if err != nil {
		http.Error(w, "invalid lat", http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(q.Get("lon"), 64)
	if err != nil {
		http.Error(w, "invalid lon", http.StatusBadRequest)
		return
	}

	z := s.Store.Config().DefaultZoom
	if zq := q.Get("z"); zq != "" {
		if zi, err := strconv.Atoi(zq); err == nil {
			z = zi
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h, meta, err := s.Store.Height(ctx, lat, lon, z)
	if err != nil {
		http.Error(w, "height lookup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"lat":         lat,
		"lon":         lon,
		"height":      h,
		"tile":        map[string]any{"z": meta.Z, "x": meta.X, "y": meta.Y},
		"tile_source": meta.Source, // mem-cache | disk-cache | download
		"grid_size":   meta.GridSize,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
