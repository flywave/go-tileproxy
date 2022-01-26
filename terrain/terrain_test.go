package terrain

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/flywave/go-cog"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestTerrainSource(t *testing.T) {
	opts := &RasterOptions{Format: tile.TileFormat("terrain"), Mode: BORDER_BILATERAL}

	source := NewTerrainSource(opts)

	source.SetSource("../data/323.terrain")
	t1 := source.GetTile()

	if t1 == nil {
		t.FailNow()
	}
}

func TestGeo(t *testing.T) {
	srs4326 := geo.NewProj(4326)
	srs900913 := geo.NewProj(900913)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs4326
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	mdGrid := geo.NewMetaGrid(grid, [2]uint32{1, 1}, 1)

	mconf := geo.DefaultTileGridOptions()
	mconf[geo.TILEGRID_SRS] = srs900913
	mconf[geo.TILEGRID_RES_FACTOR] = 2.0
	mconf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	mconf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	mgrid := geo.NewTileGrid(mconf)

	tileid := [3]int{6779, 1210, 13}

	realbbox := grid.TileBBox(tileid, false)

	tbbox := mdGrid.GetMetaTile(tileid)

	bbox := tbbox.GetBBox()

	bbox2 := srs4326.TransformRectTo(srs900913, bbox, 16)

	_, grids, it, err := mgrid.GetAffectedLevelTiles(bbox2, 13)

	if err != nil {
		t.FailNow()
	}

	tilesCoord := [][3]int{}
	minx, miny := 0, 0
	for {
		x, y, z, done := it.Next()

		if minx == 0 || x < minx {
			minx = x
		}

		if miny == 0 || y < miny {
			miny = y
		}

		tilesCoord = append(tilesCoord, [3]int{x, y, z})

		if done {
			break
		}
	}

	if len(tilesCoord) == 0 {
		t.FailNow()
	}

	var sources []tile.Source
	opts := &RasterOptions{Format: tile.TileFormat("webp"), Mode: BORDER_BILATERAL}

	var psize []float64

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

		source := NewDemRasterSource(ModeMapbox, opts)

		b := mgrid.TileBBox([3]int{x, y, z}, false)

		bbox_ := srs900913.TransformRectTo(srs4326, b, 16)
		if psize == nil {
			psize = caclulatePixelSize(512, 512, bbox_)
		}

		source.georef = geo.NewGeoReference(bbox_, srs4326)

		source.SetSource(fmt.Sprintf("../data/%d_%d_%d.webp", z, x, y))

		sources = append(sources, source)
	}

	m := NewRasterMerger(grids, [2]uint32{512, 512})
	rr := m.Merge(sources, opts)

	if rr == nil {
		t.FailNow()
	}

	wopts := &RasterOptions{Format: tile.TileFormat("webp"), MaxError: 2, Mode: BORDER_UNILATERAL}

	splitter := NewRasterSplitter(rr, wopts)

	newSize := [2]uint32{
		uint32(math.Ceil((realbbox.Max[0] - realbbox.Min[0]) / psize[0])),
		uint32(math.Ceil((realbbox.Max[1] - realbbox.Min[1]) / psize[1])),
	}

	newTile := splitter.GetTile(realbbox, srs4326, newSize)

	smtd := newTile.GetTile().(*TileData)

	smtd.Box = realbbox
	smtd.Boxsrs = srs4326

	rect := image.Rect(0, 0, int(newSize[0]), int(newSize[1]))

	src := cog.NewSource(smtd.Datas, &rect, cog.CTLZW)

	cog.WriteTile("./test.tif", src, realbbox, srs4326, newSize, nil)

	topts := &RasterOptions{Format: tile.TileFormat("terrain"), MaxError: 2}

	source, err := GenTerrainSource(smtd, topts)

	t1 := source.GetTile()

	buff := source.GetBuffer(nil, nil)

	f, _ := os.Create("./data.terrain")
	f.Write(buff)
	f.Close()

	if t1 == nil && err != nil {
		t.FailNow()
	}
}

func TestGenTerrainSourceFromDem(t *testing.T) {
	f, _ := os.Open("../data/14_13733_6366.webp")
	data, _ := LoadDEM(f, ModeMapbox)
	f.Close()

	tiledata := NewTileData([2]uint32{uint32(data.Dim - 2), uint32(data.Dim - 2)}, BORDER_UNILATERAL)
	for x := 0; x < data.Dim; x++ {
		for y := 0; y < data.Dim; y++ {
			if x > 0 && y > 0 && x < data.Dim-1 && y < data.Dim-1 {
				tiledata.Set(x-1, y-1, data.Get(x, y))
			}

			if x == 0 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_LEFT, y-1, data.Get(x, y))
			}

			if x == data.Dim-1 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_RIGHT, y-1, data.Get(x, y))
			}

			if y == 0 {
				tiledata.FillBorder(BORDER_TOP, x, data.Get(x, y))
			}

			if y == data.Dim-1 {
				tiledata.FillBorder(BORDER_BOTTOM, x, data.Get(x, y))
			}
		}
	}

	srs4326 := geo.NewProj(4326)
	srs900913 := geo.NewProj(900913)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs4326
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_SW

	grid := geo.NewTileGrid(conf)

	//bbox := grid.TileBBox([3]int{53958, 24829, 16}, false)
	//bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)
	//bbox = srs4326.TransformRectTo(pgcj02, bbox2, 16)

	bbox := grid.TileBBox([3]int{13733, 6366, 14}, false)

	tiledata.Box = bbox
	tiledata.Boxsrs = srs900913

	opts := &RasterOptions{Format: tile.TileFormat("terrain"), MaxError: 2}

	source, err := GenTerrainSource(tiledata, opts)

	t1 := source.GetTile()

	buff := source.GetBuffer(nil, nil)

	f, _ = os.Create("./data.terrain")
	f.Write(buff)
	f.Close()

	if t1 == nil && err != nil {
		t.FailNow()
	}
}

func TestGenTerrainSourceFromLerc(t *testing.T) {
	lercio := &LercIO{Mode: BORDER_UNILATERAL}

	f, _ := os.Open("../data/title_13_3252_6773.atm")
	tiledata, _ := lercio.Decode(f)
	f.Close()

	if tiledata == nil {
		t.FailNow()
	}

	srs900913 := geo.NewProj(900913)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{3252, 6773, 13}, false)

	tiledata.Box = bbox
	tiledata.Boxsrs = srs900913

	opts := &RasterOptions{Format: tile.TileFormat("terrain"), MaxError: 2}

	source, err := GenTerrainSource(tiledata, opts)

	t1 := source.GetTile()

	buff := source.GetBuffer(nil, nil)

	f, _ = os.Create("./data.terrain")
	f.Write(buff)
	f.Close()

	if t1 == nil && err != nil {
		t.FailNow()
	}
}

const (
	mp_turl = "https://api.maptiler.com/tiles/terrain-quantized-mesh/%d/%d/%d.terrain?key=RQ1nmTeb2C0dgCERCEVO"
)

func download_mp(x, y, z int, sourceName string) {
	data := get_url(fmt.Sprintf(mp_turl, z, x, y))

	dst := fmt.Sprintf("%s/%d/%d/%d.terrain", sourceName, z, x, y)

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		fmt.Printf("mkdirAll error")
	}
	f, _ := os.Create(dst)
	f.Write(data)
	f.Close()
}

func TestGetTerrain(t *testing.T) {
	bbox := vec2d.Rect{
		Min: vec2d.T{118.04672241210939, 36.46105407505434},
		Max: vec2d.T{118.14971923828126, 36.54384614538856},
	}

	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs4326
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	r, _, _ := grid.GetAffectedBBoxAndLevel(bbox, [2]uint32{512, 512}, srs4326)

	for l := 0; l < 13; l++ {

		_, _, it, _ := grid.GetAffectedLevelTiles(r, l)

		tilesCoord := [][3]int{}
		minx, miny := 0, 0
		for {
			x, y, z, done := it.Next()

			if minx == 0 || x < minx {
				minx = x
			}

			if miny == 0 || y < miny {
				miny = y
			}

			tilesCoord = append(tilesCoord, [3]int{x, y, z})

			if done {
				break
			}
		}

		if len(tilesCoord) == 0 {
			t.FailNow()
		}

		for i := range tilesCoord {
			z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

			src := fmt.Sprintf("./data/%d/%d/%d.terrain", z, x, y)

			if !fileExists(src) {
				download_mp(x, y, z, "./data")
			}

		}
	}
}
