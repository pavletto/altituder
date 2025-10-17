package elevation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

type tileData struct {
	Z, X, Y   int
	GridSize  int       // gs x gs
	Values    []float32 // длина gs*gs
	NoDataSet map[float32]struct{}
	Factor    float32
}

func parseDDM(raw []byte, z, x, y int, factor float32, noData []float32) (*tileData, error) {
	if len(raw)%4 != 0 {
		return nil, fmt.Errorf("ddm: payload not multiple of float32: %d", len(raw))
	}
	n := len(raw) / 4
	gs := int(math.Round(math.Sqrt(float64(n))))
	if gs*gs != n {
		return nil, fmt.Errorf("ddm: non-square grid: n=%d sqrt=%d", n, gs)
	}

	vals := make([]float32, n)
	// JS делает new Float32Array(arrayBuffer), в браузерах это LE.
	if err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, vals); err != nil {
		return nil, err
	}
	if factor != 1 {
		for i := range vals {
			vals[i] *= factor
		}
	}
	ns := make(map[float32]struct{}, len(noData))
	for _, v := range noData {
		ns[v*factor] = struct{}{}
	}
	return &tileData{
		Z:         z,
		X:         x,
		Y:         y,
		GridSize:  gs,
		Values:    vals,
		NoDataSet: ns,
		Factor:    factor,
	}, nil
}

func (t *tileData) at(i, j int) (float32, error) {
	gs := t.GridSize
	if i < 0 || j < 0 || i >= gs || j >= gs {
		return 0, errors.New("index oob")
	}
	return t.Values[i*gs+j], nil
}

// Интерполяция как в _getGroundHeightMerc, но проще: билинейная по узлам.
func (t *tileData) heightAtFrac(dx, dy float64) (float64, bool) {
	// dx,dy в [0..1) по тайлу
	gs := t.GridSize
	if gs < 2 {
		return 0, false
	}
	px := dx * float64(gs-1)
	py := dy * float64(gs-1)

	j := int(math.Floor(px))
	i := int(math.Floor(py))
	if j < 0 {
		j = 0
	}
	if i < 0 {
		i = 0
	}
	if j >= gs-1 {
		j = gs - 2
	}
	if i >= gs-1 {
		i = gs - 2
	}

	fx := float64(px) - float64(j)
	fy := float64(py) - float64(i)

	p00, _ := t.at(i, j)
	p10, _ := t.at(i, j+1)
	p01, _ := t.at(i+1, j)
	p11, _ := t.at(i+1, j+1)

	// noData обработка: усредняем валидные
	isND := func(v float32) bool {
		_, ok := t.NoDataSet[v]
		return ok
	}
	var sum float64
	var cnt int
	for _, v := range []float32{p00, p10, p01, p11} {
		if !isND(v) {
			sum += float64(v)
			cnt++
		}
	}
	if cnt == 0 {
		return 0, false
	}
	// если все валидны — билинейная
	if cnt == 4 {
		a := (1-fx)*float64(p00) + fx*float64(p10)
		b := (1-fx)*float64(p01) + fx*float64(p11)
		return (1-fy)*a + fy*b, true
	}
	// иначе — просто среднее валидных (как fallback)
	return sum / float64(cnt), true
}
