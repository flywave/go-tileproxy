package cache

import (
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
)

type Manager interface {
	GetGrid() *geo.TileGrid
	GetCache() Cache
	GetMetaGrid() *geo.MetaGrid
	GetImageOptions() *images.ImageOptions
	Cleanup() bool
	GetRequestFormat() string
	SetMinimizeMetaRequests(f bool)
	GetMinimizeMetaRequests() bool
	GetRescaleTiles() int
	LoadTileCoord(tile_coord [3]int, dimensions map[string]string, with_metadata bool) (error, *Tile)
	LoadTileCoords(tile_coords [][3]int, dimensions map[string]string, with_metadata bool) (error, *TileCollection)
	RemoveTileCoords(tile_coord [][3]int) error
	IsCached(tile_coord [3]int, dimensions map[string]string) bool
	IsStale(tile_coord [3]int, dimensions map[string]string) bool
	ExpireTimestamp(tile *Tile) *time.Time
	ApplyTileFilter(tile *Tile) *Tile
	Creator(dimensions map[string]string) *TileCreator
	Lock(tile *Tile, run func() error) error
}

type TileManager struct {
	Manager
	grid                 *geo.TileGrid
	cache                Cache
	identifier           string
	metaGrid             *geo.MetaGrid
	format               string
	imageOpts            *images.ImageOptions
	requestFormat        string
	sources              []layer.MapLayer
	minimizeMetaRequests bool
	expireTimestamp      *time.Time
	preStoreFilter       []Filter
	rescaleTiles         int
	cacheRescaledTiles   bool
	locker               TileLocker
}

func NewTileManager(sources []layer.MapLayer, grid *geo.TileGrid, cache Cache, identifier string, format string, opts *images.ImageOptions, minimize_meta_requests bool, pre_store_filter []Filter, rescale_tiles int, cache_rescaled_tiles bool, metaBuffer int, metaSize [2]uint32) *TileManager {
	ret := &TileManager{}
	ret.grid = grid
	ret.cache = cache
	ret.identifier = identifier
	ret.format = format
	ret.imageOpts = opts
	ret.requestFormat = format
	ret.sources = sources
	ret.minimizeMetaRequests = minimize_meta_requests
	ret.preStoreFilter = pre_store_filter
	ret.rescaleTiles = rescale_tiles
	ret.cacheRescaledTiles = cache_rescaled_tiles

	if metaBuffer != -1 || (metaSize != [2]uint32{1, 1}) {
		allsm := true
		for i := range sources {
			if !sources[i].SupportMetaTiles {
				allsm = false
			}
		}
		if allsm {
			ret.metaGrid = geo.NewMetaGrid(ret.grid, metaSize, metaBuffer)
		}
	}
	return ret
}

func (tm *TileManager) GetGrid() *geo.TileGrid {
	return tm.grid
}

func (tm *TileManager) GetCache() Cache {
	return tm.cache
}

func (tm *TileManager) GetMetaGrid() *geo.MetaGrid {
	return tm.metaGrid
}

func (tm *TileManager) Cleanup() bool {
	if xw, ok := tm.cache.(interface {
		Cleanup() bool
	}); ok {
		return xw.Cleanup()
	}
	return false
}

func (tm *TileManager) GetImageOptions() *images.ImageOptions {
	return tm.imageOpts
}

func (tm *TileManager) GetRequestFormat() string {
	return tm.format
}

func (tm *TileManager) SetMinimizeMetaRequests(f bool) {
	tm.minimizeMetaRequests = f
}

func (tm *TileManager) GetMinimizeMetaRequests() bool {
	return tm.minimizeMetaRequests
}

func (tm *TileManager) GetRescaleTiles() int {
	return tm.rescaleTiles
}

func (tm *TileManager) LoadTileCoord(tile_coord [3]int, dimensions map[string]string, with_metadata bool) (error, *Tile) {
	err, tiles := tm.LoadTileCoords([][3]int{tile_coord}, dimensions, with_metadata)
	return err, tiles.GetItem(0)
}

func (tm *TileManager) LoadTileCoords(tile_coords [][3]int, dimensions map[string]string, with_metadata bool) (error, *TileCollection) {
	tiles := NewTileCollection(tile_coords)
	rescale_till_zoom := 0

	if tm.rescaleTiles != -1 {
		for _, tile := range tiles.tiles {
			rescale_till_zoom = tile.Coord[2] + tm.rescaleTiles
			break
		}

		if rescale_till_zoom < 0 {
			rescale_till_zoom = 0
		}
		if rescale_till_zoom > int(tm.grid.Levels) {
			rescale_till_zoom = int(tm.grid.Levels)
		}
	}

	tiles = tm.loadTileCoords(
		tiles, dimensions, with_metadata,
		rescale_till_zoom, nil)

	for _, t := range tiles.tiles {
		if t.Source == RESCALE_TILE_MISSING {
			t.Source = nil
		}
	}

	return nil, tiles
}

func (tm *TileManager) loadTileCoords(tiles *TileCollection, dimensions map[string]string, with_metadata bool, rescale_till_zoom int, rescaled_tiles *TileCollection) *TileCollection {
	uncached_tiles := []*Tile{}

	if rescaled_tiles != nil {
		for _, t := range tiles.tiles {
			if rescaled_tiles.Contains(t.Coord) {
				t.Source = rescaled_tiles.GetItem(t.Coord).Source
			}
		}
	}

	tm.cache.LoadTiles(tiles, with_metadata)

	for _, tile := range tiles.tiles {
		if !tm.IsCached(tile.Coord, dimensions) {
			uncached_tiles = append(uncached_tiles, tile)
		}
	}

	if len(uncached_tiles) > 0 {
		creator := tm.Creator(dimensions)
		created_tiles := creator.CreateTiles(uncached_tiles)
		if created_tiles == nil && tm.rescaleTiles != -1 {
			created_tiles = make([]*Tile, len(uncached_tiles))
			for i, t := range uncached_tiles {
				created_tiles[i] = tm.scaledTile(t, rescale_till_zoom, rescaled_tiles)
			}
		}

		for _, created_tile := range created_tiles {
			if tiles.Contains(created_tile.Coord) {
				tiles.GetItem(created_tile.Coord).Source = created_tile.Source
			}
		}
	}

	return tiles
}

var (
	RESCALE_TILE_MISSING = images.NewBlankImageSource([2]uint32{256, 256}, &images.ImageOptions{}, false)
)

func (tm *TileManager) scaledTile(tile *Tile, stop_zoom int, rescaled_tiles *TileCollection) *Tile {
	if rescaled_tiles.Contains(tile.Coord) {
		return rescaled_tiles.GetItem(tile.Coord)
	}

	tile.Source = RESCALE_TILE_MISSING
	rescaled_tiles.SetItem(tile)

	tile_bbox := tm.grid.TileBBox(tile.Coord, false)
	current_zoom := tile.Coord[2]
	if stop_zoom == current_zoom {
		return tile
	}
	var src_level int
	if stop_zoom > current_zoom {
		src_level = current_zoom + 1
	} else {
		src_level = current_zoom - 1
	}

	src_bbox, src_tile_grid, affected_tile_coords, _ := tm.grid.GetAffectedLevelTiles(tile_bbox, src_level)

	coords := [][3]int{}
	x, y, zoom, done := affected_tile_coords.Next()
	for !done {
		coords = append(coords, [3]int{x, y, zoom})
		x, y, zoom, done = affected_tile_coords.Next()
	}

	affected_tiles := NewTileCollection(coords)
	for _, t := range affected_tiles.tiles {
		if rescaled_tiles.Contains(t.Coord) {
			t.Source = rescaled_tiles.GetItem(t.Coord).Source
		}
	}

	tile_collection := tm.loadTileCoords(
		affected_tiles,
		nil,
		false,
		stop_zoom,
		rescaled_tiles)

	if tile_collection.Empty() {
		return tile
	}

	tile_sources := []images.Source{}
	for _, t := range tile_collection.tiles {
		if t.Source != nil && t.Source != RESCALE_TILE_MISSING {
			tile_sources = append(tile_sources, t.Source)
		}
	}

	tiled_image := images.NewTiledImage(tile_sources, src_tile_grid, [2]uint32{tm.grid.TileSize[0], tm.grid.TileSize[1]}, src_bbox, tm.grid.Srs)
	tile.Source = tiled_image.Transform(tile_bbox, tm.grid.Srs, [2]uint32{tm.grid.TileSize[0], tm.grid.TileSize[1]}, tm.imageOpts)

	if tm.cacheRescaledTiles {
		tm.cache.StoreTile(tile)
	}
	return tile
}

func (tm *TileManager) RemoveTileCoords(tile_coords [][3]int) error {
	tiles := NewTileCollection(tile_coords)
	return tm.cache.RemoveTiles(tiles)
}

func (tm *TileManager) IsCached(tile_coord [3]int, dimensions map[string]string) bool {
	tile := NewTile(tile_coord)
	cached := tm.cache.IsCached(tile)
	max_mtime := tm.ExpireTimestamp(tile)
	if cached && max_mtime != nil {
		tm.cache.LoadTileMetadata(tile)
		stale := tile.Timestamp.Before(*max_mtime)
		if stale {
			cached = false
		}
	}
	return cached
}

func (tm *TileManager) IsStale(tile_coord [3]int, dimensions map[string]string) bool {
	tile := NewTile(tile_coord)
	if tm.cache.IsCached(tile) {
		if !tm.IsCached(tile_coord, nil) {
			return true
		}
		return false
	}
	return false
}

func (tm *TileManager) ExpireTimestamp(tile *Tile) *time.Time {
	return tm.expireTimestamp
}

func (tm *TileManager) ApplyTileFilter(tile *Tile) *Tile {
	if tile.Stored {
		return tile
	}

	for _, filter := range tm.preStoreFilter {
		tile = filter.Apply(tile)
	}
	return tile
}

func (tm *TileManager) Creator(dimensions map[string]string) *TileCreator {
	return NewTileCreator(tm, dimensions, nil, false)
}

func (tm *TileManager) Lock(tile *Tile, run func() error) error {
	if tm.metaGrid != nil {
		tile = NewTile(tm.metaGrid.MainTile(tile.Coord))
	}
	return tm.locker.Lock(tile, run)
}
