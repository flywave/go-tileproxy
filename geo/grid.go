package geo

import (
	"errors"
	"fmt"
	"math"
	"sort"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

var (
	geodetic_epsg_codes = []uint32{4326}
	MERC_BBOX           = vec2d.Rect{
		Min: vec2d.T{-20037508.342789244, -20037508.342789244},
		Max: vec2d.T{20037508.342789244, 20037508.342789244},
	}
	DEFAULT_EPSG_BBOX = map[uint32]vec2d.Rect{
		4326:   {Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}},
		900913: MERC_BBOX,
		3857:   MERC_BBOX,
		102100: MERC_BBOX,
		102113: MERC_BBOX,
	}
	DEFAULT_SRS_BBOX = map[string]vec2d.Rect{
		"EPSG:4326":   {Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}},
		"EPSG:900913": MERC_BBOX,
		"EPSG:3857":   MERC_BBOX,
		"EPSG:102100": MERC_BBOX,
		"EPSG:102113": MERC_BBOX,
	}
)

func GetResolution(bbox vec2d.Rect, size [2]uint32) float64 {
	w := math.Abs(bbox.Min[0] - bbox.Max[0])
	h := math.Abs(bbox.Min[1] - bbox.Max[1])
	return math.Min(w/float64(size[0]), h/float64(size[1]))
}

func pyramidResLevel(initial_res float64, factor *float32, levels *uint32) []float64 {
	nlevel := 20
	if levels != nil {
		nlevel = int(*levels)
	}
	fac := 2.0
	if factor != nil {
		fac = float64(*factor)
	}
	ret := make([]float64, nlevel)
	for i := range ret {
		ret[i] = math.Pow(initial_res/fac, float64(i))
	}
	return ret
}

type OriginType uint32

const (
	ORIGIN_UL OriginType = 0
	ORIGIN_LL OriginType = 1
)

func alignedResolutions(min_res *float64, max_res *float64, res_factor interface{}, num_levels *int,
	bbox *vec2d.Rect, tile_size []uint32, align_with *TileGrid) []float64 {
	alinged_res := align_with.Resolutions
	res := make([]float64, 0, len(alinged_res))

	var width, height, cmin_res float64
	if min_res == nil {
		if bbox != nil {
			width = bbox.Max[0] - bbox.Min[0]
			height = bbox.Max[1] - bbox.Min[1]
			cmin_res = math.Max(width/float64(tile_size[0]), height/float64(tile_size[1]))
		}
	}

	for i := 0; i < len(alinged_res); i++ {
		if alinged_res[i] <= cmin_res {
			if max_res != nil {
				if alinged_res[i] >= *max_res {
					res = append(res, alinged_res[i])
				}
			} else {
				res = append(res, alinged_res[i])
			}
		}
	}

	if num_levels != nil {
		res = res[:*num_levels]
	}

	factor_calculated := res[0] / res[1]

	switch fac := res_factor.(type) {
	case string:
		if fac == "sqrt2" && round(factor_calculated, 8, closest) != round(math.Sqrt(2), 8, closest) {
			if round(factor_calculated, 8, closest) == 2.0 {
				new_res := make([]float64, 0, len(res)*2)
				for _, r := range res {
					new_res = append(new_res, r)
					new_res = append(new_res, r/math.Sqrt(2))
				}
				res = new_res
			}
		}
	case float64:
		if fac == 2.0 && round(factor_calculated, 8, closest) != round(2.0, 8, closest) {
			if round(factor_calculated, 8, closest) == round(math.Sqrt(2), 8, closest) {
				new_res := make([]float64, 0, len(alinged_res)/2)
				for i := 0; i < len(res); i += 2 {
					new_res = append(new_res, res[i])
				}
				res = new_res
			}
		}
	}

	return res
}

func caclResolutions(min_res *float64, max_res *float64, res_factor interface{}, num_levels *int,
	bbox *vec2d.Rect, tile_size []uint32) []float64 {
	factor := 2.0
	if res_factor != nil {
		if str, ok := res_factor.(string); ok {
			if str == "sqrt2" {
				factor = math.Sqrt(2)
			}
		} else if fac, ok := res_factor.(float64); ok {
			factor = fac
		}
	}
	var tileSize []uint32
	if tile_size == nil {
		tileSize = []uint32{256, 256}
	} else {
		tileSize = tile_size[:]
	}

	res := make([]float64, 0, 20)

	var width, height, cmin_res float64
	if min_res == nil {
		if bbox != nil {
			width = bbox.Max[0] - bbox.Min[0]
			height = bbox.Max[1] - bbox.Min[1]
			cmin_res = math.Max(width/float64(tileSize[0]), height/float64(tileSize[1]))
		}
	} else {
		cmin_res = *min_res
	}

	if max_res != nil {
		if num_levels != nil {
			res_step := (math.Log10(*min_res) - math.Log10(*max_res)) / float64(*num_levels-1)
			for i := 0; i < *num_levels; i++ {
				res = append(res, math.Pow(10, (math.Log10(*min_res)-res_step*float64(i))))
			}
		} else {
			res = []float64{cmin_res}
			for {
				next_res := res[len(res)-1] / factor
				if *max_res >= next_res {
					break
				}
				res = append(res, next_res)
			}
		}
	} else {
		var cnum_levels int
		if num_levels == nil {
			if factor != math.Sqrt(2) {
				cnum_levels = 20
			} else {
				cnum_levels = 40
			}
		} else {
			cnum_levels = *num_levels
		}
		res = []float64{cmin_res}
		for len(res) < cnum_levels {
			res = append(res, res[len(res)-1]/factor)
		}
	}

	return res
}

func GridBBox(bbox vec2d.Rect, bbox_srs *SRSProj4, srs *SRSProj4) vec2d.Rect {
	if bbox_srs != nil {
		bbox = bbox_srs.TransformRectTo(srs, bbox, 16)
	}
	return bbox
}

func BBoxWidth(bbox vec2d.Rect) float64 {
	return bbox.Max[0] - bbox.Min[0]
}

func BBoxHeight(bbox vec2d.Rect) float64 {
	return bbox.Max[1] - bbox.Min[1]
}

func BBoxSize(bbox vec2d.Rect) (width, height float64) {
	return BBoxWidth(bbox), BBoxHeight(bbox)
}

type TileGrid struct {
	Name            string
	IsGeodetic      bool
	Srs             *SRSProj4
	TileSize        []uint32
	Origin          OriginType
	FlippedYAxis    bool
	StretchFactor   float64
	MaxShrinkFactor float64
	Levels          uint32
	BBox            *vec2d.Rect
	Resolutions     []float64
	ThresholdRes    []float64
	GridSizes       [][2]uint32
	SpheroidA       float64
}

type TileGridOptions map[string]interface{}

const (
	TILEGRID_SRS                = "srs"
	TILEGRID_BBOX               = "bbox"
	TILEGRID_BBOX_SRS           = "bbox_srs"
	TILEGRID_TILE_SIZE          = "tile_size"
	TILEGRID_RES                = "res"
	TILEGRID_RES_FACTOR         = "res_factor"
	TILEGRID_THRESHOLD_RES      = "threshold_res"
	TILEGRID_NUM_LEVELS         = "num_levels"
	TILEGRID_MIN_RES            = "min_res"
	TILEGRID_MAX_RES            = "max_res"
	TILEGRID_MAX_STRETCH_FACTOR = "stretch_factor"
	TILEGRID_MAX_SHRINK_FACTOR  = "max_shrink_factor"
	TILEGRID_ALIGN_WITH         = "align_with"
	TILEGRID_ORIGIN             = "origin"
	TILEGRID_NAME               = "name"
	TILEGRID_IS_GEODETIC        = "is_geodetic"
)

func DefaultTileGridOptions() TileGridOptions {
	conf := make(TileGridOptions)
	conf[TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[TILEGRID_IS_GEODETIC] = false
	conf[TILEGRID_MAX_STRETCH_FACTOR] = 1.15
	conf[TILEGRID_MAX_SHRINK_FACTOR] = 4.0
	conf[TILEGRID_ORIGIN] = ORIGIN_LL
	return conf
}

func NewTileGrid(options TileGridOptions) *TileGrid {
	var srs *SRSProj4
	if v, ok := options[TILEGRID_SRS]; ok {
		switch sv := v.(type) {
		case string:
			srs = NewSRSProj4(sv)
		case *SRSProj4:
			srs = sv
		}
	}
	var bbox *vec2d.Rect
	if v, ok := options[TILEGRID_BBOX]; ok {
		bbox = NewRect(v.(vec2d.Rect))
	}
	var bbox_srs *SRSProj4
	if v, ok := options[TILEGRID_BBOX_SRS]; ok {
		switch sv := v.(type) {
		case string:
			bbox_srs = NewSRSProj4(sv)
		case *SRSProj4:
			bbox_srs = sv
		}
	}
	var tile_size []uint32
	if v, ok := options[TILEGRID_TILE_SIZE]; ok {
		tile_size = v.([]uint32)
	}
	var res []float64
	if v, ok := options[TILEGRID_RES]; ok {
		res = v.([]float64)
	}
	var res_factor interface{}
	if v, ok := options[TILEGRID_RES_FACTOR]; ok {
		res_factor = v
	}
	var threshold_res []float64
	if v, ok := options[TILEGRID_THRESHOLD_RES]; ok {
		threshold_res = v.([]float64)
	}
	var num_levels *int
	if v, ok := options[TILEGRID_NUM_LEVELS]; ok {
		num_levels = NewInt(v.(int))
	}
	var min_res *float64
	if v, ok := options[TILEGRID_MIN_RES]; ok {
		min_res = NewFloat64(v.(float64))
	}
	var max_res *float64
	if v, ok := options[TILEGRID_MAX_RES]; ok {
		max_res = NewFloat64(v.(float64))
	}
	var stretch_factor float64
	if v, ok := options[TILEGRID_MAX_STRETCH_FACTOR]; ok {
		stretch_factor = v.(float64)
	} else {
		stretch_factor = 1.15
	}
	var max_shrink_factor float64
	if v, ok := options[TILEGRID_MAX_SHRINK_FACTOR]; ok {
		max_shrink_factor = v.(float64)
	} else {
		max_shrink_factor = 4.0
	}
	var align_with *TileGrid
	if v, ok := options[TILEGRID_ALIGN_WITH]; ok {
		align_with = v.(*TileGrid)
	}
	var origin OriginType
	if v, ok := options[TILEGRID_ORIGIN]; ok {
		origin = v.(OriginType)
	} else {
		origin = ORIGIN_LL
	}
	var name string
	if v, ok := options[TILEGRID_NAME]; ok {
		name = v.(string)
	}
	var is_geodetic bool
	if v, ok := options[TILEGRID_IS_GEODETIC]; ok {
		is_geodetic = v.(bool)
	} else {
		is_geodetic = false
	}

	if srs == nil {
		srs = NewSRSProj4("EPSG:900913")
	}

	if tile_size == nil {
		tile_size = []uint32{256, 256}
	}

	var cbbox vec2d.Rect

	if bbox == nil {
		var ok bool
		cbbox, ok = DEFAULT_SRS_BBOX[srs.SrsCode]
		if !ok {
			return nil
		} else {
			bbox = &cbbox
		}
	}

	cbbox = GridBBox(cbbox, bbox_srs, srs)

	if res != nil {
		sort.Sort(sort.Reverse(sort.Float64Slice(res)))
	} else if align_with != nil {
		res = alignedResolutions(min_res, max_res, res_factor, num_levels, bbox, tile_size, align_with)
	} else {
		res = caclResolutions(min_res, max_res, res_factor, num_levels, bbox, tile_size)
	}

	return newTileGrid(name, is_geodetic, origin, srs, bbox, res_factor, tile_size, res, threshold_res, stretch_factor, max_shrink_factor)
}

func (t *TileGrid) calcGrids() [][2]uint32 {
	width := t.BBox.Max[0] - t.BBox.Min[0]
	height := t.BBox.Max[1] - t.BBox.Min[1]
	grids := make([][2]uint32, len(t.Resolutions))
	for i := range t.Resolutions {
		res := t.Resolutions[i]

		x := math.Max(math.Ceil(math.Floor(width/res)/float64(t.TileSize[0])), 1)
		y := math.Max(math.Ceil(math.Floor(height/res)/float64(t.TileSize[1])), 1)

		grids[i] = [2]uint32{uint32(x), uint32(y)}
	}
	return grids
}

func (t *TileGrid) calcRes(factor *float32) []float64 {
	width := t.BBox.Max[0] - t.BBox.Min[0]
	height := t.BBox.Max[1] - t.BBox.Min[1]
	initial_res := math.Max(width/float64(t.TileSize[0]), height/float64(t.TileSize[1]))
	if factor == nil {
		return pyramidResLevel(initial_res, nil, &t.Levels)
	} else {
		return pyramidResLevel(initial_res, factor, &t.Levels)
	}
}

func (t *TileGrid) calcBBox() vec2d.Rect {
	if t.IsGeodetic {
		return vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	} else {
		circum := 2 * math.Pi * t.SpheroidA
		offset := circum / 2.0
		return vec2d.Rect{Min: vec2d.T{-offset, -offset}, Max: vec2d.T{offset, offset}}
	}
}

func newTileGrid(name string, is_geodetic bool, origin OriginType, srs *SRSProj4, bbox *vec2d.Rect, res_factor interface{}, tile_size []uint32, res []float64, threshold_res []float64, stretch_factor float64, max_shrink_factor float64) *TileGrid {
	ret := &TileGrid{}
	ret.Name = name
	ret.SpheroidA = 6378137.0
	ret.Srs = srs
	ret.BBox = bbox
	ret.TileSize = tile_size
	ret.IsGeodetic = is_geodetic
	ret.StretchFactor = stretch_factor
	ret.MaxShrinkFactor = max_shrink_factor
	ret.Origin = origin
	if ret.Origin == ORIGIN_UL {
		ret.FlippedYAxis = true
	} else {
		ret.FlippedYAxis = false
	}
	if ret.BBox == nil {
		ret.calcBBox()
	}
	if res != nil {
		ret.Resolutions = res
	} else {
		var fac float32
		if res_factor == nil {
			ret.Levels = 20
			fac = 2.0
			ret.Resolutions = ret.calcRes(&fac)
		} else if str, ok := res_factor.(string); ok {
			if str == "sqrt2" {
				ret.Levels = 40
				fac = float32(math.Sqrt(2))
				ret.Resolutions = ret.calcRes(&fac)
			}
		} else if ffac, ok := res_factor.(float32); ok {
			fac = ffac
			ret.Resolutions = ret.calcRes(&fac)
		} else if ffac, ok := res_factor.(float64); ok {
			fac = float32(ffac)
			ret.Resolutions = ret.calcRes(&fac)
		}
	}

	if threshold_res != nil {
		ret.ThresholdRes = threshold_res[:]
	}

	ret.Levels = uint32(len(ret.Resolutions))
	ret.GridSizes = ret.calcGrids()
	return ret
}

func NewDefaultTileGrid() *TileGrid {
	ret := &TileGrid{}
	ret.SpheroidA = 6378137.0
	ret.Srs = NewSRSProj4("900913")
	ret.BBox = nil
	ret.TileSize = []uint32{256, 256}
	ret.IsGeodetic = false
	ret.StretchFactor = 1.15
	ret.MaxShrinkFactor = 4.0
	ret.Origin = ORIGIN_LL
	ret.FlippedYAxis = false
	ret.Levels = 20
	ret.calcBBox()
	fac := float32(2.0)
	ret.Resolutions = ret.calcRes(&fac)
	ret.Levels = uint32(len(ret.Resolutions))
	ret.GridSizes = ret.calcGrids()
	return ret
}

func (t *TileGrid) Resolution(level int) float64 {
	if level >= int(t.Levels) {
		return 0
	}
	return t.Resolutions[level]
}

func (t *TileGrid) ClosestLevel(res float64) int {
	prev_l_res := t.Resolutions[0]
	threshold := float64(-1)
	thresholds := make([]float64, 0)
	if t.ThresholdRes != nil {
		thresholds = t.ThresholdRes[:]
		threshold = thresholds[len(thresholds)-1]
		thresholds = thresholds[:len(thresholds)-1]
		for threshold > prev_l_res && len(thresholds) > 0 {
			threshold = thresholds[len(thresholds)-1]
			thresholds = thresholds[:len(thresholds)-1]
		}
	}

	threshold_result := int(-1)
	var level int
	for level = range t.Resolutions {
		l_res := t.Resolutions[level]
		if threshold >= 0 && prev_l_res > threshold && prev_l_res >= l_res {
			if res > threshold {
				return level - 1
			} else if res >= l_res {
				return level
			}
			if len(thresholds) > 0 {
				threshold = thresholds[len(thresholds)-1]
				thresholds = thresholds[:len(thresholds)-1]
			} else {
				threshold = -1
			}
		}

		if threshold_result >= 0 {
			if l_res < res {
				return threshold_result
			}
		}

		if l_res <= res*t.StretchFactor {
			threshold_result = level
		}
		prev_l_res = l_res
	}
	return level
}

func (t *TileGrid) Tile(x, y, level int) (cx, cy, clevel int) {
	res := t.Resolution(level)
	var fx, fy float64
	fx = float64(x) - t.BBox.Min[0]
	if t.FlippedYAxis {
		fy = t.BBox.Max[1] - float64(y)
	} else {
		fy = float64(y) - t.BBox.Min[1]
	}
	tile_x := fx / (res * float64(t.TileSize[0]))
	tile_y := fy / (res * float64(t.TileSize[1]))
	return int(math.Floor(tile_x)), int(math.Floor(tile_y)), level
}

func (t *TileGrid) FlipTileCoord(x, y, level int) (cx, xy, clevel int) {
	return x, int(t.GridSizes[level][1]) - 1 - y, level
}

func (t *TileGrid) SupportsAccessWithOrigin(origin OriginType) bool {
	if t.Origin == origin {
		return true
	}

	delta := math.Max(math.Abs(t.BBox.Min[1]), math.Abs(t.BBox.Max[1])) / 1e12

	for level := range t.GridSizes {
		tiles := [][3]int{{0, 0, level},
			{int(t.GridSizes[level][0]) - 1, int(t.GridSizes[level][1]) - 1, level}}
		level_bbox := t.tilesBBox(tiles)

		if math.Abs(t.BBox.Min[1]-level_bbox.Min[1]) > delta || math.Abs(t.BBox.Max[1]-level_bbox.Max[1]) > delta {
			return false
		}
	}
	return true
}

func (t *TileGrid) OriginTile(level int, origin OriginType) (cx, cy, clevel int) {
	if t.SupportsAccessWithOrigin(origin) {
		panic("tile origins are incompatible")
	}
	cx, cy, clevel = 0, 0, level

	if t.Origin == origin {
		return
	}

	return t.FlipTileCoord(cx, cy, clevel)
}

func (t *TileGrid) GetAffectedTiles(bbox vec2d.Rect, size [2]uint32, req_srs *SRSProj4) (vec2d.Rect, [2]int, *TileIter, error) {
	src_bbox, level, err := t.GetAffectedBBoxAndLevel(bbox, size, req_srs)
	if err != nil {
		return src_bbox, [2]int{}, nil, err
	}
	return t.GetAffectedLevelTiles(src_bbox, level)
}

func (t *TileGrid) GetAffectedBBoxAndLevel(bbox vec2d.Rect, size [2]uint32, req_srs *SRSProj4) (vec2d.Rect, int, error) {
	var src_bbox vec2d.Rect
	if req_srs != nil && !req_srs.Eq(t.Srs) {
		src_bbox = req_srs.TransformRectTo(t.Srs, bbox, 16)
	} else {
		src_bbox = bbox
	}

	if !BBoxIntersects(*t.BBox, src_bbox) {
		return vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}, -1, errors.New("no tiles")
	}

	res := GetResolution(src_bbox, size)
	level := t.ClosestLevel(res)

	if res > t.Resolutions[0]*t.MaxShrinkFactor {
		return vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}, -1, errors.New("no tiles")
	}

	return src_bbox, level, nil
}

func (t *TileGrid) GetAffectedLevelTiles(bbox vec2d.Rect, level int) (vec2d.Rect, [2]int, *TileIter, error) {
	delta := t.Resolutions[level] / 10.0
	x0, y0, _ := t.Tile(int(bbox.Min[0]+delta), int(bbox.Min[1]+delta), level)
	x1, y1, _ := t.Tile(int(bbox.Max[0]-delta), int(bbox.Max[1]-delta), level)
	return t.tileIter(x0, y0, x1, y1, level)
}

type TileIter struct {
	grid_size [2]uint32
	level     int
	xs        []int
	ys        []int
	x_off     int
	y_off     int
}

func createTileTileIter(xs, ys []int, level int, grid_size [2]uint32) *TileIter {
	ret := &TileIter{grid_size: grid_size, level: level, xs: xs, ys: ys, x_off: 0, y_off: 0}
	return ret
}

func (i *TileIter) Reset() {
	i.x_off = 0
	i.y_off = 0
}

func (i *TileIter) Next() (x, y, level int, done bool) {
	x = i.xs[i.x_off]
	y = i.ys[i.y_off]
	level = i.level
	if i.x_off < len(i.xs)-1 {
		i.x_off++
	} else {
		i.x_off = 0
		i.y_off++
	}

	if i.y_off >= len(i.ys) {
		done = true
		i.y_off = 0
		return
	} else {
		done = false
	}

	return
}

func (t *TileGrid) tileIter(x0, y0, x1, y1, level int) (vec2d.Rect, [2]int, *TileIter, error) {
	xs := make([]int, 0)
	for x := x0; x <= x1; x++ {
		xs = append(xs, x)
	}
	ys := make([]int, 0)
	if t.FlippedYAxis {
		y0, y1 = y1, y0
		for y := y0; y <= y1; y++ {
			ys = append(ys, y)
		}
	} else {
		for y := y1; y >= y0; y-- {
			ys = append(ys, y)
		}
	}
	ll := [3]int{xs[0], ys[len(ys)-1], level}
	ur := [3]int{xs[len(xs)-1], ys[0], level}

	abbox := t.tilesBBox([][3]int{ll, ur})
	return abbox, [2]int{len(xs), len(ys)},
		createTileTileIter(xs, ys, level, t.GridSizes[level]), nil
}

func (t *TileGrid) tilesBBox(tiles [][3]int) vec2d.Rect {
	ll_bbox := t.TileBBox(tiles[0], false)
	ur_bbox := t.TileBBox(tiles[len(tiles)-1], false)
	return MergeBBox(ll_bbox, ur_bbox)
}

func (t *TileGrid) TileBBox(tile_coord [3]int, limit bool) vec2d.Rect {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]
	res := t.Resolution(z)

	x0 := t.BBox.Min[0] + round(float64(x)*res*float64(t.TileSize[0]), 12, closest)
	x1 := x0 + round(res*float64(t.TileSize[0]), 12, closest)
	var y1, y0 float64
	if t.FlippedYAxis {
		y1 = t.BBox.Max[1] - round(float64(y)*res*float64(t.TileSize[1]), 12, closest)
		y0 = y1 - round(res*float64(t.TileSize[1]), 12, closest)
	} else {
		y0 = t.BBox.Min[1] + round(float64(y)*res*float64(t.TileSize[1]), 12, closest)
		y1 = y0 + round(res*float64(t.TileSize[1]), 12, closest)
	}

	if limit {
		return vec2d.Rect{Min: vec2d.T{
			math.Max(x0, t.BBox.Min[0]),
			math.Max(y0, t.BBox.Min[1])},
			Max: vec2d.T{math.Min(x1, t.BBox.Max[0]),
				math.Min(y1, t.BBox.Max[1])},
		}
	}
	return vec2d.Rect{Min: vec2d.T{x0, y0}, Max: vec2d.T{x1, y1}}
}

func (t *TileGrid) LimitTile(tile_coord [3]int) []int {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]
	if z < 0 || z >= int(t.Levels) {
		return nil
	}
	grid := t.GridSizes[z]
	if x < 0 || y < 0 || x >= int(grid[0]) || y >= int(grid[1]) {
		return nil
	}
	return []int{x, y, z}
}

func (t *TileGrid) ToString() string {
	return fmt.Sprintf("%s(%s, (%.4f, %.4f, %.4f, %.4f)", t.Name, t.Srs.ToString(), t.BBox.Min[0], t.BBox.Min[1], t.BBox.Max[0], t.BBox.Max[1])
}

func (t *TileGrid) isSubsetOf(other *TileGrid) bool {
	if !t.Srs.Eq(other.Srs) {
		return false
	}

	if t.TileSize != nil && other.TileSize != nil && (t.TileSize[0] != other.TileSize[0] || t.TileSize[1] != other.TileSize[1]) {
		return false
	}

	for self_level := range t.Resolutions {
		level_size := [2]uint32{
			t.GridSizes[self_level][0] * t.TileSize[0],
			t.GridSizes[self_level][1] * t.TileSize[1],
		}
		level_bbox := t.tilesBBox([][3]int{
			{0, 0, self_level},
			{int(t.GridSizes[self_level][0]) - 1, int(t.GridSizes[self_level][1]) - 1, self_level},
		})

		bbox, level, err := other.GetAffectedBBoxAndLevel(level_bbox, level_size, nil)
		if err != nil {
			return false
		}

		bbox, _, _, err = other.GetAffectedLevelTiles(level_bbox, level)
		if err != nil {
			return false
		}

		if other.Resolution(level) != t.Resolutions[self_level] {
			return false
		}
		if !BBoxEquals(bbox, level_bbox, math.Inf(1), math.Inf(1)) {
			return false
		}
	}
	return true
}

func TileGridForEpsg(SrsCode string, bbox *vec2d.Rect, tile_size []uint32, res []float64) *TileGrid {
	epsg := getEpsgNum(SrsCode)
	for c := range geodetic_epsg_codes {
		if c == int(epsg) {
			srs := NewSRSProj4(fmt.Sprintf("EPSG:%d", epsg))
			conf := DefaultTileGridOptions()
			conf[TILEGRID_SRS] = srs
			conf[TILEGRID_BBOX] = bbox
			conf[TILEGRID_TILE_SIZE] = tile_size
			conf[TILEGRID_RES] = res
			tg := NewTileGrid(conf)
			tg.IsGeodetic = true
			return tg
		}
	}
	srs := NewSRSProj4(fmt.Sprintf("EPSG:%d", epsg))
	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = srs
	conf[TILEGRID_BBOX] = bbox
	conf[TILEGRID_TILE_SIZE] = tile_size
	conf[TILEGRID_RES] = res
	return NewTileGrid(conf)
}

type MetaGrid struct {
	TileGrid
	MetaSize   [2]uint32
	MetaBuffer int
}

func NewMetaGrid(grid *TileGrid, metaSize [2]uint32, metaBuffer int) *MetaGrid {
	return &MetaGrid{TileGrid: *grid, MetaSize: metaSize, MetaBuffer: metaBuffer}
}

func (g *MetaGrid) metaBBox(tile_coord *[3]int, tiles [][3]int, limit_to_bbox bool) (vec2d.Rect, []int) {
	var level int
	var bbox vec2d.Rect
	if tiles != nil {
		level = tiles[0][2]
		bbox = g.tilesBBox(tiles)
	} else {
		level = tile_coord[2]
		bbox = g.unbufferedMetaBBox(*tile_coord)
	}
	return g.bufferedBBox(bbox, level, limit_to_bbox)
}

func (g *MetaGrid) unbufferedMetaBBox(tile_coord [3]int) vec2d.Rect {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]

	meta_size := g.metaSize(z)

	return g.tilesBBox([][3]int{tile_coord, {x + int(meta_size[0]) - 1, y + int(meta_size[1]) - 1, z}})
}

func (g *MetaGrid) bufferedBBox(bbox vec2d.Rect, level int, limit_to_grid_bbox bool) (vec2d.Rect, []int) {
	minx, miny, maxx, maxy := bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1]

	buffers := []int{0, 0, 0, 0}
	if g.MetaBuffer > 0 {
		res := g.Resolution(level)
		minx -= float64(g.MetaBuffer) * res
		miny -= float64(g.MetaBuffer) * res
		maxx += float64(g.MetaBuffer) * res
		maxy += float64(g.MetaBuffer) * res
		buffers = []int{g.MetaBuffer, g.MetaBuffer, g.MetaBuffer, g.MetaBuffer}

		if limit_to_grid_bbox {
			if g.BBox.Min[0] > minx {
				delta := g.BBox.Min[0] - minx
				buffers[0] = buffers[0] - int(round(delta/res, 5, closest))
				minx = g.BBox.Min[0]
			}
			if g.BBox.Min[1] > miny {
				delta := g.BBox.Min[1] - miny
				buffers[1] = buffers[1] - int(round(delta/res, 5, closest))
				miny = g.BBox.Min[1]
			}
			if g.BBox.Max[0] < maxx {
				delta := maxx - g.BBox.Max[0]
				buffers[2] = buffers[2] - int(round(delta/res, 5, closest))
				maxx = g.BBox.Max[0]
			}
			if g.BBox.Max[1] < maxy {
				delta := maxy - g.BBox.Max[1]
				buffers[3] = buffers[3] - int(round(delta/res, 5, closest))
				maxy = g.BBox.Max[1]
			}
		}
	}
	return vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}, buffers
}

func (g *MetaGrid) GetMetaTile(tile_coord [3]int) *MetaTile {
	tile_coord = g.MainTile(tile_coord)
	level := tile_coord[2]
	bbox, buffers := g.metaBBox(&tile_coord, nil, true)
	grid_size := g.metaSize(level)
	size := g.sizeFromBufferedBBox(bbox, level)

	tile_patterns := g.tilesPattern(grid_size, buffers, &tile_coord, nil)

	return NewMetaTile(bbox, size, tile_patterns, grid_size)
}

func (g *MetaGrid) minimalMetaTile(tiles [][3]int) *MetaTile {
	tiles, grid_size, bounds := g.fullTileList(tiles)
	bbox, buffers := g.metaBBox(nil, bounds, true)

	level := tiles[0][2]
	size := g.sizeFromBufferedBBox(bbox, level)

	tile_pattern := g.tilesPattern(grid_size, buffers, nil, tiles)

	return NewMetaTile(bbox, size, tile_pattern, grid_size)
}

func (g *MetaGrid) sizeFromBufferedBBox(bbox vec2d.Rect, level int) [2]uint32 {
	res := g.Resolution(level)
	width := int(math.Round((bbox.Max[0] - bbox.Min[0]) / res))
	height := int(math.Round((bbox.Max[1] - bbox.Min[1]) / res))
	return [2]uint32{uint32(width), uint32(height)}
}

func (g *MetaGrid) fullTileList(tiles [][3]int) ([][3]int, [2]uint32, [][3]int) {
	tile := tiles[len(tiles)-1]
	tiles = tiles[:len(tiles)-1]
	z := tile[2]
	minx := tile[0]
	maxx := tile[0]
	miny := tile[1]
	maxy := tile[1]

	for _, tile := range tiles {
		x, y := tile[0], tile[1]
		minx = MinInt(minx, x)
		maxx = MaxInt(maxx, x)
		miny = MinInt(miny, y)
		maxy = MaxInt(maxy, y)
	}

	grid_size := [2]uint32{uint32(1 + maxx - minx), uint32(1 + maxy - miny)}
	xs := make([]int, 0)
	ys := make([]int, 0)
	if g.FlippedYAxis {
		for y := miny; y <= maxy; y++ {
			ys = append(ys, y)
		}
	} else {
		for y := maxy; y >= miny; y-- {
			ys = append(ys, y)
		}
	}
	for x := minx; x <= maxx; x++ {
		xs = append(xs, x)
	}

	bounds := [][3]int{{minx, miny, z}, {maxx, maxy, z}}

	return createTileList(xs, ys, z, [2]uint32{uint32(maxx + 1), uint32(maxy + 1)}), grid_size, bounds
}

func (g *MetaGrid) MainTile(tile_coord [3]int) [3]int {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]

	meta_size := g.metaSize(z)

	x0 := int(math.Floor(float64(x)/float64(meta_size[0])) * float64(meta_size[0]))
	y0 := int(math.Floor(float64(y)/float64(meta_size[1])) * float64(meta_size[1]))

	return [3]int{x0, y0, z}
}

func (g *MetaGrid) TileList(main_tile [3]int) [][3]int {
	tile_grid := g.metaSize(main_tile[2])
	return g.metaTileList(main_tile, tile_grid)
}

func (g *MetaGrid) metaTileList(main_tile [3]int, tile_grid [2]uint32) [][3]int {
	t := g.MainTile(main_tile)
	minx, miny, z := t[0], t[1], t[2]
	maxx := minx + int(tile_grid[0]) - 1
	maxy := miny + int(tile_grid[1]) - 1

	xs := make([]int, 0)
	ys := make([]int, 0)

	if g.FlippedYAxis {
		for y := miny; y <= maxy; y++ {
			ys = append(ys, y)
		}
	} else {
		for y := maxy; y >= miny; y-- {
			ys = append(ys, y)
		}
	}
	for x := minx; x <= maxx; x++ {
		xs = append(xs, x)
	}
	return createTileList(xs, ys, z, g.GridSizes[z])
}

func createTileList(xs, ys []int, level int, grid_size [2]uint32) [][3]int {
	ret := make([][3]int, 0)
	x_limit := int(grid_size[0])
	y_limit := int(grid_size[1])
	for _, y := range ys {
		for _, x := range xs {
			if x < 0 || y < 0 || x >= x_limit || y >= y_limit {
				continue
			} else {
				ret = append(ret, [3]int{x, y, level})
			}
		}
	}
	return ret
}

type TilePattern struct {
	Tiles [3]int
	Sizes [2]int
}

func (g *MetaGrid) tilesPattern(grid_size [2]uint32, buffers []int, tile *[3]int, tiles [][3]int) []TilePattern {
	if tile != nil {
		tiles = g.metaTileList(*tile, grid_size)
	}
	ret := make([]TilePattern, 0)
	for i := 0; i < int(grid_size[1]); i++ {
		for j := 0; j < int(grid_size[0]); j++ {
			ret = append(ret, TilePattern{Tiles: tiles[j+i*int(grid_size[0])],
				Sizes: [2]int{j*int(g.TileSize[0]) + buffers[0], i*int(g.TileSize[1]) + buffers[3]}})
		}
	}
	return ret
}

func (g *MetaGrid) metaSize(level int) [2]uint32 {
	grid_size := g.GridSizes[level]
	return [2]uint32{MinUInt32(g.MetaSize[0], grid_size[0]), MinUInt32(g.MetaSize[1], grid_size[1])}
}

func (g *MetaGrid) GetAffectedLevelTiles(bbox vec2d.Rect, level int) (vec2d.Rect, [2]int, [][3]int) {
	delta := g.Resolutions[level] / 10.0
	x0, y0, _ := g.Tile(int(bbox.Min[0]+delta), int(bbox.Min[1]+delta), level)
	x1, y1, _ := g.Tile(int(bbox.Max[0]-delta), int(bbox.Max[1]-delta), level)

	meta_size := g.metaSize(level)

	x0 = int(math.Floor(float64(x0)/float64(meta_size[0])) * float64(meta_size[0]))
	x1 = int(math.Floor(float64(x1)/float64(meta_size[0])) * float64(meta_size[0]))
	y0 = int(math.Floor(float64(y0)/float64(meta_size[1])) * float64(meta_size[1]))
	y1 = int(math.Floor(float64(y1)/float64(meta_size[1])) * float64(meta_size[1]))

	return g.tileList(x0, y0, x1, y1, level)
}

func (g *MetaGrid) tileList(x0, y0, x1, y1, level int) (vec2d.Rect, [2]int, [][3]int) {
	meta_size := g.metaSize(level)

	xs := make([]int, 0)
	ys := make([]int, 0)

	for x := x0; x < x1+1; x += int(meta_size[0]) {
		xs = append(xs, x)
	}

	if g.FlippedYAxis {
		y0, y1 = y1, y0
		for y := y0; y < y1+1; y += int(meta_size[1]) {
			ys = append(ys, y)
		}
	} else {
		for y := y1; y > y0-1; y -= int(meta_size[1]) {
			ys = append(ys, y)
		}
	}

	ll := [3]int{xs[0], ys[len(ys)-1], level}
	ur := [3]int{xs[len(xs)-1], ys[0], level}

	ur = [3]int{ur[0] + int(meta_size[0]) - 1, ur[1] + int(meta_size[1]) - 1, ur[2]}
	abbox := g.tilesBBox([][3]int{ll, ur})
	return abbox, [2]int{len(xs), len(ys)}, createTileList(xs, ys, level, g.GridSizes[level])
}

type MetaTile struct {
	bbox          vec2d.Rect
	size          [2]uint32
	tile_patterns []TilePattern
	grid_size     [2]uint32
}

func (t *MetaTile) GetBBox() vec2d.Rect {
	return t.bbox
}

func NewMetaTile(bbox vec2d.Rect, size [2]uint32, tiles []TilePattern, grid_size [2]uint32) *MetaTile {
	return &MetaTile{bbox: bbox, size: size, tile_patterns: tiles, grid_size: grid_size}
}

func BBoxIntersects(one, two vec2d.Rect) bool {
	a_x0, a_y0, a_x1, a_y1 := one.Min[0], one.Min[1], one.Max[0], one.Max[1]
	b_x0, b_y0, b_x1, b_y1 := two.Min[0], two.Min[1], two.Max[0], two.Max[1]

	if a_x0 < b_x1 &&
		a_x1 > b_x0 &&
		a_y0 < b_y1 &&
		a_y1 > b_y0 {
		return true
	}

	return false
}

func BBoxContains(one, two vec2d.Rect) bool {
	a_x0, a_y0, a_x1, a_y1 := one.Min[0], one.Min[1], one.Max[0], one.Max[1]
	b_x0, b_y0, b_x1, b_y1 := two.Min[0], two.Min[1], two.Max[0], two.Max[1]

	x_delta := math.Abs(a_x1-a_x0) / 10e12
	y_delta := math.Abs(a_y1-a_y0) / 10e12

	if a_x0 <= b_x0+x_delta &&
		a_x1 >= b_x1-x_delta &&
		a_y0 <= b_y0+y_delta &&
		a_y1 >= b_y1-y_delta {
		return true
	}
	return false
}

func deg_to_m(deg float64) float64 {
	return deg * (6378137 * 2 * math.Pi) / 360
}

const (
	OGC_PIXEL_SIZE = 0.00028
)

func ogc_scale_to_res(scale float64) float64 {
	return scale * OGC_PIXEL_SIZE
}

func res_to_ogc_scale(res float64) float64 {
	return res / OGC_PIXEL_SIZE
}

type ResolutionRange struct {
	Min *float64
	Max *float64
}

func NewResolutionRange(min_res *float64, max_res *float64) *ResolutionRange {
	return &ResolutionRange{Min: min_res, Max: max_res}
}

func (r *ResolutionRange) ScaleDenominator() (min_scale, max_scale float64) {
	if r.Min != nil {
		min_scale = res_to_ogc_scale(*r.Max)
	}
	if r.Max != nil {
		max_scale = res_to_ogc_scale(*r.Min)
	}
	return
}

func (r *ResolutionRange) ScaleHint() (float64, float64) {
	var min_res, max_res float64
	if r.Min != nil {
		min_res = *r.Min
	}
	if r.Max != nil {
		max_res = *r.Max
	}
	if min_res != 0 {
		min_res = math.Sqrt(math.Pow(2*min_res, 2))
	}
	if max_res != 0 {
		max_res = math.Sqrt(math.Pow(2*max_res, 2))
	}
	return min_res, max_res
}

func (r *ResolutionRange) Contains(bbox vec2d.Rect, size [2]uint32, srs Proj) bool {
	width, height := BBoxSize(bbox)
	if srs.IsLatLong() {
		width = deg_to_m(width)
		height = deg_to_m(height)
	}

	x_res := width / float64(size[0])
	y_res := height / float64(size[1])

	if r.Min != nil {
		min_res := *r.Min + 1e-6
		if min_res <= x_res || min_res <= y_res {
			return false
		}
	}
	if r.Max != nil {
		max_res := *r.Max
		if max_res > x_res || max_res > y_res {
			return false
		}
	}

	return true
}

func (r *ResolutionRange) Eq(o *ResolutionRange) bool {
	return (*r.Min == *o.Min && *r.Max == *o.Max)
}

func (r *ResolutionRange) ToString() string {
	var min_res, max_res float64
	if r.Min != nil {
		min_res = *r.Min
	} else {
		min_res = 9e99
	}
	if r.Max != nil {
		max_res = *r.Max
	} else {
		max_res = 0
	}
	return fmt.Sprintf("ResolutionRange(min_res=%.3f, max_res=%.3f)", min_res, max_res)
}

func CacleResolutionRange(min_res *float64, max_res *float64, max_scale *float64, min_scale *float64) *ResolutionRange {
	if min_scale == nil && max_scale == nil && min_res == nil && max_res == nil {
		return nil
	}
	if min_res != nil || max_res != nil {
		if max_scale == nil && min_scale == nil {
			return NewResolutionRange(min_res, max_res)
		}
	} else if max_scale != nil && min_scale != nil {
		if min_res == nil && max_res == nil {
			cmin_res := ogc_scale_to_res(*max_scale)
			cmax_res := ogc_scale_to_res(*min_scale)
			return NewResolutionRange(&cmin_res, &cmax_res)
		}
	}
	return nil
}

func maxWithNone(a, b *float64) *float64 {
	if a == nil || b == nil {
		return nil
	} else {
		ret := math.Max(*a, *b)
		return &ret
	}
}

func minWithNone(a, b *float64) *float64 {
	if a == nil || b == nil {
		return nil
	} else {
		ret := math.Min(*a, *b)
		return &ret
	}
}

func MergeResolutionRange(a, b *ResolutionRange) *ResolutionRange {
	if a != nil && b != nil {
		return CacleResolutionRange(maxWithNone(a.Min, b.Min),
			minWithNone(a.Max, b.Max), nil, nil)
	}
	return nil
}
