package imagery

import (
	"image"
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"

	"github.com/flywave/gg"
	"github.com/flywave/imaging"
)

type TileMerger struct {
	Grid [2]int
	Size [2]uint32
}

func NewTileMerger(tile_grid [2]int, tile_size [2]uint32) *TileMerger {
	return &TileMerger{Grid: tile_grid, Size: tile_size}
}

func (t *TileMerger) Merge(ordered_tiles []tile.Source, image_opts *ImageOptions) tile.Source {
	if t.Grid[0] == 1 && t.Grid[1] == 1 {
		if len(ordered_tiles) >= 1 && ordered_tiles[0] != nil {
			tile := ordered_tiles[0]
			return tile
		}
	}

	src_size := t.srcSize()

	result := CreateImage(src_size, image_opts)
	dcresult := gg.NewContextForImage(result)

	var cacheable *tile.CacheInfo

	for i, source := range ordered_tiles {
		if source == nil {
			continue
		}

		if source.GetCacheable() == nil {
			cacheable = source.GetCacheable()
		}

		tile := source.GetTile().(image.Image)
		pos := t.tileOffset(i)
		tile = imaging.Resize(tile, int(t.Size[0]), int(t.Size[1]), imaging.Lanczos)
		dcresult.DrawImage(tile, pos[0], pos[1])
	}
	result = dcresult.Image()
	return &ImageSource{image: result, size: src_size[:], Options: image_opts, cacheable: cacheable}
}

func (t *TileMerger) srcSize() [2]uint32 {
	width := uint32(t.Grid[0]) * t.Size[0]
	height := uint32(t.Grid[1]) * t.Size[1]
	return [2]uint32{width, height}
}

func (t *TileMerger) tileOffset(i int) [2]int {
	return [2]int{int(math.Mod(float64(i), float64(t.Grid[0])) * float64(t.Size[0])), int(math.Floor(float64(i)/(float64(t.Grid[0]))) * float64(t.Size[1]))}
}

type TileSplitter struct {
	MetaImage image.Image
	Options   *ImageOptions
}

func NewTileSplitter(meta_tile tile.Source, image_opts *ImageOptions) *TileSplitter {
	return &TileSplitter{MetaImage: meta_tile.GetTile().(image.Image), Options: image_opts}
}

func (t *TileSplitter) GetTile(crop_coord [2]int, tile_size [2]uint32) *ImageSource {
	minx, miny := crop_coord[0], crop_coord[1]
	maxx := minx + int(tile_size[0])
	maxy := miny + int(tile_size[1])

	mrect := t.MetaImage.Bounds()
	var crop image.Image

	if minx < 0 || miny < 0 || maxx > mrect.Dx() || maxy > mrect.Dy() {
		crop = imaging.Crop(t.MetaImage, image.Rect(geo.MaxInt(minx, 0), geo.MaxInt(miny, 0), geo.MinInt(maxx, mrect.Dx()),
			geo.MinInt(maxy, mrect.Dy())))

		result := CreateImage(tile_size, t.Options)
		dcresult := gg.NewContextForImage(result)

		dcresult.DrawImage(crop, geo.AbsInt(geo.MinInt(minx, 0)), geo.AbsInt(geo.MinInt(miny, 0)))

		crop = result
	} else {
		crop = imaging.Crop(t.MetaImage, image.Rect(minx, miny, maxx, maxy))
	}
	return &ImageSource{image: crop, size: tile_size[:], Options: t.Options}
}

type TiledImage struct {
	Tiles    []tile.Source
	TileGrid [2]int
	TileSize [2]uint32
	SrcBBox  vec2d.Rect
	SrcSRS   geo.Proj
}

func NewTiledImage(tiles []tile.Source, tile_grid [2]int, tile_size [2]uint32, src_bbox vec2d.Rect, src_srs geo.Proj) *TiledImage {
	return &TiledImage{Tiles: tiles, TileGrid: tile_grid, TileSize: tile_size, SrcBBox: src_bbox, SrcSRS: src_srs}
}

func (t *TiledImage) GetImage(image_opts *ImageOptions) tile.Source {
	tm := NewTileMerger(t.TileGrid, t.TileSize)
	return tm.Merge(t.Tiles, image_opts)
}

func (t *TiledImage) Transform(req_bbox vec2d.Rect, req_srs geo.Proj, out_size [2]uint32, image_opts *ImageOptions) tile.Source {
	transformer := NewImageTransformer(t.SrcSRS, req_srs, nil)
	src_img := t.GetImage(image_opts)
	return transformer.Transform(src_img, t.SrcBBox, out_size, req_bbox, image_opts)
}

func imageTileOffset(srcbox vec2d.Rect, src_srs geo.Proj, src_size [2]uint32, dst_bbox vec2d.Rect, dst_srs geo.Proj) [2]int {
	if src_srs != nil && dst_srs != nil && !src_srs.Eq(dst_srs) {
		srcbox = src_srs.TransformRectTo(dst_srs, srcbox, 16)
	}

	facx := (dst_bbox.Min[0] - srcbox.Min[0]) / (srcbox.Max[0] - srcbox.Min[0])
	facy := (srcbox.Max[1] - dst_bbox.Max[1]) / (srcbox.Max[1] - srcbox.Min[1])

	return [2]int{int(facx * float64(src_size[0])), int(facy * float64(src_size[1]))}
}
