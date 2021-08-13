package imagery

import (
	"fmt"
	"image"
	"image/color"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"

	"github.com/fogleman/gg"
)

type LayerMerger struct {
	tile.Merger
	Layers    []tile.Source
	Coverages []geo.Coverage
	Cacheable *tile.CacheInfo
}

func (l *LayerMerger) AddSource(src tile.Source, cov geo.Coverage) {
	l.Layers = append(l.Layers, src)
	l.Coverages = append(l.Coverages, cov)
}

func (l *LayerMerger) Merge(opts tile.TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) tile.Source {
	image_opts := opts.(*ImageOptions)
	if l.Layers == nil {
		return NewBlankImageSource([2]uint32{size[0], size[1]}, image_opts, l.Cacheable)
	}

	if len(l.Layers) == 1 {
		layer_img := l.Layers[0]
		var layer_coverage geo.Coverage
		if len(l.Coverages) > 0 {
			layer_coverage = l.Coverages[0]
		}
		layer_opts := layer_img.GetTileOptions().(*ImageOptions)
		if ((layer_opts != nil && layer_opts.Transparent != nil && !*layer_opts.Transparent) || (layer_opts.Transparent != nil && *image_opts.Transparent)) && (size != nil || (size != nil && size[0] == layer_img.GetSize()[0] && size[1] == layer_img.GetSize()[1])) && (layer_coverage != nil || !layer_coverage.IsClip()) && coverage != nil {
			return layer_img
		}
	}

	if size == nil {
		ss := l.Layers[0].GetSize()
		size = ss[:]
	}

	cacheable := l.Cacheable
	result := CreateImage([2]uint32{size[0], size[1]}, image_opts)

	var opacity *float64
	for i := range l.Layers {
		layer_img := l.Layers[i]
		var layer_coverage geo.Coverage
		if len(l.Coverages) > 0 {
			layer_coverage = l.Coverages[i]
		}

		if layer_img.GetCacheable() == nil {
			cacheable = layer_img.GetCacheable()
		}

		var mask *image.Alpha
		img := layer_img.GetTile().(image.Image)
		layer_image_opts := layer_img.GetTileOptions().(*ImageOptions)
		if layer_image_opts == nil || layer_image_opts.Opacity == nil {
			opacity = nil
		} else {
			opacity = layer_image_opts.Opacity
		}

		if layer_coverage != nil && layer_coverage.IsClip() {
			img, mask = maskImage(img, bbox, bbox_srs, layer_coverage)
		}

		if opacity != nil && *opacity > 0 {
			img = opacityAdjust(img, *opacity)
		}

		if opacity != nil && *opacity < 1.0 {
			result = ImagingBlend(result, img, *layer_image_opts.Opacity)
		} else {
			dc := gg.NewContextForImage(result)

			if mask != nil {
				dc.SetMask(mask)
			}

			dc.DrawImage(img, 0, 0)

			result = dc.Image()
		}
	}

	if coverage != nil {
		bg := CreateImage([2]uint32{size[0], size[1]}, image_opts)
		dbg := gg.NewContextForImage(bg)
		img, mask := maskImage(result, bbox, bbox_srs, coverage)
		dbg.SetMask(mask)
		dbg.DrawImage(img, 0, 0)
		result = dbg.Image()
	}

	return &ImageSource{image: result, size: size, Options: image_opts, cacheable: cacheable}
}

func opacityAdjust(m image.Image, percentage float64) *image.NRGBA {
	bounds := m.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	newRgba := image.NewNRGBA(bounds)
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			colorRgb := m.At(i, j)
			r, g, b, a := colorRgb.RGBA()
			opacity := uint16(float64(a) * percentage)
			v := newRgba.ColorModel().Convert(color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: opacity})
			rr, gg, bb, aa := v.RGBA()
			newRgba.SetNRGBA(i, j, color.NRGBA{R: uint8(rr), G: uint8(gg), B: uint8(bb), A: uint8(aa)})
		}
	}
	return newRgba
}

type BandOption struct {
	DstBand int
	SrcImg  int
	SrcBand int
	Factor  float64
}

type BandMerger struct {
	tile.Merger
	Layers       []tile.Source
	Ops          []BandOption
	Cacheable    *tile.CacheInfo
	Mode         ImageMode
	MaxBand      map[int]int
	MaxSrcImages int
}

func NewBandMerger(mode ImageMode) *BandMerger {
	return &BandMerger{Ops: make([]BandOption, 0), Cacheable: nil, Mode: mode, MaxBand: make(map[int]int), MaxSrcImages: 0}
}

func (l *BandMerger) AddSource(src tile.Source, cov geo.Coverage) {
}

func (l *BandMerger) AddOps(dst_band, src_img, src_band int, factor float64) {
	l.Ops = append(l.Ops, BandOption{
		DstBand: dst_band,
		SrcImg:  src_img,
		SrcBand: src_band,
		Factor:  factor,
	})
	if src, ok := l.MaxBand[src_img]; ok {
		l.MaxBand[src_img] = geo.MaxInt(src, src_band)
	} else {
		l.MaxBand[src_img] = src_band
	}
	l.MaxSrcImages = geo.MaxInt(src_img+1, l.MaxSrcImages)
}

func splitImage(img image.Image, mode ImageMode) (cha [][]uint32, rect image.Rectangle) {
	rect = img.Bounds()
	numcha := 1
	switch mode {
	case RGB:
		numcha = 3
		break
	case RGBA:
		numcha = 4
		break
	case GRAY:
		numcha = 1
		break
	}
	si := rect.Bounds().Dx() * rect.Bounds().Dy()
	cha = make([][]uint32, numcha)
	for i := range cha {
		cha[i] = make([]uint32, si)
	}
	for y := 0; y < rect.Bounds().Dy(); y++ {
		for x := 0; x < rect.Bounds().Dx(); x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			if numcha == 4 {
				cha[0][y*rect.Bounds().Dx()+x] = r
				cha[1][y*rect.Bounds().Dx()+x] = g
				cha[2][y*rect.Bounds().Dx()+x] = b
				cha[3][y*rect.Bounds().Dx()+x] = a
			} else if numcha == 3 {
				cha[0][y*rect.Bounds().Dx()+x] = r
				cha[1][y*rect.Bounds().Dx()+x] = g
				cha[2][y*rect.Bounds().Dx()+x] = b
			} else if numcha == 1 {
				cha[0][y*rect.Bounds().Dx()+x] = r
			}
		}
	}
	return
}

func mergeImage(mode ImageMode, rect image.Rectangle, bands [][]uint32) image.Image {
	var out image.Image
	switch mode {
	case RGB:
	case RGBA:
		out = image.NewNRGBA(rect)
		break
	case GRAY:
		out = image.NewGray(rect)
		break
	}
	for y := 0; y < rect.Bounds().Dy(); y++ {
		for x := 0; x < rect.Bounds().Dx(); x++ {
			off := y*rect.Bounds().Dx() + x
			switch outt := out.(type) {
			case *image.NRGBA:
				if len(bands) == 3 {
					c := color.NRGBA{R: uint8(bands[0][off]), G: uint8(bands[1][off]), B: uint8(bands[2][off]), A: 255}
					outt.Set(x, y, c)
				} else if len(bands) == 4 {
					c := color.NRGBA{R: uint8(bands[0][off]), G: uint8(bands[1][off]), B: uint8(bands[2][off]), A: uint8(bands[3][off])}
					outt.Set(x, y, c)
				}
				break
			case *image.Gray:
				c := color.Gray{Y: uint8(bands[3][off])}
				outt.Set(x, y, c)
				break
			}
		}
	}
	return out
}

func (l *BandMerger) Merge(opts tile.TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) tile.Source {
	image_opts := opts.(*ImageOptions)

	if len(l.Layers) < l.MaxSrcImages {
		return NewBlankImageSource([2]uint32{size[0], size[1]}, image_opts, l.Cacheable)
	}

	if size == nil {
		ss := l.Layers[0].GetSize()
		size = ss[:]
	}

	src_img_bands := make([][][]uint32, 0)
	var src_image_rect image.Rectangle
	var bands [][]uint32
	for i, layer := range l.Layers {
		img := layer.GetTile().(image.Image)

		if _, ok := l.MaxBand[i]; !ok {
			src_img_bands = append(src_img_bands, nil)
			continue
		}
		bands, src_image_rect = splitImage(img, l.Mode)
		src_img_bands = append(src_img_bands, bands)
	}

	tmp_mode := l.Mode

	var result_bands [][]uint32

	if tmp_mode == RGBA {
		result_bands = make([][]uint32, 4)
	} else if tmp_mode == RGB {
		result_bands = make([][]uint32, 3)
	} else if tmp_mode == GRAY {
		result_bands = make([][]uint32, 1)
	} else {
		panic(fmt.Sprintf("unsupported destination mode %d", image_opts.Mode))
	}

	for _, op := range l.Ops {
		c := src_img_bands[op.SrcImg][op.SrcBand]
		if op.Factor != 1.0 {
			if result_bands[op.DstBand] == nil {
				result_bands[op.DstBand] = c
			} else {
				for i := range result_bands {
					result_bands[op.DstBand][i] = uint32(float64(result_bands[op.DstBand][i]) * op.Factor)
				}
			}
		} else {
			result_bands[op.DstBand] = c
		}
	}

	blen := 0
	for i := range result_bands {
		if result_bands[i] != nil {
			blen = len(result_bands[i])
			break
		}
	}

	if blen != 0 {
		for i := range result_bands {
			if result_bands[i] == nil {
				result_bands[i] = make([]uint32, blen)
			}
		}
	}

	result := mergeImage(tmp_mode, src_image_rect, result_bands)

	return &ImageSource{image: result, size: size, Options: image_opts, cacheable: l.Cacheable}
}

func MergeImages(layers []tile.Source, image_opts *ImageOptions, size [2]uint32, bbox vec2d.Rect, bbox_srs geo.Proj, merger tile.Merger) tile.Source {
	if merger == nil {
		merger = &LayerMerger{}
	}

	if m, ok := merger.(*BandMerger); ok {
		m.Layers = layers
		return m.Merge(image_opts, size[:], bbox, bbox_srs, nil)
	} else if ml, ok := merger.(*LayerMerger); ok {
		ml.Layers = layers
		return ml.Merge(image_opts, size[:], bbox, bbox_srs, nil)
	}
	return nil
}

func ConcatLegends(legends []tile.Source, mode ImageMode, format tile.TileFormat, size []uint32, bgcolor color.Color, transparent bool) tile.Source {
	if legends == nil {
		return NewBlankImageSource([2]uint32{1, 1}, &ImageOptions{BgColor: bgcolor, Transparent: geo.NewBool(transparent)}, nil)
	}

	if len(legends) == 1 {
		return legends[0]
	}
	legend_position_y := make([]int, 0)

	if size == nil {
		legend_width := 0
		legend_height := 0
		for _, legend := range legends {
			legend_position_y = append(legend_position_y, legend_height)
			tmp_img := legend.GetTile().(image.Image)
			legend_width = geo.MaxInt(legend_width, tmp_img.Bounds().Dx())
			legend_height += tmp_img.Bounds().Dy()
		}

		size = []uint32{uint32(legend_width), uint32(legend_height)}
	}

	var img image.Image
	switch mode {
	case RGB:
	case RGBA:
		img = image.NewNRGBA(image.Rect(0, 0, int(size[0]), int(size[1])))
		break
	case GRAY:
		img = image.NewGray(image.Rect(0, 0, int(size[0]), int(size[1])))
		break
	}
	for y := 0; y < int(size[0]); y++ {
		for x := 0; x < int(size[1]); x++ {
			switch a := img.(type) {
			case *image.Gray:
				a.SetGray(x, y, bgcolor.(color.Gray))
			case *image.NRGBA:
				a.SetNRGBA(x, y, bgcolor.(color.NRGBA))
			}
		}
	}

	dc := gg.NewContextForImage(img)

	for i := range legends {
		legend_img := legends[i].GetTile().(image.Image)
		dc.DrawImage(legend_img, 0, legend_position_y[i])
	}
	return &ImageSource{image: dc.Image(), Options: &ImageOptions{Format: format}}
}
