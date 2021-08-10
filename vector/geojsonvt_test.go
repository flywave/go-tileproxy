package vector

import (
	"testing"

	"github.com/flywave/go-mbgeom/geojsonvt"
)

func TestGeojsonVTSource(t *testing.T) {
	source := NewGeoJSONVTSource([3]int{13515, 6392, 14}, &GeoJSONVTOptions{Options: geojsonvt.TileOptions{Extent: 4096, Buffer: 1024}})

	source.SetSource("../data/us-states-tiles.json")
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}

	source2 := NewGeoJSONVTSource([3]int{13515, 6392, 14}, &GeoJSONVTOptions{Options: geojsonvt.TileOptions{Extent: 4096, Buffer: 1024}})

	source2.SetSource(tile)

	bytes := source2.GetBuffer(nil, nil)

	jsonvt := string(bytes)

	if jsonvt == "" {
		t.FailNow()
	}

}
