package cache

import (
	"errors"
	"os"

	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type SourceCreater func(location string) tile.Source

type Cache interface {
	LoadTile(tile *Tile, withMetadata bool) error
	LoadTiles(tiles *TileCollection, withMetadata bool) error
	StoreTile(tile *Tile) error
	StoreTiles(tiles *TileCollection) error
	RemoveTile(tile *Tile) error
	RemoveTiles(tile *TileCollection) error
	IsCached(tile *Tile) bool
	LoadTileMetadata(tile *Tile) error
}

type LocalCache struct {
	Cache
	cacheDir      string
	fileExt       string
	tileLocation  func(*Tile, string, string, bool) string
	levelLocation func(int, string) string
	creater       SourceCreater
}

func NewLocalCache(cache_dir string, file_ext string, directory_layout string, creater SourceCreater) *LocalCache {
	c := &LocalCache{cacheDir: cache_dir, fileExt: file_ext, creater: creater}
	c.tileLocation, c.levelLocation, _ = LocationPaths(directory_layout)
	return c
}

func (c *LocalCache) TileLocation(tile *Tile, create_dir bool) string {
	return c.tileLocation(tile, c.cacheDir, c.fileExt, create_dir)
}

func (c *LocalCache) level_location(level int) string {
	return c.levelLocation(level, c.cacheDir)
}

func (c *LocalCache) LoadTile(tile *Tile, withMetadata bool) error {
	if !tile.IsMissing() {
		return nil
	}

	location := c.TileLocation(tile, false)

	if ok, _ := utils.FileExists(location); ok {
		if withMetadata {
			c.LoadTileMetadata(tile)
		}
		tile.Source = c.creater(location)
		return nil
	}
	return errors.New("not found")
}

func (c *LocalCache) LoadTiles(tiles *TileCollection, withMetadata bool) error {
	var errs error
	for _, tile := range tiles.tiles {
		if err := c.LoadTile(tile, withMetadata); err != nil {
			errs = err
		}
	}
	return errs
}

func (c *LocalCache) StoreTile(tile *Tile) error {
	if tile.Stored {
		return nil
	}
	tile_loc := c.TileLocation(tile, true)
	return c.store(tile, tile_loc)
}

func (c *LocalCache) store(tile *Tile, location string) error {
	if ok, _ := utils.IsSymlink(location); ok {
		os.Remove(location)
	}
	data := tile.Source.GetBuffer(nil, nil)
	return os.WriteFile(location, data, os.ModePerm)
}

func (c *LocalCache) StoreTiles(tiles *TileCollection) error {
	var errs error
	for _, tile := range tiles.tiles {
		if err := c.StoreTile(tile); err != nil {
			errs = err
		}
	}
	return errs
}

func (c *LocalCache) RemoveTile(tile *Tile) error {
	location := c.TileLocation(tile, false)
	return os.Remove(location)
}

func (c *LocalCache) RemoveTiles(tiles *TileCollection) error {
	var errs error
	for _, tile := range tiles.tiles {
		if err := c.RemoveTile(tile); err != nil {
			errs = err
		}
	}
	return errs
}

func (c *LocalCache) IsCached(tile *Tile) bool {
	if tile.IsMissing() {
		location := c.TileLocation(tile, false)
		if ok, _ := utils.FileExists(location); ok {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

func (c *LocalCache) LoadTileMetadata(tile *Tile) error {
	location := c.TileLocation(tile, false)
	stats, err := os.Stat(location)
	if err != nil {
		return err
	}
	tile.Timestamp = stats.ModTime()
	tile.Size = stats.Size()
	return nil
}
