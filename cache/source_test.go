package cache

import (
	"testing"

	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/vector"
)

func TestGetEmptyTile(t *testing.T) {
	imageOpt := &imagery.ImageOptions{Format: "png"}
	empty := GetEmptyTile([2]uint32{256, 256}, imageOpt)
	buff := empty.GetBuffer(nil, nil)
	if empty == nil || buff == nil {
		t.FailNow()
	}

	rasertOpt := &terrain.RasterOptions{Format: "webp", Mode: terrain.BORDER_BILATERAL}

	empty = GetEmptyTile([2]uint32{256, 256}, rasertOpt)
	buff = empty.GetBuffer(nil, nil)
	if empty == nil || buff == nil {
		t.FailNow()
	}

	terrainOpt := &terrain.RasterOptions{Format: "terrain", Mode: terrain.BORDER_BILATERAL}

	empty = GetEmptyTile([2]uint32{256, 256}, terrainOpt)
	buff = empty.GetBuffer(nil, nil)
	if empty == nil || buff == nil {
		t.FailNow()
	}

	mvtOpt := &vector.VectorOptions{Format: "mvt"}
	empty = GetEmptyTile([2]uint32{4096, 4096}, mvtOpt)
	buff = empty.GetBuffer(nil, nil)
	if empty == nil || buff == nil {
		t.FailNow()
	}

	jsonOpt := &vector.VectorOptions{Format: "json"}
	empty = GetEmptyTile([2]uint32{4096, 4096}, jsonOpt)
	buff = empty.GetBuffer(nil, nil)
	if empty == nil || buff == nil {
		t.FailNow()
	}
}

func TestResampleTile(t *testing.T) {

}
