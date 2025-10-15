package ddm

import "math"

func floor(v float64) int {
	return int(math.Floor(v))
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

const (
	gridSize = 1201 // SRTM3
	voidVal  = int16(-32768)
)

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
