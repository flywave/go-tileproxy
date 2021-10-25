package exports

import (
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

func TestArchiveExport(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./export.tar.gz"

	export, err := NewArchiveExport(filename, grid, imageopts, "tms")

	if err != nil {
		t.FailNow()
	}

	tile := cache.NewTile([3]int{1, 1, 1})
	tile.Source = cache.GetEmptyTile([2]uint32{256, 256}, imageopts)

	export.StoreTile(tile)

	export.Close()

	os.Remove("./export.tar.gz")
}

func TestGeoPackageExport(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./export.gpkg"

	export, err := NewGeoPackageExport(filename, "test", grid, imageopts)

	if err != nil {
		t.FailNow()
	}

	tile := cache.NewTile([3]int{1, 1, 1})
	tile.Source = cache.GetEmptyTile([2]uint32{256, 256}, imageopts)

	export.StoreTile(tile)

	export.Close()

	os.Remove("./export.gpkg")
}

func TestMBTilesExport(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	filename := "./export.mbtils"

	export, err := NewMBTilesExport(filename, grid, imageopts)

	if err != nil {
		t.FailNow()
	}

	tile := cache.NewTile([3]int{1, 1, 1})
	tile.Source = cache.GetEmptyTile([2]uint32{256, 256}, imageopts)

	export.StoreTile(tile)

	export.Close()

	os.Remove("./export.mbtils")
}
