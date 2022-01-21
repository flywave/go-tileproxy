package imagery

import (
	"image"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
)

func TestTileMerge(t *testing.T) {
	cleanup_tiles := make([]string, 0, 9)
	for i := 0; i < 9; i++ {
		cleanup_tiles = append(cleanup_tiles, createTmpImageFile([2]uint32{100, 100}))
	}
	var tiles []tile.Source
	for i := 0; i < 9; i++ {
		tiles = append(tiles, &ImageSource{fname: cleanup_tiles[i], Options: PNG_FORMAT})
	}
	m := NewTileMerger([2]int{3, 3}, [2]uint32{100, 100})
	result := m.Merge(tiles, PNG_FORMAT)
	img := result.GetTile().(image.Image)

	if img.Bounds().Dx() != 300 || img.Bounds().Dy() != 300 {
		t.FailNow()
	}
}

func TestOneTileMerge(t *testing.T) {
	tiles := []tile.Source{&ImageSource{fname: createTmpImageFile([2]uint32{100, 100}), Options: PNG_FORMAT}}

	m := NewTileMerger([2]int{1, 1}, [2]uint32{100, 100})
	result := m.Merge(tiles, PNG_FORMAT)
	img := result.GetTile().(image.Image)

	if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
		t.FailNow()
	}
}

func TestMissingTileMerge(t *testing.T) {
	tiles := []tile.Source{&ImageSource{fname: createTmpImageFile([2]uint32{100, 100}), Options: PNG_FORMAT}}
	for i := 0; i < 8; i++ {
		tiles = append(tiles, nil)
	}
	m := NewTileMerger([2]int{3, 3}, [2]uint32{100, 100})
	result := m.Merge(tiles, PNG_FORMAT)
	img := result.GetTile().(image.Image)

	if img.Bounds().Dx() != 300 || img.Bounds().Dy() != 300 {
		t.FailNow()
	}
}

func TestTileSplitter(t *testing.T) {
	img := CreateImageSource([2]uint32{356, 266}, PNG_FORMAT)
	splitter := NewTileSplitter(img, PNG_FORMAT)

	tile := splitter.GetTile([2]int{0, 0}, [2]uint32{256, 256})

	if tile.size[0] != 256 || tile.size[1] != 256 {
		t.FailNow()
	}

	tile = splitter.GetTile([2]int{256, 256}, [2]uint32{256, 256})

	if tile.size[0] != 256 || tile.size[1] != 256 {
		t.FailNow()
	}
}

func TestBackgroundLargerCropWithTransparent(t *testing.T) {
	img := CreateImageSource([2]uint32{356, 266}, PNG_FORMAT)
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(true)
	splitter := NewTileSplitter(img, &img_opts)

	tile := splitter.GetTile([2]int{0, 0}, [2]uint32{256, 256})

	if tile.size[0] != 256 || tile.size[1] != 256 {
		t.FailNow()
	}

	tile = splitter.GetTile([2]int{256, 256}, [2]uint32{256, 256})

	if tile.size[0] != 256 || tile.size[1] != 256 {
		t.FailNow()
	}
}
