package elevation

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type IntersectionResponse struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Ground float64 `json:"ground"`
	Hit    bool    `json:"hit"`
}

// TileMeta описывает метаданные тайла в ответе HandleHeight
type TileMeta struct {
	Z int `json:"z"`
	X int `json:"x"`
	Y int `json:"y"`
}

// HeightResponse описывает JSON-ответ HandleHeight
type HeightResponse struct {
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	Height     float64  `json:"height"`
	Tile       TileMeta `json:"tile"`
	TileSource string   `json:"tile_source"`
	GridSize   int      `json:"grid_size"`
}
type Server struct {
	Store *Store
}

func (s *Server) HandleIntersection(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	camLat, err := strconv.ParseFloat(q.Get("cam_lat"), 64)
	if err != nil {
		http.Error(w, "invalid cam_lat", http.StatusBadRequest)
		return
	}
	camLon, err := strconv.ParseFloat(q.Get("cam_lon"), 64)
	if err != nil {
		http.Error(w, "invalid cam_lon", http.StatusBadRequest)
		return
	}
	camAlt, err := strconv.ParseFloat(q.Get("cam_alt"), 64)
	if err != nil {
		http.Error(w, "invalid cam_alt", http.StatusBadRequest)
		return
	}

	quatStr := q.Get("quat")
	if quatStr == "" {
		http.Error(w, "quat parameter required (4 comma-separated values)", http.StatusBadRequest)
		return
	}
	quatParts := splitFloat64(quatStr, ",")
	if len(quatParts) != 4 {
		http.Error(w, "quat must have exactly 4 values", http.StatusBadRequest)
		return
	}
	var quat [4]float64
	copy(quat[:], quatParts)

	zoom := s.Store.Config().DefaultZoom
	if zq := q.Get("z"); zq != "" {
		if zi, err := strconv.Atoi(zq); err == nil {
			zoom = zi
		}
	}

	step := 1.0
	if sq := q.Get("step"); sq != "" {
		if sf, err := strconv.ParseFloat(sq, 64); err == nil {
			step = sf
		}
	}

	maxDist := 5000.0
	if md := q.Get("max_dist"); md != "" {
		if mf, err := strconv.ParseFloat(md, 64); err == nil {
			maxDist = mf
		}
	}

	req := IntersectionRequest{
		CamLon:  camLon,
		CamLat:  camLat,
		CamAlt:  camAlt,
		Quat:    quat,
		Zoom:    zoom,
		Step:    step,
		MaxDist: maxDist,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := SearchIntersection(ctx, s.Store, req)
	if err != nil {
		http.Error(w, "intersection search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := IntersectionResponse{
		Lat:    result.Lat,
		Lon:    result.Lon,
		Ground: result.Ground,
		Hit:    result.Hit,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func splitFloat64(s, delim string) []float64 {
	parts := splitString(s, delim)
	result := make([]float64, 0, len(parts))
	for _, p := range parts {
		if f, err := strconv.ParseFloat(p, 64); err == nil {
			result = append(result, f)
		}
	}
	return result
}

func splitString(s, delim string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(delim) <= len(s) && s[i:i+len(delim)] == delim {
			result = append(result, s[start:i])
			start = i + len(delim)
			i += len(delim) - 1
		}
	}
	result = append(result, s[start:])
	return result
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

	req := HeightRequest{
		Lat:  lat,
		Lon:  lon,
		Zoom: z,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := PickHeight(ctx, s.Store, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := HeightResponse{
		Lat:        result.Lat,
		Lon:        result.Lon,
		Height:     result.Height,
		Tile:       TileMeta{Z: result.Meta.Z, X: result.Meta.X, Y: result.Meta.Y},
		TileSource: result.Meta.Source,
		GridSize:   result.Meta.GridSize,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
