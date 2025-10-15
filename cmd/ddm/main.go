package ddm

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Store *Store
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

func main() {
	cacheDir := getenv("DDM_CACHE_DIR", "./cache")
	urlTpl := getenv("DDM_URL_TEMPLATE", "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm")
	subs := getenv("DDM_SUBDOMAINS", "a,b,c")
	heightFactor := getenvFloat("DDM_HEIGHT_FACTOR", 1.0)
	defaultZoom := getenvInt("DDM_DEFAULT_Z", 17) // как в вашем примере maxZoom=17
	maxNativeZoom := getenvInt("DDM_MAX_NATIVE_Z", 14)
	noDataCSV := os.Getenv("DDM_NODATA_CSV") // например: "-32768,3.4028235e+38"

	cfg := StoreConfig{
		CacheDir:          cacheDir,
		URLTemplate:       urlTpl,
		Subdomains:        strings.Split(subs, ","),
		PermitDownload:    urlTpl != "",
		HTTPClientTimeout: 15 * time.Second,
		DefaultZoom:       defaultZoom,
		MaxNativeZoom:     maxNativeZoom,
		HeightFactor:      float32(heightFactor),
		NoDataValues:      ParseNoData(noDataCSV),
	}

	store, err := NewStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{Store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("/height", s.HandleHeight)
	mux.HandleFunc("/health", s.HandleHealth)

	addr := getenv("ADDR", ":8080")
	log.Printf("listening on %s cache=%s download=%v tpl=%s", addr, cacheDir, cfg.PermitDownload, urlTpl)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func getenvInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}
func getenvFloat(k string, d float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return d
}
