package imagery

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flywave/go-cog"
	"github.com/flywave/go-tileproxy/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/flywave/imaging"

	"github.com/flywave/go-geo"

	"github.com/flywave/go-geos"
)

func TestMergeSingleCoverage(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(true)
	img_opts.BgColor = color.Transparent
	img := CreateImageSource([2]uint32{10, 10}, &img_opts)

	nimg := img.GetImage().(*image.NRGBA)

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			nimg.Set(x, y, color.NRGBA{128, 128, 255, 255})
		}
	}

	geom := geos.CreatePolygon([]geos.Coord{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}})

	coverage1 := geo.NewGeosCoverage(geom, geo.NewProj(3857), true)

	merger := &LayerMerger{}
	merger.AddSource(img, coverage1)

	result := merger.Merge(&img_opts, nil, vec2d.Rect{Min: vec2d.T{5, 0}, Max: vec2d.T{15, 10}}, geo.NewProj(3857), nil)

	ri := result.GetTile().(image.Image)
	c := ri.At(6, 0)
	_, _, _, A := c.RGBA()
	if A != 0 {
		t.FailNow()
	}
}

const (
	tile_url = "https://api.mapbox.com/v4/mapbox.satellite/%d/%d/%d.webp?sku=101h7nLHLyzgw&access_token=pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ"
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

	dst := fmt.Sprintf("%s/satellite_%d_%d_%d.webp", sourceName, z, x, y)

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

	bbox := vec2d.Rect{
		Min: vec2d.T{118.0787624999999963, 36.4794427545898472},
		Max: vec2d.T{118.1429638549804650, 36.5374643000000034},
	}

	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	r, _, _ := grid.GetAffectedBBoxAndLevel(bbox, [2]uint32{256, 256}, srs4326)

	cbox, grids, it, err := grid.GetAffectedLevelTiles(r, 18)

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

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

		src := fmt.Sprintf("../data/satellite_%d_%d_%d.webp", z, x, y)

		if !fileExists(src) {
			download(x, y, z, "../data")
		}

		source := &ImageSource{Options: WEBP_FORMAT}

		b := grid.TileBBox([3]int{x, y, z}, false)

		source.georef = geo.NewGeoReference(b, srs900913)

		source.SetSource(fmt.Sprintf("../data/satellite_%d_%d_%d.webp", z, x, y))

		sources = append(sources, source)
	}

	m := NewTileMerger(grids, [2]uint32{256, 256})
	result := m.Merge(sources, WEBP_FORMAT)

	img := result.GetTile().(image.Image)

	imaging.Save(img, "./test.png")

	src := cog.NewSource(img, nil, cog.CTLZW)

	cog.WriteTile("./satellite_taihe.tif", src, sbox, srs4326, [2]uint32{uint32(img.Bounds().Dx()), uint32(img.Bounds().Dy())}, nil)
}
