package ddm

import "math"

// Вычисление индексов в тайле и долей для билинейной интерполяции.
// Координаты внутри тайла:
//   - столбцы слева-направо:   col = (lon - lon0) * 1200
//   - строки сверху-вниз (север->юг): row = (lat0+1 - lat) * 1200
//
// где lon0=floor(lon), lat0=floor(lat)
func pixelCoords(lat, lon float64) (rowF, colF float64, row0, col0 int, fy, fx float64, row1, col1 int) {
	lat0 := math.Floor(lat)
	lon0 := math.Floor(lon)

	col := (lon - lon0) * (gridSize - 1) // 0..1200
	row := (lat0 + 1.0 - lat) * (gridSize - 1)

	colF = col
	rowF = row

	col0 = int(math.Floor(col))
	row0 = int(math.Floor(row))
	if col0 < 0 {
		col0 = 0
	}
	if row0 < 0 {
		row0 = 0
	}
	if col0 >= gridSize-1 {
		col0 = gridSize - 2
	}
	if row0 >= gridSize-1 {
		row0 = gridSize - 2
	}

	fx = col - float64(col0) // 0..1
	fy = row - float64(row0) // 0..1

	col1 = col0 + 1
	row1 = row0 + 1
	return
}

func bilinear(p00, p10, p01, p11 float64, fx, fy float64) float64 {
	// p00 = (row0,col0), p10 = (row0,col1), p01 = (row1,col0), p11 = (row1,col1)
	a := p00*(1-fx) + p10*fx
	b := p01*(1-fx) + p11*fx
	return a*(1-fy) + b*fy
}

func resolveVoid(vals ...int16) (float64, bool) {
	sum := 0.0
	cnt := 0
	for _, v := range vals {
		if v != voidVal {
			sum += float64(v)
			cnt++
		}
	}
	if cnt == 0 {
		return 0, false
	}
	return sum / float64(cnt), true
}
