package vector

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	m "github.com/flywave/go-mapbox/tileid"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geom"
	"github.com/flywave/go-mapbox/mvt"
	"github.com/flywave/go-mbgeom/geojson"
	"github.com/flywave/go-mbgeom/geojsonvt"
	"github.com/flywave/go-tileproxy/tile"
)

func TestMergeLK(t *testing.T) {
	pgcj02 := geo.NewProj("EPSG:GCJ02")
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{4096, 4096}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)

	bbox := grid.TileBBox([3]int{53958, 24829, 16}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)
	bbox = srs4326.TransformRectTo(pgcj02, bbox2, 16)

	_, _, tiles, _ := grid.GetAffectedTiles(bbox, [2]uint32{4096, 4096}, srs4326)

	tilesCoord := [][3]int{}
	for {
		x, y, z, done := tiles.Next()

		tilesCoord = append(tilesCoord, [3]int{x, y, z})

		if done {
			break
		}
	}

	sources := []tile.Source{}
	layers := []Vector{}

	merger := &VectorMerger{}

	for i := range tilesCoord {
		z, x, y := tilesCoord[i][2], tilesCoord[i][0], tilesCoord[i][1]
		source := NewMVTSource([3]int{x, y, z}, PBF_PTOTO_LUOKUANG, &VectorOptions{Format: PBF_MIME_LUOKUANG})

		source.SetSource(fmt.Sprintf("../data/%d_%d_%d.pbf", z, x, y))

		sources = append(sources, source)

		merger.AddSource(source, nil)

		tt := source.GetTile()

		layers = append(layers, tt.(Vector))
	}

	if len(sources) == 0 {
		t.FailNow()
	}

	tran := NewVectorTransformer(pgcj02, srs4326)

	tranlayers := []Vector{}

	for _, l := range layers {
		newfeats := make(Vector)
		for k, f := range l {
			newfeats[k] = tran.Apply(f)
		}
		tranlayers = append(tranlayers, newfeats)
	}

	if len(tranlayers) == 0 {
		t.FailNow()
	}

	all := make(Vector)

	for _, l := range tranlayers {
		for k, f := range l {
			if _, ok := all[k]; !ok {
				all[k] = []*geom.Feature{}
			}
			all[k] = append(all[k], f...)
		}
	}
	f, _ := os.Create("./test.mvt")
	SavePBF(f, [3]int{53958, 24829, 16}, PBF_PTOTO_MAPBOX, all)
	f.Close()

	os.Remove("./test.mvt")

	fc := geom.NewFeatureCollection()

	for _, f := range all {
		fc.Features = append(fc.Features, f...)
	}

	json, _ := fc.MarshalJSON()

	f2, _ := os.Create("./test.json")
	f2.Write(json)
	f2.Close()

	fjson, _ := os.Open("test.json")
	bytes, _ := ioutil.ReadAll(fjson)

	jsonvtdata := geojson.Parse(string(bytes))

	os.Remove("./test.json")

	opts := &geojsonvt.TileOptions{Tolerance: 0, LineMetrics: true, Buffer: 2048, Extent: 4096, MaxZoom: 20, IndexMaxZoom: 5, IndexMaxPoints: 100000, GenerateId: false}

	vt := geojsonvt.NewGeoJSONVT(jsonvtdata, *opts)
	vttile := vt.GetTile(16, 53958, 24829)

	fcc := vttile.GetFeatureCollection()

	jsonvt := fcc.Stringify()

	f3, _ := os.Create("./testvt.json")
	f3.Write([]byte(jsonvt))
	f3.Close()
	os.Remove("./testvt.json")
}

func TestLK(t *testing.T) {
	tileid := m.TileID{X: 53958, Y: 24829, Z: 16}

	name := fmt.Sprintf("../data/%d_%d_%d.pbf", 16, 53958, 24829)
	f, _ := os.Open(name)
	ddd, _ := ioutil.ReadAll(f)
	data := []byte{}
	tile, _ := mvt.NewTile(ddd, mvt.PROTO_LK)
	for _, layer := range tile.LayerMap {
		conf := mvt.NewConfig(layer.Name, tileid, mvt.PROTO_MAPBOX)
		conf.ExtentBool = false
		feats := []*geom.Feature{}
		for layer.Next() {
			feat, _ := layer.Feature()
			featg, _ := feat.ToGeoJSON(tileid)
			feats = append(feats, featg)
		}

		data = append(data, mvt.WriteLayer(feats, conf)...)
	}

	f2, _ := os.Create("./test1.mvt")
	f2.Write(data)
	f2.Close()

	os.Remove("./test1.mvt")
}
