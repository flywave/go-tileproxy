package terrain

import (
	"testing"

	"github.com/flywave/go-tileproxy/tile"
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
