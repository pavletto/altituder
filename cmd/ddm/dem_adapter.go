package ddm

import (
	"context"
	"time"
)

// Обёртка, приводящая ddm.Store к простому интерфейсу ElevationSource
type DEMAdapter struct {
	Store   *Store
	Zoom    int
	Timeout time.Duration
}

func (a *DEMAdapter) Height(lat, lon float64) float64 {
	ctx, cancel := context.WithTimeout(context.Background(), a.Timeout)
	defer cancel()
	h, _, err := a.Store.Height(ctx, lat, lon, a.Zoom)
	if err != nil {
		return 0
	}
	return h
}
