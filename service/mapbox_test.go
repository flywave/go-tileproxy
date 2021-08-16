package service

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

type dummyCreater struct {
}

func (c *dummyCreater) Create(size [2]uint32, opts tile.TileOptions, data interface{}) tile.Source {
	if data != nil {
		reader := data.(io.Reader)
		buf, _ := ioutil.ReadAll(reader)
		return &tile.DummyTileSource{Data: string(buf)}
	}
	return nil
}

func TestMapboxServiceGetTile(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{}}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	ccreater := func(location string) tile.Source {
		source := vector.NewMVTSource([3]int{13515, 6392, 14}, vector.PBF_PTOTO_MAPBOX, &vector.VectorOptions{Format: vector.PBF_MIME_MAPBOX})
		source.SetSource("../data/3194.mvt")
		return source
	}

	c := cache.NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	creater := &dummyCreater{}

	client := client.NewMapboxTileClient("https://api.mapbox.com", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", "mapbox.mapbox-streets-v8", mock)

	source := &sources.MapboxTileSource{Grid: grid, Client: client, SourceCreater: creater}

	locker := &cache.DummyTileLocker{}

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, -1, false, 0, [2]uint32{1, 1})

	md := make(map[string]string)

	tp := NewMapboxTileProvider("test", md, manager)

	if tp == nil {
		t.FailNow()
	}

	service := NewMapboxService(map[string]Provider{"mapbox.mapbox-streets-v8": tp}, nil, nil, md, nil, "ul")

	hreq := &http.Request{}
	hreq.URL, _ = url.Parse("https://127.0.0.1/v4/mapbox.mapbox-streets-v8/14/13515/6392.mvt")

	tileReq := request.NewMapboxTileRequest(hreq)

	resp := service.GetTile(tileReq)

	if resp == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestMapboxServiceGetStyle(t *testing.T) {

}

func TestMapboxServiceGetSprite(t *testing.T) {

}

func TestMapboxServiceGetGlyphs(t *testing.T) {

}
