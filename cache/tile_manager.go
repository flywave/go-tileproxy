package cache

import (
	"context"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type TileManager struct {
	Manager
	grid                 *geo.TileGrid
	cache                Cache
	identifier           string
	metaGrid             *geo.MetaGrid
	format               string
	tileOpts             tile.TileOptions
	requestFormat        string
	sources              []layer.Layer
	minimizeMetaRequests bool
	expireTimestamp      *time.Time
	preStoreFilter       []Filter
	rescaleTiles         int
	cacheRescaledTiles   bool
	locker               TileLocker
	bulkMetaTiles        bool
	reprojectSrcSrs      geo.Proj
	reprojectDstSrs      geo.Proj
	queryBuffer          *int
	merger               tile.Merger
	siteURL              string
}

type TileManagerOptions struct {
	Sources              []layer.Layer
	Grid                 *geo.TileGrid
	Cache                Cache
	Locker               TileLocker
	Identifier           string
	Format               string
	RequestFormat        string
	Options              tile.TileOptions
	MinimizeMetaRequests bool
	BulkMetaTiles        bool
	PreStoreFilter       []Filter
	RescaleTiles         int
	CacheRescaledTiles   bool
	MetaBuffer           int
	MetaSize             [2]uint32
	ReprojectSrcSrs      geo.Proj
	ReprojectDstSrs      geo.Proj
	QueryBuffer          *int
	SiteURL              string
}

func NewTileManager(opts *TileManagerOptions) *TileManager {
	ret := &TileManager{}
	ret.grid = opts.Grid
	ret.cache = opts.Cache
	ret.identifier = opts.Identifier
	ret.format = opts.Format
	ret.tileOpts = opts.Options
	ret.requestFormat = opts.RequestFormat
	ret.sources = opts.Sources
	ret.minimizeMetaRequests = opts.MinimizeMetaRequests
	ret.preStoreFilter = opts.PreStoreFilter
	ret.rescaleTiles = opts.RescaleTiles
	ret.cacheRescaledTiles = opts.CacheRescaledTiles
	ret.locker = opts.Locker
	ret.bulkMetaTiles = false
	ret.reprojectSrcSrs = opts.ReprojectSrcSrs
	ret.reprojectDstSrs = opts.ReprojectDstSrs
	ret.queryBuffer = opts.QueryBuffer
	ret.siteURL = opts.SiteURL

	if opts.MetaBuffer != -1 || (opts.MetaSize != [2]uint32{1, 1}) {
		allsm := true
		for i := range opts.Sources {
			if !opts.Sources[i].IsSupportMetaTiles() {
				allsm = false
			}
		}
		if allsm {
			ret.metaGrid = geo.NewMetaGrid(ret.grid, opts.MetaSize, opts.MetaBuffer)
		} else if opts.MetaSize != [2]uint32{1, 1} && opts.BulkMetaTiles {
			ret.metaGrid = geo.NewMetaGrid(ret.grid, opts.MetaSize, 0)
			ret.bulkMetaTiles = true
		}
	}
	switch opt := ret.tileOpts.(type) {
	case *imagery.ImageOptions:
		ret.merger = imagery.NewBandMerger(opt.Mode)
	}
	return ret
}

func (tm *TileManager) SiteURL() string {
	return tm.siteURL
}

func (tm *TileManager) GetSources() []layer.Layer {
	return tm.sources
}

func (tm *TileManager) SetSources(layer []layer.Layer) {
	tm.sources = layer
}

func (tm *TileManager) GetGrid() *geo.TileGrid {
	return tm.grid
}

func (tm *TileManager) GetCache() Cache {
	return tm.cache
}

func (tm *TileManager) SetCache(c Cache) {
	tm.cache = c
}

func (tm *TileManager) GetQueryBuffer() *int {
	return tm.queryBuffer
}

func (tm *TileManager) GetReprojectSrcSrs() geo.Proj {
	return tm.reprojectSrcSrs
}

func (tm *TileManager) GetReprojectDstSrs() geo.Proj {
	return tm.reprojectDstSrs
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

func (tm *TileManager) GetTileOptions() tile.TileOptions {
	return tm.tileOpts
}

func (tm *TileManager) SetTileOptions(opt tile.TileOptions) {
	tm.tileOpts = opt
}

func (tm *TileManager) GetFormat() string {
	return tm.format
}

func (tm *TileManager) GetRequestFormat() string {
	return tm.requestFormat
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

func (tm *TileManager) LoadTileCoord(tileCoord [3]int, dimensions utils.Dimensions, with_metadata bool) (*Tile, error) {
	tiles, err := tm.LoadTileCoords([][3]int{tileCoord}, dimensions, with_metadata)
	if err != nil {
		return nil, err
	}
	return tiles.GetItem(0), err
}

func (tm *TileManager) LoadTileCoords(tileCoords [][3]int, dimensions utils.Dimensions, with_metadata bool) (*TileCollection, error) {
	tiles := NewTileCollection(tileCoords)
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

	tiles, err := tm.loadTileCoords(tiles, dimensions, with_metadata, rescale_till_zoom, nil)

	if err != nil {
		return nil, err
	}

	for _, t := range tiles.tiles {
		if t.Source == RESCALE_TILE_MISSING {
			t.Source = nil
		}
	}

	return tiles, nil
}

func (tm *TileManager) loadTileCoords(tiles *TileCollection, dimensions utils.Dimensions, with_metadata bool, rescale_till_zoom int, rescaled_tiles *TileCollection) (*TileCollection, error) {
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
		created_tiles, err := creator.CreateTiles(uncached_tiles)

		if err != nil {
			return nil, err
		}

		if created_tiles == nil && tm.rescaleTiles != -1 {
			created_tiles = make([]*Tile, len(uncached_tiles))
			for i, t := range uncached_tiles {
				created_tiles[i], err = tm.scaledTile(t, rescale_till_zoom, rescaled_tiles)
				if err != nil {
					return nil, err
				}
			}
		}

		for _, created_tile := range created_tiles {
			if tiles.Contains(created_tile.Coord) {
				tiles.GetItem(created_tile.Coord).Source = created_tile.Source
			}
		}
	}

	return tiles, nil
}

var (
	RESCALE_TILE_MISSING = imagery.NewBlankImageSource([2]uint32{256, 256}, &imagery.ImageOptions{}, nil)
)

func (tm *TileManager) scaledTile(t *Tile, stop_zoom int, rescaled_tiles *TileCollection) (*Tile, error) {
	if rescaled_tiles.Contains(t.Coord) {
		return rescaled_tiles.GetItem(t.Coord), nil
	}

	t.Source = RESCALE_TILE_MISSING
	rescaled_tiles.SetItem(t)

	tile_bbox := tm.grid.TileBBox(t.Coord, false)
	current_zoom := t.Coord[2]
	if stop_zoom == current_zoom {
		return t, nil
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

	tile_collection, err := tm.loadTileCoords(
		affected_tiles,
		nil,
		false,
		stop_zoom,
		rescaled_tiles)

	if err != nil {
		return nil, err
	}

	if tile_collection.AllBlank() {
		return t, nil
	}

	tile_sources := []tile.Source{}
	for _, t := range tile_collection.tiles {
		if t.Source != nil && t.Source != RESCALE_TILE_MISSING {
			tile_sources = append(tile_sources, t.Source)
		}
	}

	t.Source, err = ScaleTiles(tile_sources, tile_bbox, tm.grid.Srs, src_tile_grid, tm.grid, src_bbox, tm.tileOpts)

	if err != nil {
		return nil, err
	}

	if tm.cacheRescaledTiles {
		tm.cache.StoreTile(t)
	}
	return t, nil
}

func (tm *TileManager) RemoveTileCoords(tile_coords [][3]int) error {
	tiles := NewTileCollection(tile_coords)
	return tm.cache.RemoveTiles(tiles)
}

func (tm *TileManager) StoreTile(tile *Tile) error {
	return tm.cache.StoreTile(tile)
}

func (tm *TileManager) StoreTiles(tiles *TileCollection) error {
	return tm.cache.StoreTiles(tiles)
}

func (tm *TileManager) IsCached(tile_coord [3]int, dimensions utils.Dimensions) bool {
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

func (tm *TileManager) IsStale(tile_coord [3]int, dimensions utils.Dimensions) bool {
	tile := NewTile(tile_coord)
	if tm.cache.IsCached(tile) {
		return !tm.IsCached(tile_coord, nil)
	}
	return false
}

func (tm *TileManager) ExpireTimestamp(tile *Tile) *time.Time {
	return tm.expireTimestamp
}

func (tm *TileManager) SetExpireTimestamp(t *time.Time) {
	tm.expireTimestamp = t
}

func (tm *TileManager) ApplyTileFilter(tile *Tile) (*Tile, error) {
	if tile.Stored {
		return tile, nil
	}

	var err error
	for _, filter := range tm.preStoreFilter {
		tile, err = filter.Apply(tile)
		if err != nil {
			return nil, err
		}
	}
	return tile, nil
}

func (tm *TileManager) Creator(dimensions utils.Dimensions) *TileCreator {
	return NewTileCreator(tm, dimensions, tm.merger, tm.bulkMetaTiles)
}

func (tm *TileManager) Lock(ctx context.Context, tile *Tile, run func() error) error {
	if tm.metaGrid != nil {
		tile = NewTile(tm.metaGrid.MainTile(tile.Coord))
	}
	return tm.locker.Lock(ctx, tile, run)
}
