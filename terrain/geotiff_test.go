package terrain

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flywave/go-cog"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

var (
	tile_url = "https://api.mapbox.com/raster/v1/mapbox.mapbox-terrain-dem-v1/%d/%d/%d.webp?sku=101XxiLvoFYxL&access_token=pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja3pjOXRqcWkybWY3MnVwaGxkbTgzcXAwIn0._tCv9fpOyCT4O_Tdpl6h0w"
)

func get_url(url string) []byte {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	return result.Bytes()
}

func download(x, y, z int, sourceName string) {
	data := get_url(fmt.Sprintf(tile_url, z, x, y))

	dst := fmt.Sprintf("%s/%d_%d_%d.webp", sourceName, z, x, y)

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		fmt.Printf("mkdirAll error")
	}
	f, _ := os.Create(dst)
	f.Write(data)
	f.Close()
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	} else if err != nil {
		return false
	}
	return true
}

func TestGetGeotiff(t *testing.T) {
	var bbox vec2d.Rect

	srs32650 := geo.NewProj(32650)
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	if false {
		bbox = vec2d.Rect{
			Min: vec2d.T{117.4879, 36.7371},
			Max: vec2d.T{117.7070, 36.9141},
		}
		bbox = srs32650.TransformRectTo(srs4326, bbox, 16)
	} else {
		bbox = vec2d.Rect{
			Min: vec2d.T{117.4879, 36.7371},
			Max: vec2d.T{117.7070, 36.9141},
		}
	}

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	r, _, _ := grid.GetAffectedBBoxAndLevel(bbox, [2]uint32{512, 512}, srs4326)

	cbox, grids, it, err := grid.GetAffectedLevelTiles(r, 14)

	sbox := srs900913.TransformRectTo(srs4326, cbox, 16)

	if err != nil || sbox.Min[0] == 0 {
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

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

		src := fmt.Sprintf("../data/%d_%d_%d.webp", z, x, y)

		if !fileExists(src) {
			download(x, y, z, "../data")
		}

		source := NewDemRasterSource(ModeMapbox, opts)

		b := grid.TileBBox([3]int{x, y, z}, false)

		source.georef = geo.NewGeoReference(b, srs900913)

		source.SetSource(fmt.Sprintf("../data/%d_%d_%d.webp", z, x, y))

		sources = append(sources, source)
	}

	m := NewRasterMerger(grids, [2]uint32{512, 512})
	rr := m.Merge(sources, opts)

	tiledata := rr.GetTile().(*TileData)

	io := &DemIO{Mode: ModeMapbox, Format: tile.TileFormat("webp")}

	raw, _ := io.Encode(tiledata)

	f, _ := os.Create("./taihe.webp")
	f.Write(raw)
	f.Close()

	rect := image.Rect(0, 0, int(tiledata.Size[0]), int(tiledata.Size[1]))

	src := cog.NewSource(tiledata.Datas, &rect, cog.CTLZW)

	cog.WriteTile("./taihe.tif", src, sbox, srs4326, tiledata.Size, nil)
}

func TestGeoTIFF(t *testing.T) {
	f, _ := os.Open("../data/14_13733_6366.webp")
	data, _ := LoadDEM(f, ModeMapbox)
	f.Close()

	tiledata := NewTileData([2]uint32{uint32(data.Dim - 2), uint32(data.Dim - 2)}, BORDER_BILATERAL)
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

	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{514, 514}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{13733, 6366, 14}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)

	tiledata.Box = bbox2
	tiledata.Boxsrs = srs4326

	tiffio := &GeoTIFFIO{Mode: BORDER_BILATERAL}

	tiff, _ := tiffio.Encode(tiledata)

	f, _ = os.Create("./data.tiff")
	f.Write(tiff)
	f.Close()

	tiffio2 := &GeoTIFFIO{Mode: BORDER_BILATERAL}

	f, _ = os.Open("./data.tiff")
	tiledata2, _ := tiffio2.Decode(f)

	if tiledata2 == nil {
		t.FailNow()
	}
}
