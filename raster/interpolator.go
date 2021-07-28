package raster

type Interpolator interface {
	Interpolate(southWestHeight, southEastHeight, northWestHeight, northEastHeight, x, y float64) float64
}

type BilinearInterpolator struct {
	Interpolator
}

func Lerp(value1, value2, amount float64) float64 { return value1 + (value2-value1)*amount }

func (i *BilinearInterpolator) Interpolate(southWestHeight, southEastHeight, northWestHeight, northEastHeight, x, y float64) float64 {

	sw := southWestHeight
	se := southEastHeight
	nw := northWestHeight
	ne := northEastHeight

	//ha := Lerp(nw, sw, y)
	//hb := Lerp(ne, se, y)
	hi_linear := Lerp(Lerp(nw, sw, y), Lerp(ne, se, y), x)

	return hi_linear
}

type HyperbolicInterpolator struct {
	Interpolator
}

func (i *HyperbolicInterpolator) Interpolate(southWestHeight, southEastHeight, northWestHeight, northEastHeight, x, y float64) float64 {
	h1 := southWestHeight
	h2 := southEastHeight
	h3 := northWestHeight
	h4 := northEastHeight
	a00 := h1
	a10 := h2 - h1
	a01 := h3 - h1
	a11 := h1 - h2 - h3 + h4
	hi_hyperbolic := a00 + a10*x + a01*y + a11*x*y
	return hi_hyperbolic
}
