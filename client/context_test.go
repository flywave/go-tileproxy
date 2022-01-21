package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flywave/go-geo"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestCollectorContext(t *testing.T) {
	conf := &Config{SkipSSL: false, Threads: 1, UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36", RandomDelay: 0, DisableKeepAlives: false, RequestTimeout: 5 * time.Second}

	client := NewCollectorClient(conf, nil)

	code, data := client.Open("https://api.mapbox.com/v4/mapbox.mapbox-streets-v8/1/0/0.mvt?access_token=pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", nil)

	if code != 200 || data == nil {
		t.FailNow()
	}

	ioutil.WriteFile("./test.mvt", data, os.ModePerm)
}

const (
	tile_url = "https://api.mapbox.com/v4/mapbox.satellite/%d/%d/%d.webp?sku=101h7nLHLyzgw&access_token=pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ"
)

func download(x, y, z int, sourceName string, client *CollectorClient) {
	_, data := client.Open(fmt.Sprintf(tile_url, z, x, y), nil)

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

func TestGet(t *testing.T) {
	cconf := &Config{SkipSSL: false, Threads: 16, UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36", RandomDelay: 0, DisableKeepAlives: false, RequestTimeout: 5 * time.Second}
	cclient := NewCollectorClient(cconf, nil)

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

	cbox, _, it, err := grid.GetAffectedLevelTiles(r, 18)

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

	os.Mkdir("./test_data", os.ModePerm)

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]

		src := fmt.Sprintf("./test_data/satellite_%d_%d_%d.webp", z, x, y)

		if !fileExists(src) {
			download(x, y, z, "./test_data", cclient)
		}
	}

	os.RemoveAll("./test_data")
}
