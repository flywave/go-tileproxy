package terrain

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type RasterMerger struct {
	Grid [2]int
	Size [2]uint32
}

func NewRasterMerger(tile_grid [2]int, tile_size [2]uint32) *RasterMerger {
	return &RasterMerger{Grid: tile_grid, Size: tile_size}
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

		tiledata.copyFrom(tdata, pos)
	}
	tiledata.Box = bbox
	tiledata.Boxsrs = bbox_srs

	return CreateRasterSourceFromTileData(tiledata, opts, cacheable)
}

func (t *RasterMerger) srcSize() [2]uint32 {
	width := uint32(t.Grid[0]) * t.Size[0]
	height := uint32(t.Grid[1]) * t.Size[1]
	return [2]uint32{width, height}
}

func (t *RasterMerger) tileOffset(i int) [2]int {
	return [2]int{int(math.Mod(float64(i), float64(t.Grid[0])) * float64(t.Size[0])), int(math.Floor(float64(i)/(float64(t.Grid[0]))) * float64(t.Size[1]))}
}

type RasterSplitter struct {
	dem     tile.Source
	Options *RasterOptions
}

func NewRasterSplitter(dem_tile tile.Source, dem_opts *RasterOptions) *RasterSplitter {
	return &RasterSplitter{dem: dem_tile, Options: dem_opts}
}

func (t *RasterSplitter) GetTile(newbox vec2d.Rect, boxsrs geo.Proj, tile_size [2]uint32) tile.Source {
	type _rasterSource interface {
		GetRasterOptions() *RasterOptions
		Resample(georef *geo.GeoReference, grid *Grid) error
	}
	rasterS, ok := t.dem.(_rasterSource)

	if !ok {
		return nil
	}

	georef := geo.NewGeoReference(newbox, boxsrs)

	grid := CaclulateGrid(int(tile_size[0]), int(tile_size[1]), rasterS.GetRasterOptions().Mode, georef)

	err := rasterS.Resample(nil, grid)

	if err != nil {
		return nil
	}

	smtd := grid.GetTileDate(t.Options.Mode)

	return CreateRasterSourceFromTileData(smtd, t.Options, nil)
}

type TiledRaster struct {
	Tiles    []tile.Source
	TileGrid [2]int
	TileSize [2]uint32
}

func NewTiledRaster(tiles []tile.Source, tile_grid [2]int, tile_size [2]uint32) *TiledRaster {
	return &TiledRaster{Tiles: tiles, TileGrid: tile_grid, TileSize: tile_size}
}

func (t *TiledRaster) GetRaster(dem_opts *RasterOptions) tile.Source {
	if dem_opts.Format.Extension() == "terrain" {
		tm := NewTerrainMerger(t.TileGrid)
		return tm.Merge(t.Tiles, dem_opts)
	} else {
		tm := NewRasterMerger(t.TileGrid, t.TileSize)
		return tm.Merge(t.Tiles, dem_opts)
	}
}

func (t *TiledRaster) Transform(req_bbox vec2d.Rect, req_srs geo.Proj, out_size [2]uint32, dem_opts *RasterOptions) tile.Source {
	src_img := t.GetRaster(dem_opts)
	transformer := NewRasterSplitter(src_img, dem_opts)
	return transformer.GetTile(req_bbox, req_srs, out_size)
}
