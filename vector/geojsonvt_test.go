package vector

import (
	"testing"

	"github.com/flywave/go-tileproxy/tile"
)

func TestGeojsonVTSource(t *testing.T) {
	source := NewGeoJSONVTSource([3]int{13515, 6392, 14}, &VectorOptions{Extent: 4096, Buffer: 1024, Format: tile.TileFormat("application/json")})

	source.SetSource("../data/us-states-tiles.json")
	t1 := source.GetTile()

	if t1 == nil {
		t.FailNow()
	}

	source2 := NewGeoJSONVTSource([3]int{13515, 6392, 14}, &VectorOptions{Extent: 4096, Buffer: 1024, Format: tile.TileFormat("application/json")})

	source2.SetSource(t1)

	bytes := source2.GetBuffer(nil, nil)

	jsonvt := string(bytes)

	if jsonvt == "" {
		t.FailNow()
	}
}
