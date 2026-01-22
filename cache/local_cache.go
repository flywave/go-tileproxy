package cache

import (
	"errors"
	"os"
	"sync"

	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type LocalCache struct {
	Cache
	cacheDir      string
	tileLocation  func(*Tile, string, string, bool) string
	levelLocation func(int, string) string
	creater       tile.SourceCreater
	readBufPool   sync.Pool
}

func NewLocalCache(cache_dir string, directory_layout string, creater tile.SourceCreater) *LocalCache {
	if !utils.FileExists(cache_dir) {
		os.MkdirAll(cache_dir, os.ModePerm)
	}
	c := &LocalCache{
		cacheDir: cache_dir,
		creater:  creater,
	}
	c.tileLocation, c.levelLocation, _ = LocationPaths(directory_layout)
	c.readBufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 256*1024)
		},
	}
	return c
}

func (c *LocalCache) TileLocation(tile *Tile, create_dir bool) string {
	return c.tileLocation(tile, c.cacheDir, c.creater.GetExtension(), create_dir)
}

func (c *LocalCache) LevelLocation(level int) string {
	return c.levelLocation(level, c.cacheDir)
}

func (c *LocalCache) LoadTile(tile *Tile, withMetadata bool) error {
	if !tile.IsMissing() {
		return nil
	}

	location := c.TileLocation(tile, false)

	if utils.FileExists(location) {
		if withMetadata {
			c.LoadTileMetadata(tile)
		}
		data, err := os.ReadFile(location)
		if err != nil {
			return err
		}
		tile.Source = c.creater.Create(data, tile.Coord)
		return nil
	}
	return errors.New("not found")
}

func (c *LocalCache) LoadTiles(tiles *TileCollection, withMetadata bool) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tiles.tiles))
	semaphore := make(chan struct{}, 10)

	for _, tile := range tiles.tiles {
		if !tile.IsMissing() {
			continue
		}

		wg.Add(1)
		go func(t *Tile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			location := c.TileLocation(t, false)
			if !utils.FileExists(location) {
				errChan <- errors.New("not found")
				return
			}

			if withMetadata {
				c.LoadTileMetadata(t)
			}

			buf := c.readBufPool.Get().([]byte)
			defer c.readBufPool.Put(buf)

			data, err := os.ReadFile(location)
			if err != nil {
				errChan <- err
				return
			}

			t.Source = c.creater.Create(data, t.Coord)
		}(tile)
	}

	wg.Wait()
	close(errChan)

	var errs error
	for err := range errChan {
		errs = err
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
	var wg sync.WaitGroup
	errChan := make(chan error, len(tiles.tiles))
	semaphore := make(chan struct{}, 10)

	for _, tile := range tiles.tiles {
		if tile.Stored {
			continue
		}

		wg.Add(1)
		go func(t *Tile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			tile_loc := c.TileLocation(t, true)
			if ok, _ := utils.IsSymlink(tile_loc); ok {
				os.Remove(tile_loc)
			}

			data := t.Source.GetBuffer(nil, nil)
			if err := os.WriteFile(tile_loc, data, os.ModePerm); err != nil {
				errChan <- err
				return
			}
		}(tile)
	}

	wg.Wait()
	close(errChan)

	var errs error
	for err := range errChan {
		errs = err
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
		if utils.FileExists(location) {
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
