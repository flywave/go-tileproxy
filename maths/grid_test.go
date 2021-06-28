package maths

import (
	"fmt"
	"math"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestMinRes(t *testing.T) {
	min_res := float64(1000)
	res := caclResolutions(&min_res, nil, nil, nil, nil, nil)

	if len(res) != 20 {
		t.FailNow()
	}
}

func TestMinMaxRes(t *testing.T) {
	min_res := float64(1000)
	max_res := float64(80)

	res := caclResolutions(&min_res, &max_res, nil, nil, nil, nil)

	if len(res) != 4 {
		t.FailNow()
	}
}

func TestMinResLevels(t *testing.T) {
	min_res := float64(1600)
	num_levels := 5

	res := caclResolutions(&min_res, nil, nil, &num_levels, nil, nil)

	if len(res) != 5 {
		t.FailNow()
	}
}

func TestMinResLevelsResFactor(t *testing.T) {
	min_res := float64(1600)
	num_levels := 4
	res_factor := 4.0

	res := caclResolutions(&min_res, nil, res_factor, &num_levels, nil, nil)

	if len(res) != 4 {
		t.FailNow()
	}
}

func TestMinResLevelsSqrt2(t *testing.T) {
	min_res := float64(1600)
	num_levels := 5

	res := caclResolutions(&min_res, nil, "sqrt2", &num_levels, nil, nil)

	if len(res) != 5 {
		t.FailNow()
	}
}

func TestMinResMaxResLevels(t *testing.T) {
	min_res := float64(1600)
	max_res := float64(10)

	num_levels := 10

	res := caclResolutions(&min_res, &max_res, nil, &num_levels, nil, nil)

	if len(res) != 10 {
		t.FailNow()
	}
}

func TestBoxLevels(t *testing.T) {
	bbox := vec2d.Rect{Min: vec2d.T{0, 40}, Max: vec2d.T{15, 50}}
	num_levels := 10
	tilesize := []uint32{256, 256}

	res := caclResolutions(nil, nil, nil, &num_levels, &bbox, tilesize)

	if len(res) != 10 {
		t.FailNow()
	}
}

func TestBBoxContains(t *testing.T) {
	if !BBoxContains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, vec2d.Rect{Min: vec2d.T{2, 2}, Max: vec2d.T{4, 4}}) {
		t.FailNow()
	}

	if BBoxContains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{11, 10}}) {
		t.FailNow()
	}
}

func TestTileGrid(t *testing.T) {
	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = NewSRSProj4("EPSG:900913")

	grid := NewTileGrid(conf)

	r := grid.Resolution(0)

	if r < 156543 {
		t.FailNow()
	}

	cx, cy, cz := grid.Tile(1000, 1000, 0)

	if cx > 0 || cy > 0 || cz > 0 {
		t.FailNow()
	}

	rect := grid.TileBBox([3]int{0, 0, 0}, false)
	if rect.Area() == 0 {
		t.FailNow()
	}

	rect = grid.TileBBox([3]int{1, 1, 1}, false)
	if rect.Area() == 0 {
		t.FailNow()
	}
}

func TestTileGridClosestLevel(t *testing.T) {
	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = NewSRSProj4("EPSG:900913")

	grid := NewTileGrid(conf)
	grid.StretchFactor = 1.1

	l1_res := grid.Resolution(1)

	res := []float64{320000.0, 160000.0, l1_res + 50, l1_res,
		l1_res - 50, l1_res * 0.91, l1_res * 0.89, 8000.0}
	result := make([]int, 0)
	for _, x := range res {
		result = append(result, grid.ClosestLevel(x))
	}
	for _, v := range result {
		t.Log(fmt.Sprintf("%d/n", v))
	}
}

func TestTileGridFlipTileCoord(t *testing.T) {
	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = NewSRSProj4("EPSG:900913")

	grid := NewTileGrid(conf)
	x, y, z := grid.FlipTileCoord(0, 1, 1)

	if x != 0 || y != 0 || z != 1 {
		t.FailNow()
	}

	x, y, z = grid.FlipTileCoord(1, 3, 2)

	if x != 1 || y != 0 || z != 2 {
		t.FailNow()
	}
}

func TestTileGridBasic(t *testing.T) {
	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = NewSRSProj4("EPSG:4326")
	conf[TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	conf[TILEGRID_ORIGIN] = ORIGIN_LL

	grid := NewTileGrid(conf)

	if !grid.SupportsAccessWithOrigin(ORIGIN_LL) {
		t.FailNow()
	}

	if grid.SupportsAccessWithOrigin(ORIGIN_UL) {
		t.FailNow()
	}
}

func TestEpsg4326BBox(t *testing.T) {

	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = NewSRSProj4("EPSG:4326")

	base := NewTileGrid(conf)
	bbox := vec2d.Rect{Min: vec2d.T{10.0, -20.0}, Max: vec2d.T{40.0, 10.0}}

	subConf := DefaultTileGridOptions()
	subConf[TILEGRID_ALIGN_WITH] = base
	subConf[TILEGRID_BBOX] = bbox

	sub := NewTileGrid(subConf)

	if !BBoxEquals(*sub.BBox, bbox, math.Inf(1), math.Inf(1)) {
		t.FailNow()
	}

	abbox, grid_size, tiles, err := sub.GetAffectedLevelTiles(bbox, 0)

	if err != nil {
		t.FailNow()
	}

	t.Log(fmt.Sprintf("%v--%v--%v--%v/n", abbox.Min[0], abbox.Min[1], abbox.Max[0], abbox.Max[1]))
	t.Log(fmt.Sprintf("%d--%d/n", grid_size[0], grid_size[1]))

	testv := make([][3]int, 0)
	for {
		x, y, z, done := tiles.Next()

		testv = append(testv, [3]int{x, y, z})
		t.Log(fmt.Sprintf("%d--%d--%d/n", x, y, z))

		if done {
			break
		}
	}
	if len(testv) != 4 {
		t.FailNow()
	}
}

func TestEpsg4326BBoxFromSqrt2(t *testing.T) {
	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = NewSRSProj4("EPSG:4326")
	conf[TILEGRID_RES_FACTOR] = "sqrt2"

	base := NewTileGrid(conf)

	bbox := vec2d.Rect{Min: vec2d.T{10.0, -20.0}, Max: vec2d.T{40.0, 10.0}}

	subConf := DefaultTileGridOptions()
	subConf[TILEGRID_ALIGN_WITH] = base
	subConf[TILEGRID_BBOX] = bbox
	subConf[TILEGRID_RES_FACTOR] = 2.0

	sub := NewTileGrid(subConf)

	if sub.Resolution(0) != base.Resolution(8) {
		t.FailNow()
	}
	if sub.Resolution(1) != base.Resolution(10) {
		t.FailNow()
	}
	if sub.Resolution(2) != base.Resolution(12) {
		t.FailNow()
	}
}

func TestEpsg4326BBoxToSqrt2(t *testing.T) {
	conf := DefaultTileGridOptions()
	conf[TILEGRID_SRS] = NewSRSProj4("EPSG:4326")
	conf[TILEGRID_RES_FACTOR] = 2.0

	base := NewTileGrid(conf)

	bbox := vec2d.Rect{Min: vec2d.T{10.0, -20.0}, Max: vec2d.T{40.0, 10.0}}

	subConf := DefaultTileGridOptions()
	subConf[TILEGRID_ALIGN_WITH] = base
	subConf[TILEGRID_BBOX] = bbox
	subConf[TILEGRID_RES_FACTOR] = "sqrt2"

	sub := NewTileGrid(subConf)

	if sub.Resolution(0) != base.Resolution(4) {
		t.FailNow()
	}
	if sub.Resolution(2) != base.Resolution(5) {
		t.FailNow()
	}
	if sub.Resolution(4) != base.Resolution(6) {
		t.FailNow()
	}
}

func TestMetaTileList(t *testing.T) {
	conf := DefaultTileGridOptions()
	mgrid := NewMetaGrid(NewTileGrid(conf), [2]uint32{2, 2}, 0)
	coord := [3]int{0, 0, 2}
	grids, _ := mgrid.metaBBox(&coord, nil, true)

	t.Log(fmt.Sprintf("%v--%v--%v--%v/n", grids.Min[0], grids.Min[1], grids.Max[0], grids.Max[1]))

	coord = [3]int{0, 0, 0}
	grids, _ = mgrid.metaBBox(&coord, nil, true)

	t.Log(fmt.Sprintf("%v--%v--%v--%v/n", grids.Min[0], grids.Min[1], grids.Max[0], grids.Max[1]))
}

func TestTilesPattern(t *testing.T) {
	conf := DefaultTileGridOptions()
	mgrid := NewMetaGrid(NewTileGrid(conf), [2]uint32{2, 2}, 0)

	tiles := mgrid.tilesPattern([2]uint32{2, 1}, []int{0, 0, 10, 10}, nil, [][3]int{{0, 1, 2}, {1, 1, 2}})

	if len(tiles) != 2 {
		t.FailNow()
	}
	tile := [3]int{1, 1, 2}

	tiles = mgrid.tilesPattern([2]uint32{2, 2}, []int{10, 20, 30, 40}, &tile, nil)

	if len(tiles) != 4 {
		t.FailNow()
	}
}

func TestMetaTileMetaBox(t *testing.T) {
	conf := DefaultTileGridOptions()
	mgrid := NewMetaGrid(NewTileGrid(conf), [2]uint32{2, 2}, 0)

	grids := mgrid.metaTileList([3]int{0, 1, 3}, [2]uint32{2, 2})

	if len(grids) != 4 {
		t.FailNow()
	}
}

func TestFullTileList(t *testing.T) {
	conf := DefaultTileGridOptions()
	mgrid := NewMetaGrid(NewTileGrid(conf), [2]uint32{2, 2}, 0)

	tiles, grid_size, bounds := mgrid.fullTileList([][3]int{{0, 0, 2}, {1, 1, 2}})

	if len(tiles) != 4 {
		t.FailNow()
	}

	if grid_size[0] != 2 || grid_size[1] != 2 {
		t.FailNow()
	}

	if len(bounds) != 2 {
		t.FailNow()
	}

}

func TestMetagridTiles(t *testing.T) {
	conf := DefaultTileGridOptions()
	mgrid := NewMetaGrid(NewTileGrid(conf), [2]uint32{2, 2}, 0)
	mtile := mgrid.GetMetaTile([3]int{0, 0, 0})

	ps := mtile.tile_patterns

	if len(ps) == 0 {
		t.FailNow()
	}

	mtile = mgrid.GetMetaTile([3]int{0, 1, 1})

	ps = mtile.tile_patterns

	if len(ps) == 0 {
		t.FailNow()
	}

	mtile = mgrid.GetMetaTile([3]int{1, 2, 2})

	ps = mtile.tile_patterns

	if len(ps) == 0 {
		t.FailNow()
	}
}

func TestResolutionRangeMeter(t *testing.T) {
	res_range := NewResolutionRange(newFloat64(1000), newFloat64(10))
	srs := NewSRSProj4("EPSG:900913")
	if res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100000, 100000}}, [2]uint32{10, 10}, srs) {
		t.FailNow()
	}
	if res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100000, 100000}}, [2]uint32{99, 99}, srs) {
		t.FailNow()
	}
	if !res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100000, 100000}}, [2]uint32{100, 100}, srs) {
		t.FailNow()
	}
	if !res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100000, 100000}}, [2]uint32{1000, 1000}, srs) {
		t.FailNow()
	}
	if !res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100000, 100000}}, [2]uint32{10000, 10000}, srs) {
		t.FailNow()
	}
	if res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100000, 100000}}, [2]uint32{10001, 10001}, srs) {
		t.FailNow()
	}
}

func TestResolutionRangeDeg(t *testing.T) {
	res_range := NewResolutionRange(newFloat64(100000), newFloat64(1000))
	srs := NewSRSProj4("EPSG:4326")
	if res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, [2]uint32{10, 10}, srs) {
		t.FailNow()
	}
	if res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, [2]uint32{11, 11}, srs) {
		t.FailNow()
	}
	if !res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, [2]uint32{12, 12}, srs) {
		t.FailNow()
	}
	if !res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, [2]uint32{100, 100}, srs) {
		t.FailNow()
	}
	if !res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, [2]uint32{1000, 1000}, srs) {
		t.FailNow()
	}
	if !res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, [2]uint32{1100, 1100}, srs) {
		t.FailNow()
	}
	if res_range.Contains(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, [2]uint32{1200, 1200}, srs) {
		t.FailNow()
	}
}

func TestFromScale(t *testing.T) {
	res_range := CacleResolutionRange(nil, nil, newFloat64(1e6), newFloat64(1e3))
	if (*res_range.Min - 280) > 0.01 {
		t.FailNow()
	}
	if (*res_range.Max - 0.28) > 0.01 {
		t.FailNow()
	}
}

func TestDifferentSRS(t *testing.T) {
	srs1 := NewSRSProj4("EPSG:4326")
	srs2 := NewSRSProj4("EPSG:3857")
	srs3 := NewSRSProj4("EPSG:900913")

	conf1 := DefaultTileGridOptions()
	conf1[TILEGRID_SRS] = srs1

	conf2 := DefaultTileGridOptions()
	conf2[TILEGRID_SRS] = srs2

	conf3 := DefaultTileGridOptions()
	conf3[TILEGRID_SRS] = srs3

	g1 := NewTileGrid(conf1)
	g2 := NewTileGrid(conf2)
	g3 := NewTileGrid(conf3)

	if g1.isSubsetOf(g2) {
		t.FailNow()
	}
	if !g3.isSubsetOf(g2) {
		t.FailNow()
	}

}

func TestLessLevels(t *testing.T) {
	srs := NewSRSProj4("EPSG:3857")

	conf1 := DefaultTileGridOptions()
	conf1[TILEGRID_SRS] = srs
	conf1[TILEGRID_NUM_LEVELS] = 10

	conf2 := DefaultTileGridOptions()
	conf2[TILEGRID_SRS] = srs

	g1 := NewTileGrid(conf1)
	g2 := NewTileGrid(conf2)

	if !g1.isSubsetOf(g2) {
		t.FailNow()
	}

	if g2.isSubsetOf(g1) {
		t.FailNow()
	}
}

func TestResSubset(t *testing.T) {
	srs := NewSRSProj4("EPSG:3857")

	conf1 := DefaultTileGridOptions()
	conf1[TILEGRID_SRS] = srs
	conf1[TILEGRID_RES] = []float64{50000, 10000, 100, 1}

	conf2 := DefaultTileGridOptions()
	conf2[TILEGRID_SRS] = srs
	conf2[TILEGRID_RES] = []float64{100000, 50000, 10000, 1000, 100, 10, 1, 0.5}

	g1 := NewTileGrid(conf1)
	g2 := NewTileGrid(conf2)

	if BBoxEquals(g1.TileBBox([3]int{0, 0, 0}, false), g2.TileBBox([3]int{0, 0, 0}, false), math.Inf(1), math.Inf(1)) {
		t.FailNow()
	}
	if !g1.isSubsetOf(g2) {
		t.FailNow()
	}
}

func TestSubbbox(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")

	conf1 := DefaultTileGridOptions()
	conf1[TILEGRID_SRS] = srs

	g1 := NewTileGrid(conf1)

	conf2 := DefaultTileGridOptions()
	conf2[TILEGRID_SRS] = srs
	conf2[TILEGRID_NUM_LEVELS] = 10
	conf2[TILEGRID_MIN_RES] = g1.Resolutions[3]
	conf2[TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{180, 90}}

	g2 := NewTileGrid(conf2)

	if !g2.isSubsetOf(g1) {
		t.FailNow()
	}
}

func TestIncompatibleSubbbox(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")

	conf1 := DefaultTileGridOptions()
	conf1[TILEGRID_SRS] = srs
	g1 := NewTileGrid(conf1)

	conf2 := DefaultTileGridOptions()
	conf2[TILEGRID_SRS] = srs
	conf2[TILEGRID_NUM_LEVELS] = 10
	conf2[TILEGRID_MIN_RES] = g1.Resolutions[3]
	conf2[TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-10, 0}, Max: vec2d.T{180, 90}}

	g2 := NewTileGrid(conf2)

	if g2.isSubsetOf(g1) {
		t.FailNow()
	}
}

func TestTileSize(t *testing.T) {
	srs := NewSRSProj4("EPSG:4326")

	conf1 := DefaultTileGridOptions()
	conf1[TILEGRID_SRS] = srs
	conf1[TILEGRID_TILE_SIZE] = []uint32{128, 128}

	conf2 := DefaultTileGridOptions()
	conf2[TILEGRID_SRS] = srs

	g1 := NewTileGrid(conf1)
	g2 := NewTileGrid(conf2)

	if g1.isSubsetOf(g2) {
		t.FailNow()
	}
}

func TestNOTileErrors(t *testing.T) {
	srs := NewSRSProj4("EPSG:3857")

	conf1 := DefaultTileGridOptions()
	conf1[TILEGRID_SRS] = srs
	conf1[TILEGRID_RES] = []float64{100000, 50000, 10000, 1000, 100, 10, 1, 0.5}

	conf2 := DefaultTileGridOptions()
	conf2[TILEGRID_SRS] = srs
	conf2[TILEGRID_RES] = []float64{100, 1}

	g1 := NewTileGrid(conf1)
	g2 := NewTileGrid(conf2)

	if g1.isSubsetOf(g2) {
		t.FailNow()
	}
}

func TestMergeResolutions(t *testing.T) {
	res_range := MergeResolutionRange(
		CacleResolutionRange(nil, newFloat64(10), nil, nil), CacleResolutionRange(newFloat64(1000), nil, nil, nil))
	if res_range != nil {
		t.FailNow()
	}

	res_range = MergeResolutionRange(
		CacleResolutionRange(newFloat64(10000), newFloat64(10), nil, nil), CacleResolutionRange(newFloat64(1000), nil, nil, nil))
	if *res_range.Min != 10000 || res_range.Max != nil {
		t.FailNow()
	}
	res_range = MergeResolutionRange(
		CacleResolutionRange(newFloat64(10000), newFloat64(10), nil, nil), CacleResolutionRange(newFloat64(1000), newFloat64(1), nil, nil))
	if *res_range.Min != 10000 || *res_range.Max != 1 {
		t.FailNow()
	}
	res_range = MergeResolutionRange(
		CacleResolutionRange(newFloat64(10000), newFloat64(10), nil, nil), CacleResolutionRange(nil, nil, nil, nil))
	if res_range != nil {
		t.FailNow()
	}

	res_range = MergeResolutionRange(
		nil, CacleResolutionRange(nil, nil, nil, nil))
	if res_range != nil {
		t.FailNow()
	}
	res_range = MergeResolutionRange(
		CacleResolutionRange(newFloat64(10000), newFloat64(10), nil, nil), nil)
	if res_range != nil {
		t.FailNow()
	}
}
