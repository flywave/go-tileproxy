package cache

import (
	"errors"
	"io"
	"io/ioutil"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

func GetEmptyTile(size [2]uint32, opts tile.TileOptions) tile.Source {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return imagery.NewBlankImageSource(size, opt, nil)
	case *terrain.RasterOptions:
		return terrain.NewBlankRasterSource(size, opt, nil)
	case *vector.VectorOptions:
		return vector.NewBlankVectorSource(size, opt, nil)
	}
	return nil
}

func BlendTiles(layers []tile.Source, opts tile.TileOptions, size [2]uint32, bbox vec2d.Rect, Srs geo.Proj, tileMerger tile.Merger) tile.Source {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return imagery.MergeImages(layers, opt, size, bbox, Srs, tileMerger)
	}
	return nil
}

func splitImageMetaTiles(meta_tile tile.Source, tiles []geo.TilePattern, tile_size [2]uint32, image_opts *imagery.ImageOptions) *TileCollection {
	splitter := imagery.NewTileSplitter(meta_tile, image_opts)
	split_tiles := NewTileCollection(nil)
	for _, tile := range tiles {
		tile_coord, crop_coord := tile.Tiles, tile.Sizes
		if tile_coord[0] < 0 || tile_coord[1] < 0 || tile_coord[2] < 0 {
			continue
		}
		data := splitter.GetTile(crop_coord, tile_size)
		new_tile := NewTile(tile_coord)
		new_tile.SetCacheInfo(meta_tile.GetCacheable())
		new_tile.Source = data
		split_tiles.SetItem(new_tile)
	}
	return split_tiles
}

func SplitTiles(layers tile.Source, tiles []geo.TilePattern, tile_size [2]uint32, opts tile.TileOptions) *TileCollection {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return splitImageMetaTiles(layers, tiles, tile_size, opt)
	}
	return nil
}

func ScaleTiles(layers []tile.Source, queryBBox vec2d.Rect, querySrs geo.Proj, src_tile_grid [2]int, grid *geo.TileGrid, src_bbox vec2d.Rect, opts tile.TileOptions) tile.Source {
	switch op := opts.(type) {
	case *imagery.ImageOptions:
		tiled_image := imagery.NewTiledImage(layers, src_tile_grid, [2]uint32{grid.TileSize[0], grid.TileSize[1]}, src_bbox, grid.Srs)
		return tiled_image.Transform(queryBBox, querySrs, [2]uint32{grid.TileSize[0], grid.TileSize[1]}, op)
	}
	return nil
}

func MaskImageSourceFromCoverage(source tile.Source, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage, opts tile.TileOptions) tile.Source {
	switch op := opts.(type) {
	case *imagery.ImageOptions:
		imagery.MaskImageSourceFromCoverage(source, bbox, bbox_srs, coverage, op)
	}
	return nil
}

func EncodeTile(opts tile.TileOptions, tile [3]int, data tile.Source) ([]byte, error) {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return data.GetBuffer(nil, opts), nil
	case *terrain.RasterOptions:
		return terrain.EncodeRaster(opt, data.GetTile().(*terrain.TileData))
	case *vector.VectorOptions:
		return vector.EncodeVector(opt, tile, data.GetTile().(vector.Vector))
	}
	return nil, errors.New("error")
}

func DecodeTile(opts tile.TileOptions, tile [3]int, reader io.Reader) (tile.Source, error) {
	data, _ := ioutil.ReadAll(reader)
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return imagery.CreateImageSourceFromBufer(data, opt), nil
	case *terrain.RasterOptions:
		return terrain.CreateRasterSourceFromBufer(data, opt), nil
	case *vector.VectorOptions:
		return vector.CreateVectorSourceFromBufer(data, tile, opt), nil
	}
	return nil, errors.New("error")
}
