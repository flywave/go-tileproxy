package draw

import (
	"fmt"
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type interval struct {
	min, max float64
}

func intervalFromEndpoints(lo, hi float64) (min, max float64) {
	i := interval{lo, hi}
	if lo == -math.Pi && hi != math.Pi {
		i.min = math.Pi
	}
	if hi == -math.Pi && lo != math.Pi {
		i.max = math.Pi
	}
	return i.min, i.max
}

func CreateBBox(nwlat float64, nwlng float64, selat float64, selng float64) (*vec2d.Rect, error) {
	if nwlat < -90 || nwlat > 90 {
		return nil, fmt.Errorf("out of range nwlat (%f) must be in [-90, 90]", nwlat)
	}
	if nwlng < -180 || nwlng > 180 {
		return nil, fmt.Errorf("out of range nwlng (%f) must be in [-180, 180]", nwlng)
	}

	if selat < -90 || selat > 90 {
		return nil, fmt.Errorf("out of range selat (%f) must be in [-90, 90]", selat)
	}
	if selng < -180 || selng > 180 {
		return nil, fmt.Errorf("out of range selng (%f) must be in [-180, 180]", selng)
	}

	if nwlat == selat {
		return nil, fmt.Errorf("nwlat and selat must not be equal")
	}
	if nwlng == selng {
		return nil, fmt.Errorf("nwlng and selng must not be equal")
	}

	bbox := new(vec2d.Rect)
	if selat < nwlat {
		bbox.Min[0] = selat * math.Pi / 180.0
		bbox.Max[0] = nwlat * math.Pi / 180.0
	} else {
		bbox.Min[1] = nwlat * math.Pi / 180.0
		bbox.Max[1] = selat * math.Pi / 180.0
	}
	bbox.Min[1], bbox.Max[1] = intervalFromEndpoints(nwlng*math.Pi/180.0, selng*math.Pi/180.0)

	return bbox, nil
}
