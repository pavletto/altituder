package elevation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StoreConfig struct {
	CacheDir          string
	URLTemplate       string // "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm"
	Subdomains        []string
	PermitDownload    bool
	HTTPClientTimeout time.Duration

	DefaultZoom   int
	MaxNativeZoom int

	HeightFactor float32
	NoDataValues []float32

	MaxMemTiles int
}

type Meta struct {
	Z, X, Y  int
	Source   string // mem-cache | disk-cache | download
	GridSize int
}

type Store struct {
	cfg   StoreConfig
	http  *http.Client
	memMu sync.Mutex
	mem   *lru // key "z/x/y"
	subIx int
}

func NewStore(cfg StoreConfig) (*Store, error) {
	if cfg.CacheDir == "" {
		return nil, fmt.Errorf("CacheDir required")
	}
	if err := os.MkdirAll(cfg.CacheDir, 0o755); err != nil {
		return nil, err
	}
	if cfg.MaxMemTiles <= 0 {
		cfg.MaxMemTiles = 64
	}
	return &Store{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.HTTPClientTimeout},
		mem:  newLRU(cfg.MaxMemTiles),
	}, nil
}

func (s *Store) Config() StoreConfig { return s.cfg }

func (s *Store) Height(ctx context.Context, lat, lon float64, z int) (float64, Meta, error) {
	if z <= 0 {
		z = s.cfg.DefaultZoom
	}

	x, y := tileXYZ(lat, lon, z)
	key := fmt.Sprintf("%d/%d/%d", z, y, x)
	meta := Meta{Z: z, X: x, Y: y}

	// 1) mem
	if td, ok := s.getMem(key); ok {
		meta.Source = "mem-cache"
		h, ok := heightFromTile(td, lat, lon, z, x, y)
		meta.GridSize = td.GridSize
		if ok {
			return h, meta, nil
		}
		return 0, meta, fmt.Errorf("nodata around point")
	}

	// 2) disk
	if td, err := s.loadFromDisk(z, x, y); err == nil {
		s.putMem(key, td)
		meta.Source = "disk-cache"
		meta.GridSize = td.GridSize
		h, ok := heightFromTile(td, lat, lon, z, x, y)
		if ok {
			return h, meta, nil
		}
		return 0, meta, fmt.Errorf("nodata around point")
	}

	// 3) download
	if s.cfg.PermitDownload {
		td, err := s.downloadTile(ctx, z, x, y)
		if err != nil {
			return 0, meta, err
		}
		s.putMem(key, td)
		meta.Source = "download"
		meta.GridSize = td.GridSize
		h, ok := heightFromTile(td, lat, lon, z, x, y)
		if ok {
			return h, meta, nil
		}
		return 0, meta, fmt.Errorf("nodata around point")
	}

	return 0, meta, fmt.Errorf("tile not found and download disabled")
}

func heightFromTile(td *tileData, lat, lon float64, z, x, y int) (float64, bool) {
	fx, fy := tileFrac(lat, lon, z, x, y) // 0..1
	return td.heightAtFrac(fx, fy)
}

func (s *Store) expandURL(z, x, y int) string {
	u := s.cfg.URLTemplate
	// подставим {s} циклически
	sub := ""
	if len(s.cfg.Subdomains) > 0 {
		s.subIx = (s.subIx + 1) % len(s.cfg.Subdomains)
		sub = s.cfg.Subdomains[s.subIx]
	}
	repl := map[string]string{
		"{s}": fmt.Sprintf("%s", sub),
		"{z}": fmt.Sprintf("%d", z),
		"{x}": fmt.Sprintf("%d", x),
		"{y}": fmt.Sprintf("%d", y),
	}
	for k, v := range repl {
		u = strings.ReplaceAll(u, k, v)
	}
	return u
}

func (s *Store) cachePath(z, x, y int) string {
	return filepath.Join(s.cfg.CacheDir, fmt.Sprintf("%d/%d/%d.ddm", z, y, x))
}

func (s *Store) loadFromDisk(z, x, y int) (*tileData, error) {
	path := s.cachePath(z, x, y)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseDDM(raw, z, x, y, s.cfg.NoDataValues)
}

func (s *Store) downloadTile(ctx context.Context, z, x, y int) (*tileData, error) {
	url := s.expandURL(z, x, y)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, url)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// сохранить
	path := s.cachePath(z, x, y)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return nil, err
	}
	return parseDDM(raw, z, x, y, s.cfg.NoDataValues)
}

// простая LRU

type lru struct {
	cap int
	ll  []string
	m   map[string]*tileData
}

func newLRU(cap int) *lru {
	if cap <= 0 {
		cap = 1
	}
	return &lru{cap: cap, ll: make([]string, 0, cap), m: make(map[string]*tileData)}
}
func (l *lru) get(k string) (*tileData, bool) {
	if v, ok := l.m[k]; ok {
		l.touch(k)
		return v, true
	}
	return nil, false
}
func (l *lru) put(k string, v *tileData) {
	if _, ok := l.m[k]; ok {
		l.m[k] = v
		l.touch(k)
		return
	}
	if len(l.ll) == l.cap {
		evict := l.ll[len(l.ll)-1]
		delete(l.m, evict)
		l.ll = l.ll[:len(l.ll)-1]
	}
	l.ll = append([]string{k}, l.ll...)
	l.m[k] = v
}
func (l *lru) touch(k string) {
	idx := -1
	for i, s := range l.ll {
		if s == k {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return
	}
	copy(l.ll[1:idx+1], l.ll[0:idx])
	l.ll[0] = k
}

func (s *Store) getMem(key string) (*tileData, bool) {
	s.memMu.Lock()
	defer s.memMu.Unlock()
	return s.mem.get(key)
}
func (s *Store) putMem(key string, td *tileData) {
	s.memMu.Lock()
	defer s.memMu.Unlock()
	s.mem.put(key, td)
}

// вспомогательное
func ParseNoData(csv string) []float32 {
	if csv == "" {
		return nil
	}
	var out []float32
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if f, err := parseFloat32(p); err == nil {
			out = append(out, f)
		}
	}
	return out
}
func parseFloat32(s string) (float32, error) {
	f, err := strconv.ParseFloat(s, 32)
	return float32(f), err
}
