package terrain

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type RasterMerger struct {
	Grid    [2]int
	Size    [2]uint32
	Creater func(data *TileData, opts tile.TileOptions, cacheable *tile.CacheInfo) tile.Source
}

func NewRasterMerger(tile_grid [2]int, tile_size [2]uint32, creater func(data *TileData, opts tile.TileOptions, cacheable *tile.CacheInfo) tile.Source) *RasterMerger {
	return &RasterMerger{Grid: tile_grid, Size: tile_size, Creater: creater}
}

func (t *RasterMerger) Merge(ordered_tiles []tile.Source, opts *RasterOptions) tile.Source {
	if t.Grid[0] == 1 && t.Grid[1] == 1 {
		if len(ordered_tiles) >= 1 && ordered_tiles[0] != nil {
			tile := ordered_tiles[0]
			return tile
		}
	}

	src_size := t.srcSize()

	var cacheable *tile.CacheInfo

	fdata := ordered_tiles[0].GetTile().(*TileData)
	georef := ordered_tiles[0].GetGeoReference()

	var bbox vec2d.Rect
	var bbox_srs geo.Proj

	if georef != nil {
		bbox = georef.GetBBox()
		bbox_srs = georef.GetSrs()
	}

	mode := fdata.Border

	tiledata := NewTileData(src_size, mode)

	for i, source := range ordered_tiles {
		if source == nil {
			continue
		}

		if source.GetCacheable() == nil {
			cacheable = source.GetCacheable()
		}

		tdata := source.GetTile().(*TileData)
		georef := source.GetGeoReference()

		if tdata.Border != mode {
			continue
		}
		if georef != nil {
			bboxss := georef.GetBBox()
			bbox = vec2d.Joined(&bbox, &bboxss)
		}
		pos := t.tileOffset(i)

		tiledata.CopyFrom(tdata, pos)
	}
	tiledata.Box = bbox
	tiledata.Boxsrs = bbox_srs

	return t.Creater(tiledata, opts, cacheable)
}

func (t *RasterMerger) srcSize() [2]uint32 {
	width := uint32(t.Grid[0]) * t.Size[0]
	height := uint32(t.Grid[1]) * t.Size[1]
	return [2]uint32{width, height}
}

func (t *RasterMerger) tileOffset(i int) [2]int {
	return [2]int{int(math.Mod(float64(i), float64(t.Grid[0])) * float64(t.Size[0])), int(math.Floor(float64(i)/(float64(t.Grid[0]))) * float64(t.Size[1]))}
}
