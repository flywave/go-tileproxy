package sources

import (
	"bytes"
	"errors"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
)

type TileSource struct {
	ImagerySource
	Grid         *geo.TileGrid
	Client       *client.TileClient
	ImageOpts    *images.ImageOptions
	Coverage     geo.Coverage
	Extent       *geo.MapExtent
	ResRange     *geo.ResolutionRange
	ErrorHandler func(error)
}

func (s *TileSource) GetMap(query *layer.MapQuery) (images.Source, error) {
	if s.Grid.TileSize[0] != query.Size[0] || s.Grid.TileSize[1] != query.Size[1] {
		return nil, errors.New("tile size of cache and tile source do not match")
	}

	if !s.Grid.Srs.Eq(query.Srs) {
		return nil, errors.New("SRS of cache and tile source do not match")
	}

	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size,
		query.Srs) {
		return images.NewBlankImageSource(query.Size, s.ImageOpts, false), nil
	}
	if s.Coverage != nil && !s.Coverage.Intersects(query.BBox, query.Srs) {
		return images.NewBlankImageSource(query.Size, s.ImageOpts, false), nil
	}

	_, grid, tiles, err := s.Grid.GetAffectedTiles(query.BBox, query.Size, nil)

	if err != nil {
		return nil, err
	}

	if grid != [2]int{1, 1} {
		return nil, errors.New("BBOX does not align to tile")
	}

	x, y, z, _ := tiles.Next()

	resp := s.Client.GetTile([3]int{x, y, z}, &query.Format)
	src := images.CreateImageSource(query.Size, s.ImageOpts)
	src.SetSource(bytes.NewBuffer(resp))
	return src, nil
}
