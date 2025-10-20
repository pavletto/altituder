package elevation

import "fmt"

type TileName struct {
	LatDeg int // целая часть южной/северной широты тайла
	LonDeg int // целая часть западной/восточной долготы тайла
	NS     byte
	EW     byte
}

func (t TileName) FileStem() string {
	// N/S + 2 цифры широты, E/W + 3 цифры долготы
	return fmt.Sprintf("%c%02d%c%03d", t.NS, t.LatDeg, t.EW, t.LonDeg)
}
