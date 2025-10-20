package elevation

import "math"

const (
	maxLat = 85.05112878
	minLat = -85.05112878
)

// WGS84 lat/lon -> Spherical Mercator tile indices
func tileXYZ(lat, lon float64, z int) (x, y int) {
	if lat > maxLat {
		lat = maxLat
	}
	if lat < minLat {
		lat = minLat
	}
	n := math.Exp2(float64(z))
	x = int((lon + 180.0) / 360.0 * n)
	y = int((1.0 - math.Log(math.Tan(rad(lat))+sec(rad(lat)))/math.Pi) / 2.0 * n)
	return
}

func tileFrac(lat, lon float64, z, x, y int) (fx, fy float64) {
	n := math.Exp2(float64(z))
	wx := (lon + 180.0) / 360.0 * n
	wy := (1.0 - math.Log(math.Tan(rad(lat))+sec(rad(lat)))/math.Pi) / 2.0 * n
	return wx - float64(x), wy - float64(y)
}

func rad(d float64) float64 { return d * math.Pi / 180.0 }
func sec(r float64) float64 { return 1.0 / math.Cos(r) }
