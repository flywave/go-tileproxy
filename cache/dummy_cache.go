package cache

import "errors"

type DummyCache struct {
}

func (d *DummyCache) LoadTile(tile *Tile, withMetadata bool) error {
	return errors.New("not found")
}

func (d *DummyCache) LoadTiles(tiles *TileCollection, withMetadata bool) error {
	return errors.New("not found")
}

func (d *DummyCache) StoreTile(tile *Tile) error {
	return nil
}

func (d *DummyCache) StoreTiles(tiles *TileCollection) error {
	return nil
}

func (d *DummyCache) RemoveTile(tile *Tile) error {
	return nil
}

func (d *DummyCache) RemoveTiles(tile *TileCollection) error {
	return nil
}

func (d *DummyCache) IsCached(tile *Tile) bool {
	return false
}

func (d *DummyCache) LoadTileMetadata(tile *Tile) error {
	return nil
}

func (d *DummyCache) LevelLocation(level int) string {
	return ""
}
