package cache

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type LocalCache struct {
	Cache
	cacheDir      string
	tileLocation  TileLocationFunc
	levelLocation func(int, string) string
	creater       tile.SourceCreater
}

func NewLocalCache(cache_dir string, directory_layout string, creater tile.SourceCreater) *LocalCache {
	if !utils.FileExists(cache_dir) {
		if err := os.MkdirAll(cache_dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("failed to create cache directory %s: %v", cache_dir, err))
		}
	}
	c := &LocalCache{
		cacheDir: cache_dir,
		creater:  creater,
	}
	c.tileLocation, c.levelLocation, _ = LocationPaths(directory_layout)
	return c
}

func (c *LocalCache) TileLocation(tile *Tile, create_dir bool) (string, error) {
	return c.tileLocation(tile, c.cacheDir, c.creater.GetExtension(), create_dir)
}

func (c *LocalCache) LevelLocation(level int) string {
	return c.levelLocation(level, c.cacheDir)
}

func (c *LocalCache) LoadTile(tile *Tile, withMetadata bool) error {
	if !tile.IsMissing() {
		return nil
	}

	location, err := c.TileLocation(tile, false)
	if err != nil {
		return err
	}

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
	tileSlice := tiles.GetSliceCopy()
	errChan := make(chan error, len(tileSlice)+100)
	semaphore := make(chan struct{}, 10)
	var notFoundCount int

	for _, tile := range tileSlice {
		if !tile.IsMissing() {
			continue
		}

		wg.Add(1)
		go func(t *Tile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			location, err := c.TileLocation(t, false)
			if err != nil {
				errChan <- err
				return
			}

			if !utils.FileExists(location) {
				errChan <- errors.New("not found")
				return
			}

			if withMetadata {
				c.LoadTileMetadata(t)
			}

			data, err := os.ReadFile(location)
			if err != nil {
				errChan <- fmt.Errorf("io error reading %s: %w", location, err)
				return
			}

			t.mu.Lock()
			t.Source = c.creater.Create(data, t.Coord)
			t.mu.Unlock()
		}(tile)
	}

	wg.Wait()
	close(errChan)

	var ioErrors []string
	for err := range errChan {
		if strings.Contains(err.Error(), "not found") {
			notFoundCount++
		} else {
			ioErrors = append(ioErrors, err.Error())
		}
	}

	if len(ioErrors) > 0 || notFoundCount > 0 {
		if len(ioErrors) > 0 {
			return fmt.Errorf("io errors (%d): %s; not found tiles: %d",
				len(ioErrors), strings.Join(ioErrors, "; "), notFoundCount)
		}
		return fmt.Errorf("not found: %d tiles", notFoundCount)
	}
	return nil
}

func (c *LocalCache) StoreTile(tile *Tile) error {
	if tile.Stored {
		return nil
	}
	tile_loc, err := c.TileLocation(tile, true)
	if err != nil {
		return err
	}
	return c.store(tile, tile_loc)
}

func (c *LocalCache) store(tile *Tile, location string) error {
	data := tile.Source.GetBuffer(nil, nil)
	return os.WriteFile(location, data, 0644)
}

func (c *LocalCache) StoreTiles(tiles *TileCollection) error {
	var wg sync.WaitGroup
	tileSlice := tiles.GetSliceCopy()
	errChan := make(chan error, len(tileSlice)+100)
	semaphore := make(chan struct{}, 10)

	for _, tile := range tileSlice {
		if tile.Stored {
			continue
		}

		wg.Add(1)
		go func(t *Tile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			tile_loc, err := c.TileLocation(t, true)
			if err != nil {
				errChan <- err
				return
			}

			data := t.Source.GetBuffer(nil, nil)
			if err := os.WriteFile(tile_loc, data, 0644); err != nil {
				errChan <- err
				return
			}
			t.Stored = true
		}(tile)
	}

	wg.Wait()
	close(errChan)

	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors encountered (%d): %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}

func (c *LocalCache) RemoveTile(tile *Tile) error {
	location, err := c.TileLocation(tile, false)
	if err != nil {
		return err
	}
	return os.Remove(location)
}

func (c *LocalCache) RemoveTiles(tiles *TileCollection) error {
	var errs []error
	for _, tile := range tiles.GetSliceCopy() {
		if err := c.RemoveTile(tile); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to remove %d tiles: %v", len(errs), errs[0])
	}
	return nil
}

func (c *LocalCache) IsCached(tile *Tile) bool {
	if tile.IsMissing() {
		location, err := c.TileLocation(tile, false)
		if err != nil {
			return false
		}
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
	location, err := c.TileLocation(tile, false)
	if err != nil {
		return err
	}
	stats, err := os.Stat(location)
	if err != nil {
		return err
	}
	tile.Timestamp = stats.ModTime()
	tile.Size = stats.Size()
	return nil
}
