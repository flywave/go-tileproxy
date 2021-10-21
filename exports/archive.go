package exports

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
	_ "github.com/mholt/archiver/v3"
)

const (
	TARGZ = ".tar.gz"
	ZIP   = ".zip"
)

type ArchiveOptions struct {
	DirectoryLayout  string
	ArchiveExt       string
	CompressionLevel int
}

type ArchiveExport struct {
	ExportIO
	Format *tile.TileFormat
	Grid   geo.Grid
}

func NewArchiveExport(g geo.Grid) *ArchiveExport {
	return &ArchiveExport{Grid: g}
}

func (a *ArchiveExport) GetTileFormat() *tile.TileFormat {
	return a.Format
}

func (a *ArchiveExport) StoreTile(t *cache.Tile) error {
	return nil
}

func (a *ArchiveExport) StoreTileCollection(ts *cache.TileCollection) error {
	return nil
}
