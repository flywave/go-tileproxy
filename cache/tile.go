package cache

import (
	"errors"
	"image"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
)

type Tile struct {
	Coord     [3]int
	Source    tile.Source
	Location  string
	Stored    bool
	Cacheable bool
	Size      int64
	Timestamp time.Time
}

func NewTile(coord [3]int) *Tile {
	return &Tile{Coord: coord}
}

func (t *Tile) GetCacheInfo() *tile.CacheInfo {
	return &tile.CacheInfo{Cacheable: t.Cacheable, Timestamp: t.Timestamp, Size: t.Size}
}

func (t *Tile) SetCacheInfo(cache *tile.CacheInfo) {
	if cache != nil {
		t.Cacheable = cache.Cacheable
		t.Timestamp = cache.Timestamp
		t.Size = cache.Size
	}
}

func (t *Tile) GetSourceBuffer(format *tile.TileFormat, in_image_opts tile.TileOptions) []byte {
	if t.Source != nil {
		return t.Source.GetBuffer(format, in_image_opts)
	} else {
		return nil
	}
}

func (t *Tile) GetSourceImage() image.Image {
	if t.Source != nil {
		return t.Source.GetTile().(image.Image)
	} else {
		return nil
	}
}

func (t *Tile) GetSource() tile.Source {
	return t.Source
}

func (t *Tile) IsMissing() bool {
	return (t.Source == nil)
}

func (t *Tile) Eq(o *Tile) bool {
	if t.Coord[0] != o.Coord[0] || t.Coord[1] != o.Coord[1] || t.Coord[2] != o.Coord[2] {
		return false
	}
	return true
}

func TransformCoord(coord [3]int, src *geo.TileGrid, dst *geo.TileGrid) ([3]int, error) {
	if src == dst || dst == nil || src == nil {
		return coord, nil
	}
	bbox := src.TileBBox(coord, false)
	_, grids, tiles, _ := dst.GetAffectedTiles(bbox, [2]uint32{dst.TileSize[0], dst.TileSize[1]}, src.Srs)

	if grids != [2]int{1, 1} {
		return [3]int{}, errors.New("BBOX does not align to tile")
	}

	x, y, z, _ := tiles.Next()

	return [3]int{x, y, z}, nil
}
