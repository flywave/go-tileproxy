package exports

import (
	"errors"
	"image"

	"github.com/flywave/go-cog"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/tile"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

type CogExport struct {
	Export
	filename string
	box      vec2d.Rect
	optios   tile.TileOptions
	grid     *geo.TileGrid
	layers   map[int]*cog.TileLayer
}

func NewCogExport(filename string, g *geo.TileGrid, optios tile.TileOptions) (*CogExport, error) {
	return &CogExport{}, nil
}

func (e *CogExport) GetTileFormat() tile.TileFormat {
	return e.optios.GetFormat()
}

func (e *CogExport) StoreTile(t *cache.Tile, srcGrid *geo.TileGrid) error {
	dc, err := cache.TransformCoord(t.Coord, srcGrid, e.grid)

	if err != nil {
		return err
	}

	dstTile := *t
	dstTile.Coord = dc

	if _, ok := e.layers[dc[2]]; !ok {
		e.layers[dc[2]] = cog.NewTileLayer(e.box, dc[2], e.grid)
	}

	switch e.optios.(type) {
	case *imagery.ImageOptions:
		data := dstTile.Source.GetTile().(image.Image)
		e.layers[dc[2]].SetSource(dc, cog.NewSource(data, nil, cog.CTLZW))
	case *terrain.RasterOptions:
		data := dstTile.Source.GetTile().(*terrain.TileData)
		rect := image.Rect(0, 0, int(data.Size[0]), int(data.Size[1]))
		e.layers[dc[2]].SetSource(dc, cog.NewSource(data.Datas, &rect, cog.CTLZW))
	default:
		return errors.New("not support format")
	}

	return nil
}

func (e *CogExport) StoreTileCollection(ts *cache.TileCollection, srcGrid *geo.TileGrid) error {
	for _, t := range ts.GetSlice() {
		if err := e.StoreTile(t, srcGrid); err != nil {
			return err
		}
	}
	return nil
}

func (e *CogExport) Close() error {
	lays := []*cog.TileLayer{}
	for i := range e.layers {
		if e.layers[i].Valid() {
			lays = append(lays, e.layers[i])
		}
	}

	if len(lays) == 0 {
		return errors.New("not valid tilelayers")
	}

	err := cog.Write(e.filename, lays, false)

	if err != nil {
		return err
	}
	return nil
}
