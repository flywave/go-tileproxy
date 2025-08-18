package cache

import (
	"errors"
	"io"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
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

func ResampleTiles(layers []tile.Source, queryBBox vec2d.Rect, querySrs geo.Proj, src_tile_grid [2]int, grid *geo.TileGrid, src_bbox vec2d.Rect, srcSrs geo.Proj, out_size [2]uint32, src_opt, dest_opt tile.TileOptions) (tile.Source, error) {
	size := [2]uint32{grid.TileSize[0], grid.TileSize[1]}
	switch opt := dest_opt.(type) {
	case *imagery.ImageOptions:
		return imagery.Resample(layers, src_tile_grid, size, src_bbox, srcSrs, queryBBox, querySrs, out_size, src_opt.(*imagery.ImageOptions), opt), nil
	case *terrain.RasterOptions:
		return terrain.Resample(layers, src_tile_grid, size, src_bbox, srcSrs, queryBBox, querySrs, out_size, src_opt.(*terrain.RasterOptions), opt), nil
	case *vector.VectorOptions:
		return vector.Resample(layers, src_tile_grid, size, grid, src_bbox, srcSrs, queryBBox, querySrs, out_size, src_opt.(*vector.VectorOptions), opt), nil
	}
	return nil, errors.New("not support source")
}

func MergeTiles(layers []tile.Source, opts tile.TileOptions, query *layer.MapQuery, tileMerger tile.Merger) (tile.Source, error) {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		size, bbox, Srs := query.Size, query.BBox, query.Srs
		return imagery.MergeImages(layers, opt, size, bbox, Srs, tileMerger), nil
	case *terrain.RasterOptions:
		return mergeRasterTile(layers, opts, query), nil
	}
	return nil, errors.New("not support source")
}

func mergeRasterTile(layers []tile.Source, opts tile.TileOptions, query *layer.MapQuery) tile.Source {
	m := terrain.NewRasterMerger([2]int{int(query.MetaSize[0]), int(query.MetaSize[0])}, query.Size)
	m.BBox = query.BBox
	m.BBoxSrs = query.Srs
	return m.Merge(layers, opts.(*terrain.RasterOptions))
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

func splitRasterMetaTiles(meta_tile tile.Source, tiles []geo.TilePattern, tile_size [2]uint32, rasterOpt *terrain.RasterOptions) *TileCollection {
	splitter := terrain.NewRasterSplitter(meta_tile, rasterOpt)
	split_tiles := NewTileCollection(nil)
	for _, tile := range tiles {
		tile_coord, crop_coord := tile.Tiles, tile.Sizes
		if tile_coord[0] < 0 || tile_coord[1] < 0 || tile_coord[2] < 0 {
			continue
		}
		data := splitter.GetSplitTile(crop_coord, tile_size)
		new_tile := NewTile(tile_coord)
		new_tile.SetCacheInfo(meta_tile.GetCacheable())
		new_tile.Source = data
		split_tiles.SetItem(new_tile)
	}
	return split_tiles
}

func SplitTiles(layers tile.Source, tiles []geo.TilePattern, tile_size [2]uint32, opts tile.TileOptions) (*TileCollection, error) {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return splitImageMetaTiles(layers, tiles, tile_size, opt), nil
	case *terrain.RasterOptions:
		return splitRasterMetaTiles(layers, tiles, tile_size, opt), nil
	}
	return nil, errors.New("not support source")
}

func ScaleTiles(layers []tile.Source, queryBBox vec2d.Rect, querySrs geo.Proj, src_tile_grid [2]int, grid *geo.TileGrid, src_bbox vec2d.Rect, opts tile.TileOptions) (tile.Source, error) {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		tiled_image := imagery.NewTiledImage(layers, src_tile_grid, [2]uint32{grid.TileSize[0], grid.TileSize[1]}, src_bbox, grid.Srs)
		return tiled_image.Transform(queryBBox, querySrs, [2]uint32{grid.TileSize[0], grid.TileSize[1]}, opt), nil
	}
	return nil, errors.New("not support source")
}

func MaskImageSourceFromCoverage(source tile.Source, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage, opts tile.TileOptions) (tile.Source, error) {
	switch opt := opts.(type) {
	case *imagery.ImageOptions:
		return imagery.MaskImageSourceFromCoverage(source, bbox, bbox_srs, coverage, opt), nil
	}
	return nil, errors.New("not support source")
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
	data, _ := io.ReadAll(reader)
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
