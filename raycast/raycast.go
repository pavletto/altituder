package raycast

import (
	"math"

	"github.com/westphae/geomag/pkg/egm96"
)

// ----------- Константы и интерфейсы -----------------

const (
	RadiusOfEarth = 6378137.0
)

type ElevationSource interface {
	Height(lat, lon float64) float64
}

// ----------- Кватернион → вектор направления -----------------

func QuaternionToForwardPX4(q [4]float64) [3]float64 {
	w := q[0]
	x := q[1]
	y := q[2]
	z := q[3]

	n := math.Sqrt(w*w + x*x + y*y + z*z)
	if n < 1e-9 {
		return [3]float64{0, 0, -1}
	}
	w /= n
	x /= n
	y /= n
	z /= n

	vx, vy, vz := 1.0, 0.0, 0.0

	ix := w*vx + y*vz - z*vy
	iy := w*vy + z*vx - x*vz
	iz := w*vz + x*vy - y*vx
	iw := -x*vx - y*vy - z*vz

	rx := ix*w - iw*x + iy*-z - iz*-y
	ry := iy*w - iw*y + iz*-x - ix*-z
	rz := iz*w - iw*z + ix*-y - iy*-x

	return [3]float64{rx, ry, rz}
}

// ----------- Основной алгоритм трассировки -----------------

type RaycastParams struct {
	CamLon, CamLat, CamAlt float64
	Quat                   [4]float64
	DEM                    ElevationSource
	Step, MaxDist          float64
}

// Raycast возвращает точку пересечения луча камеры с землёй.
// CamAlt задаётся по эллипсоиду (GPS), но переводится в MSL через EGM96.
// DEM уже по MSL — сравнение выполняется в одной системе (MSL).
func Raycast(p RaycastParams) (lon, lat, ground float64, hit bool) {
	if p.DEM == nil {
		return 0, 0, 0, false
	}

	if p.Step <= 0 {
		p.Step = 1
	}
	if p.MaxDist <= 0 {
		p.MaxDist = 3000
	}

	// направление из PX4 кватерниона
	dir := QuaternionToForwardPX4(p.Quat)
	dirN, dirE, dirD := dir[0], dir[1], dir[2]

	curLat := p.CamLat
	curLon := p.CamLon

	// --- перевод высоты дрона (WGS84 эллипсоид) → MSL через EGM96 ---
	ellAlt := p.CamAlt
	loc := egm96.NewLocationGeodetic(curLat, curLon, ellAlt)
	hMSL, err := loc.HeightAboveMSL()
	if err != nil {
		// fallback: если ошибка, оставляем как есть
		hMSL = ellAlt
	}
	curAlt := hMSL // теперь в MSL

	dist := 0.0
	prevLat, prevLon, prevAlt := curLat, curLon, curAlt

	for dist <= p.MaxDist && curLat <= 85 && curLat >= -85 {
		g := p.DEM.Height(curLat, curLon) // MSL

		if curAlt <= g {
			// бинарный поиск для уточнения
			for i := 0; i < 20; i++ {
				midLat := 0.5 * (prevLat + curLat)
				midLon := 0.5 * (prevLon + curLon)
				midAlt := 0.5 * (prevAlt + curAlt)
				gm := p.DEM.Height(midLat, midLon)
				if midAlt > gm {
					prevLat, prevLon, prevAlt = midLat, midLon, midAlt
				} else {
					curLat, curLon, curAlt = midLat, midLon, midAlt
				}
			}
			return curLon, curLat, g, true
		}

		dist += p.Step
		curAlt += -dirD * p.Step
		dNorth := dirN * p.Step
		dEast := dirE * p.Step

		curLat += (dNorth / RadiusOfEarth) * (180 / math.Pi)
		curLon += (dEast / (RadiusOfEarth * math.Cos(curLat*math.Pi/180))) * (180 / math.Pi)

		if curLon > 180 {
			curLon -= 360
		}
		if curLon < -180 {
			curLon += 360
		}

		prevLat, prevLon, prevAlt = curLat, curLon, curAlt
	}

	return curLon, curLat, curAlt, false
}
