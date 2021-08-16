package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flywave/go-tileproxy/tile"
)

type testSource struct {
	tile.Source
	data      string
	fileName  string
	cacheable *tile.CacheInfo
	Options   tile.TileOptions
}

func newTestSource(data string, fileName string) *testSource {
	return &testSource{data: data, fileName: fileName}
}

func (s *testSource) GetType() tile.TileType {
	return tile.TILE_VECTOR
}

func (s *testSource) GetSource() interface{} {
	return s.data
}

func (s *testSource) SetSource(src interface{}) {
	s.data = src.(string)
}

func (s *testSource) GetFileName() string {
	return s.fileName
}

func (s *testSource) GetSize() [2]uint32 {
	return [2]uint32{256, 256}
}

func (s *testSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	return []byte(s.data)
}

func (s *testSource) GetTile() interface{} {
	return s.data
}

func (s *testSource) GetCacheable() *tile.CacheInfo {
	return s.cacheable
}

func (s *testSource) SetCacheable(c *tile.CacheInfo) {
	s.cacheable = c
}

func (s *testSource) SetTileOptions(options tile.TileOptions) {
	s.Options = options
}

func (s *testSource) GetTileOptions() tile.TileOptions {
	return s.Options
}

func TestLocalCache(t *testing.T) {
	creater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		return newTestSource(string(data), location)
	}

	c := NewLocalCache("./test_cache", "mvt", "quadkey", creater)

	ts := newTestSource("test", "test.mvt")

	tile := NewTile([3]int{1, 1, 1})
	tile.Source = ts

	c.StoreTile(tile)

	tile2 := NewTile([3]int{1, 1, 1})

	c.LoadTile(tile2, false)

	os.RemoveAll("./test_cache")
}

type testPathData struct {
	key   string
	coord [3]int
	path  string
}

var (
	paths = []testPathData{
		{"mp", [3]int{12345, 67890, 2}, "/tmp/foo/02/0001/2345/0006/7890.png"},
		{"mp", [3]int{12345, 67890, 12}, "/tmp/foo/12/0001/2345/0006/7890.png"},

		{"tc", [3]int{12345, 67890, 2}, "/tmp/foo/02/000/012/345/000/067/890.png"},
		{"tc", [3]int{12345, 67890, 12}, "/tmp/foo/12/000/012/345/000/067/890.png"},

		{"tms", [3]int{12345, 67890, 2}, "/tmp/foo/2/12345/67890.png"},
		{"tms", [3]int{12345, 67890, 12}, "/tmp/foo/12/12345/67890.png"},

		{"quadkey", [3]int{0, 0, 0}, "/tmp/foo/.png"},
		{"quadkey", [3]int{0, 0, 1}, "/tmp/foo/0.png"},
		{"quadkey", [3]int{1, 1, 1}, "/tmp/foo/3.png"},
		{"quadkey", [3]int{12345, 67890, 12}, "/tmp/foo/200200331021.png"},

		{"arcgis", [3]int{1, 2, 3}, "/tmp/foo/L03/R00000002/C00000001.png"},
		{"arcgis", [3]int{9, 2, 3}, "/tmp/foo/L03/R00000002/C00000009.png"},
		{"arcgis", [3]int{10, 2, 3}, "/tmp/foo/L03/R00000002/C0000000a.png"},
		{"arcgis", [3]int{12345, 67890, 12}, "/tmp/foo/L12/R00010932/C00003039.png"},
	}
)

func TestPath(t *testing.T) {
	for _, p := range paths {
		cache := NewLocalCache("/tmp/foo", "png", p.key, nil)

		abs, _ := filepath.Abs(cache.TileLocation(NewTile(p.coord), false))

		if abs != p.path {
			t.FailNow()
		}
	}
}

type testLevelPathData struct {
	key   string
	level int
	path  string
}

var (
	levelPaths = []testLevelPathData{
		{"mp", 2, "/tmp/foo/02"},
		{"mp", 12, "/tmp/foo/12"},

		{"tc", 2, "/tmp/foo/02"},
		{"tc", 12, "/tmp/foo/12"},

		{"tms", 2, "/tmp/foo/02"},
		{"tms", 12, "/tmp/foo/12"},

		{"arcgis", 3, "/tmp/foo/L03"},
		{"arcgis", 3, "/tmp/foo/L03"},
		{"arcgis", 3, "/tmp/foo/L03"},
		{"arcgis", 12, "/tmp/foo/L12"},
	}
)

func TestLevelPath(t *testing.T) {
	for _, p := range levelPaths {
		cache := NewLocalCache("/tmp/foo", "png", p.key, nil)

		abs, _ := filepath.Abs(cache.LevelLocation(p.level))

		if abs != p.path {
			t.FailNow()
		}
	}
}
