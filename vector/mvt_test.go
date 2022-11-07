package vector

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-mapbox/mvt"
	m "github.com/flywave/go-mapbox/tileid"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

var bytevals1, _ = ioutil.ReadFile("../data/3194.mvt")
var tileid1 = m.TileID{X: 13515, Y: 6392, Z: 14}
var bytevals2, _ = ioutil.ReadFile("../data/tile.pbf")
var tileid2 = m.TileID{X: 1686, Y: 776, Z: 11}

func TestMVTSource(t *testing.T) {
	source := NewMVTSource([3]int{13515, 6392, 14}, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})

	source.SetSource("../data/3194.mvt")
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}

func TestMVTSourceBuffer(t *testing.T) {
	source := NewMVTSource([3]int{13515, 6392, 14}, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})

	source.SetSource(bytevals1)
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}

func TestPBFSource(t *testing.T) {
	source := NewMVTSource([3]int{1686, 776, 11}, PBF_PTOTO_LUOKUANG, &VectorOptions{Format: PBF_MIME, Proto: int(mvt.PROTO_LK)})

	source.SetSource("../data/tile.pbf")
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}

func TestPBFSourceBuffer(t *testing.T) {
	source := NewMVTSource([3]int{1686, 776, 11}, PBF_PTOTO_LUOKUANG, &VectorOptions{Format: PBF_MIME, Proto: int(mvt.PROTO_LK)})

	source.SetSource(bytevals2)
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}

var (
	tile_url = "https://a.tiles.mapbox.com/v4/mapbox.mapbox-streets-v8/%d/%d/%d.vector.pbf?sku=101mJslv5DbiL&access_token=pk.eyJ1IjoidzEyNTk0ODIyIiwiYSI6IkVfSkVqMGMifQ.av8k0fqnXvMFo1ThyV9KMQ"
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

	dst := fmt.Sprintf("%s/%d/%d/%d.mvt", sourceName, z, x, y)

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

func TestGetMVT(t *testing.T) {
	var bbox vec2d.Rect

	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	bbox = vec2d.Rect{
		Min: vec2d.T{113.50, 23.14},
		Max: vec2d.T{113.54, 23.18},
	}

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{512, 512}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	r, _, _ := grid.GetAffectedBBoxAndLevel(bbox, [2]uint32{512, 512}, srs4326)

	for _, l := range []int{14, 15, 16} {

		cbox, _, it, err := grid.GetAffectedLevelTiles(r, l)

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

		for i := range tilesCoord {
			z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

			src := fmt.Sprintf("./data/%d/%d/%d.mvt", z, x, y)

			if !fileExists(src) {
				download(x, y, z, "./data")
			}
		}
	}
}
