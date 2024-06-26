package client

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/crawler"
)

var (
	httpConf = Config{
		SkipSSL:           false,
		Threads:           2,
		UserAgent:         "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0",
		RandomDelay:       2,
		DisableKeepAlives: false,
		Proxys:            nil,
		RequestTimeout:    time.Duration(20 * time.Second),
	}
)

func TestHttpFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(serverHandler))
	defer server.Close()
	rng := rand.New(rand.NewSource(12387123712321232))
	var (
		requests uint32
		success  uint32
		failure  uint32
	)
	client := NewCollectorClient(&httpConf, nil)
	client.Collector.OnResponse(func(resp *crawler.Response) {
		if resp.StatusCode == http.StatusOK {
			atomic.AddUint32(&success, 1)
		} else {
			atomic.AddUint32(&failure, 1)
		}

		if resp.UserData != nil {
			fut := resp.UserData.(*Future)
			fut.setResult(resp)
		}
	})

	futs := []*Future{}

	for i := 0; i < 30; i++ {
		ti := time.Duration(rng.Intn(50)) * time.Microsecond
		uri := server.URL + "/delay?t=" + ti.String()

		u, _ := whatwgUrl.Parse(uri)
		u2, _ := url.Parse(u.Href(false))

		atomic.AddUint32(&requests, 1)

		fut := newFuture()

		futs = append(futs, fut)

		client.Collector.Visit(u2.String(), fut, nil)
	}

	client.Collector.Wait()

	if success+failure != requests || failure > 0 || len(futs) == 0 {
		t.Fatalf("wrong Queue implementation: "+
			" requests = %d, success = %d, failure = %d",
			requests, success, failure)
	}
}

func serverHandler(w http.ResponseWriter, req *http.Request) {
	if !serverRoute(w, req) {
		shutdown(w)
	}
}

func serverRoute(w http.ResponseWriter, req *http.Request) bool {
	if req.URL.Path == "/delay" {
		return serveDelay(w, req) == nil
	}
	return false
}

func serveDelay(w http.ResponseWriter, req *http.Request) error {
	q := req.URL.Query()
	t, err := time.ParseDuration(q.Get("t"))
	if err != nil {
		return err
	}
	time.Sleep(t)
	w.WriteHeader(http.StatusOK)
	return nil
}

func shutdown(w http.ResponseWriter) {
	taker, ok := w.(http.Hijacker)
	if !ok {
		return
	}
	raw, _, err := taker.Hijack()
	if err != nil {
		return
	}
	raw.Close()
}

const (
	tile_url = "https://api.mapbox.com/v4/mapbox.satellite/%d/%d/%d.webp?sku=101h7nLHLyzgw&access_token=pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja3pjOXRqcWkybWY3MnVwaGxkbTgzcXAwIn0._tCv9fpOyCT4O_Tdpl6h0w"
)

func download(x, y, z int, sourceName string, client *CollectorClient) {
	_, data := client.Open(fmt.Sprintf(tile_url, z, x, y), nil, nil)

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
