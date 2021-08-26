package imagery

import (
	"image"
	"image/color"

	"github.com/flywave/go-tileproxy/tile"
)

type ImageMode uint32

const (
	AUTO ImageMode = 0
	RGB  ImageMode = 1
	RGBA ImageMode = 2
	GRAY ImageMode = 3
)

func ImageModeFromString(m string) ImageMode {
	if m == "rgb" {
		return RGB
	} else if m == "rgba" {
		return RGBA
	} else if m == "gray" {
		return GRAY
	}
	return AUTO
}

type ImageOptions struct {
	Transparent     *bool
	Opacity         *float64
	Format          tile.TileFormat
	Resampling      string
	Mode            ImageMode
	BgColor         color.Color
	Colors          int
	EncodingOptions map[string]interface{}
}

func (o *ImageOptions) GetFormat() tile.TileFormat {
	return o.Format
}

func CreateImage(size [2]uint32, image_opts *ImageOptions) image.Image {
	var mode ImageMode
	var bgcolor color.Color
	if image_opts == nil {
		mode = RGB
		bgcolor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	} else {
		mode = image_opts.Mode
		if mode == AUTO {
			if image_opts.Transparent != nil && *image_opts.Transparent {
				mode = RGBA
			} else {
				mode = RGB
			}
		}

		if image_opts.BgColor != nil {
			bgcolor = image_opts.BgColor
		} else {
			bgcolor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
	}
	r := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: int(size[0]), Y: int(size[1])}}
	img := image.NewNRGBA(r)

	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			img.Set(x, y, bgcolor)
		}
	}
	return img
}

func CompatibleImageOptions(img_opts []ImageOptions, base_opts *ImageOptions) ImageOptions {
	colors := 0
	for i := range img_opts {
		if img_opts[i].Colors != 0 {
			if img_opts[i].Colors > colors {
				colors = img_opts[i].Colors
			}
		}
	}

	transparent := false
	for _, o := range img_opts {
		if o.Transparent != nil {
			transparent = *o.Transparent
			break
		}
	}
	mode := AUTO

	for i := range img_opts {
		if img_opts[i].Mode != AUTO {
			if img_opts[i].Mode > mode {
				mode = img_opts[i].Mode
			}
		}
	}
	var options ImageOptions
	if base_opts != nil {
		options = *base_opts
		if options.Colors != 0 {
			options.Colors = colors
		}
		if options.Mode != AUTO {
			options.Mode = mode
		}
		if options.Transparent != nil {
			options.Transparent = &transparent
		}
	} else {
		options = img_opts[0]
		options.Colors = colors
		options.Transparent = &transparent
		options.Mode = mode
	}

	return options
}
