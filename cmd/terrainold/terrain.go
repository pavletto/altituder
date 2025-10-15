package terrain

import "math"

const (
	wgsA  = 6378137.0
	wgsE2 = 6.69437999014e-3
)

// Источник высот (например, SRTM)
type ElevationSource interface {
	Height(lat, lon float64) float64
}

// ---------------- Геодезия ----------------

// Широта, долгота, высота → ECEF
func ToECEF(lonDeg, latDeg, h float64) (x, y, z float64) {
	lon := lonDeg * math.Pi / 180
	lat := latDeg * math.Pi / 180
	sinLat, cosLat := math.Sin(lat), math.Cos(lat)
	sinLon, cosLon := math.Sin(lon), math.Cos(lon)

	N := wgsA / math.Sqrt(1-wgsE2*sinLat*sinLat)
	x = (N + h) * cosLat * cosLon
	y = (N + h) * cosLat * sinLon
	z = (N*(1-wgsE2) + h) * sinLat
	return
}

// ECEF → географические координаты
func ECEFToLLA(x, y, z float64) (lonDeg, latDeg, h float64) {
	lon := math.Atan2(y, x)
	r := math.Sqrt(x*x + y*y)
	lat := math.Atan2(z, r*(1-wgsE2))
	for i := 0; i < 5; i++ {
		sinLat := math.Sin(lat)
		N := wgsA / math.Sqrt(1-wgsE2*sinLat*sinLat)
		h = r/math.Cos(lat) - N
		lat = math.Atan2(z, r*(1-wgsE2*(N/(N+h))))
	}
	return lon * 180 / math.Pi, lat * 180 / math.Pi, h
}

// ---------------- Вектор направления ----------------

// yaw: 0 = север, 90 = восток
// pitch > 0 — вниз
func DirectionNED(yawDeg, pitchDeg float64) [3]float64 {
	yaw := yawDeg * math.Pi / 180
	pitch := pitchDeg * math.Pi / 180
	n := math.Cos(pitch) * math.Cos(yaw)
	e := math.Cos(pitch) * math.Sin(yaw)
	d := math.Sin(pitch)
	return [3]float64{n, e, d}
}

// ---------------- NED → ECEF ----------------

func nedToECEFMatrix(lonDeg, latDeg float64) [3][3]float64 {
	lon := lonDeg * math.Pi / 180
	lat := latDeg * math.Pi / 180
	sinLat, cosLat := math.Sin(lat), math.Cos(lat)
	sinLon, cosLon := math.Sin(lon), math.Cos(lon)

	return [3][3]float64{
		{-sinLat * cosLon, -sinLon, -cosLat * cosLon},
		{-sinLat * sinLon, cosLon, -cosLat * sinLon},
		{cosLat, 0, -sinLat},
	}
}

func mulMatVec(m [3][3]float64, v [3]float64) (x, y, z float64) {
	x = m[0][0]*v[0] + m[0][1]*v[1] + m[0][2]*v[2]
	y = m[1][0]*v[0] + m[1][1]*v[1] + m[1][2]*v[2]
	z = m[2][0]*v[0] + m[2][1]*v[1] + m[2][2]*v[2]
	return
}

// ---------------- Основной алгоритм ----------------

type RaycastParams struct {
	CamLon, CamLat, CamAlt float64
	Yaw, Pitch             float64
	DEM                    ElevationSource
	Step, MaxDist          float64
}

// Raycast возвращает точку пересечения луча камеры с землёй
func Raycast(p RaycastParams) (lon, lat, ground float64, hit bool) {
	if p.DEM == nil {
		return 0, 0, 0, false
	}

	if p.Step <= 0 {
		p.Step = 10
	}
	if p.MaxDist <= 0 {
		p.MaxDist = 5000
	}

	// Камера в ECEF
	ox, oy, oz := ToECEF(p.CamLon, p.CamLat, p.CamAlt)

	// Направление в NED
	dirNED := DirectionNED(p.Yaw, p.Pitch)

	// Вектор направления в ECEF
	M := nedToECEFMatrix(p.CamLon, p.CamLat)
	dx, dy, dz := mulMatVec(M, dirNED)

	// Марш по лучу
	x, y, z := ox, oy, oz
	prevX, prevY, prevZ := x, y, z

	for dist := 0.0; dist < p.MaxDist; dist += p.Step {
		prevX, prevY, prevZ = x, y, z
		x += dx * p.Step
		y += dy * p.Step
		z += dz * p.Step

		lon, lat, alt := ECEFToLLA(x, y, z)
		g := p.DEM.Height(lat, lon)

		if alt <= g {
			// Попали в землю — бинарный поиск для уточнения
			for i := 0; i < 20; i++ {
				mx := 0.5 * (prevX + x)
				my := 0.5 * (prevY + y)
				mz := 0.5 * (prevZ + z)
				lonM, latM, altM := ECEFToLLA(mx, my, mz)
				gM := p.DEM.Height(latM, lonM)
				if altM > gM {
					prevX, prevY, prevZ = mx, my, mz
				} else {
					x, y, z = mx, my, mz
				}
			}
			lon, lat, ground = ECEFToLLA(x, y, z)
			return lon, lat, ground, true
		}
	}

	// Земля не найдена
	lon, lat, h := ECEFToLLA(x, y, z)
	return lon, lat, h, false
}
