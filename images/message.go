package images

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"path"
	"strings"

	"github.com/flywave/go-tileproxy/geo"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"golang.org/x/image/font/gofont/goregular"
)

var (
	font_paths string
)

func init() {
	font_paths = "./fonts"
}

func SetFontPath(p string) {
	font_paths = p
}

func newBool(v bool) *bool {
	return &v
}

func GenMessageImage(message string, size [2]uint32, image_opts *ImageOptions, bgcolor color.Color,
	transparent bool) {
	eimg := NewExceptionImage(message, image_opts)
	eimg.Draw(nil, size[:], true)
}

func GenAttributionImage(message string, size [2]uint32, image_opts *ImageOptions, inverse bool) {
	if image_opts == nil {
		image_opts = &ImageOptions{Transparent: newBool(true)}
	}

	aimg := NewAttributionImage(message, image_opts, &inverse)
	aimg.Draw(nil, size[:], true)
}

type MessageImage struct {
	message     string
	image_opts  *ImageOptions
	font_name   string
	font_size   int
	font_color  color.Color
	box_color   color.Color
	linespacing int
	padding     int
	placement   string
	font_face   font.Face
}

func loadDefaultFontFace() font.Face {
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: 12,
	})
	return face
}

func newMessageImage(message string, image_opts *ImageOptions) *MessageImage {
	ret := &MessageImage{}
	ret.font_name = "DejaVu Sans Mono"
	ret.font_size = 10
	ret.font_color = color.Black
	ret.box_color = nil
	ret.linespacing = 5
	ret.padding = 3
	ret.placement = "ul"
	ret.message = message
	ret.image_opts = image_opts
	ret.font_face = nil
	return ret
}

func (m *MessageImage) GetFont() font.Face {
	if m.font_face == nil {
		if m.font_name != "default" {
			var err error
			m.font_face, err = gg.LoadFontFace(fontFile(m.font_name), float64(m.font_size))
			if err != nil {
				return nil
			}
		}
		if m.font_face == nil {
			m.font_face = loadDefaultFontFace()
		}
	}
	return m.font_face
}

func (m *MessageImage) newImage(size []uint32) image.Image {
	return image.NewNRGBA(image.Rect(0, 0, int(size[0]), int(size[1])))
}

func (m *MessageImage) Draw(img Source, size []uint32, in_place bool) (Source, error) {
	if !((img != nil && size == nil) || (size != nil && img == nil)) {
		return nil, errors.New("need either img or size argument")
	}
	var base_img image.Image
	if img == nil {
		base_img = m.newImage(size)
	} else if !in_place {
		ss := img.GetSize()
		size = ss[:]
		base_img = m.newImage(size)
	} else {
		base_img = img.GetImage()
		size = []uint32{uint32(base_img.Bounds().Dx()), uint32(base_img.Bounds().Dy())}
	}

	if m.message == "" {
		if img != nil {
			return img, nil
		}
		return &ImageSource{image: base_img, size: size, Options: *m.image_opts}, nil
	}

	draw := gg.NewContextForImage(base_img)
	m.drawMsg(draw, size)
	image_opts := m.image_opts
	base_img = draw.Image()

	if !in_place && img != nil {
		if image_opts == nil && img.GetImageOptions() != nil {
			image_opts = img.GetImageOptions()
		}
		img := img.GetImage()

		imgd := gg.NewContextForImage(img)
		imgd.DrawImage(base_img, 0, 0)
		base_img = imgd.Image()
	}

	return &ImageSource{image: base_img, size: size, Options: *m.image_opts}, nil
}

func (m *MessageImage) drawMsg(draw *gg.Context, size []uint32) {
	td := NewTextDraw(m.message, m.GetFont(), m.font_color,
		m.box_color, m.placement, m.padding,
		m.linespacing)
	td.draw(draw, size)
}

type ExceptionImage struct {
	MessageImage
}

func NewExceptionImage(message string, image_opts *ImageOptions) *ExceptionImage {
	ret := &ExceptionImage{MessageImage: *newMessageImage(message, image_opts)}
	if ret.image_opts.BgColor == nil {
		ret.image_opts.BgColor = color.White
	}
	return ret
}

func (e *ExceptionImage) GetFontColor() color.Color {
	if e.image_opts.Transparent != nil && *e.image_opts.Transparent {
		return color.Black
	}

	if _luminance(e.image_opts.BgColor) < 128 {
		return color.White
	}
	return color.Black
}

type WatermarkImage struct {
	MessageImage
}

func NewWatermarkImage(message string, image_opts *ImageOptions, placement string, in_opacity *float64, font_color *color.Color, font_size *int) *WatermarkImage {
	ret := &WatermarkImage{MessageImage: *newMessageImage(message, image_opts)}
	var opacity float64
	if in_opacity == nil {
		opacity = 30
	} else {
		opacity = *in_opacity
	}

	if font_size != nil {
		ret.font_size = *font_size
	} else {
		ret.font_size = 24
	}

	if font_color != nil {
		ret.font_color = *font_color
	} else {
		ret.font_color = color.NRGBA{R: 128, G: 128, B: 128, A: 30}
	}

	switch c := ret.font_color.(type) {
	case *color.NRGBA:
		c.A = uint8(opacity)
		break
	case *color.NRGBA64:
		c.A = uint16(opacity)
		break
	}

	ret.placement = placement
	return ret
}

func (e *WatermarkImage) drawMsg(draw *gg.Context, size []uint32) {
	td := NewTextDraw(e.message, e.font_face, e.font_color, nil, "", 3, 5)
	if strings.ContainsAny(e.placement, "lb") {
		td.placement = "cL"
		td.draw(draw, size)
	}
	if strings.ContainsAny(e.placement, "rb") {
		td.placement = "cR"
		td.draw(draw, size)
	}
	if strings.Contains(e.placement, "c") {
		td.placement = "cc"
		td.draw(draw, size)
	}
}

type AttributionImage struct {
	MessageImage
	Inverse bool
}

func NewAttributionImage(message string, image_opts *ImageOptions, inverse *bool) *AttributionImage {
	ret := &AttributionImage{MessageImage: *newMessageImage(message, image_opts)}
	if inverse != nil {
		ret.Inverse = *inverse
	}
	ret.font_size = 10
	ret.placement = "lr"
	return ret
}

func (e *AttributionImage) GetFontColor() color.Color {
	if e.Inverse {
		return color.White
	} else {
		return color.Black
	}
}

func (e *AttributionImage) GetBoxColor() color.Color {
	if e.Inverse {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	} else {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 120}
	}
}

type TextDraw struct {
	text        []string
	font        font.Face
	bg_color    color.Color
	font_color  color.Color
	placement   string
	padding     int
	linespacing int
}

func NewTextDraw(text string, font font.Face, font_color color.Color, bg_color color.Color,
	placement string, padding int, linespacing int) *TextDraw {
	texts := strings.Split(text, "\n")
	return &TextDraw{text: texts, font: font, bg_color: bg_color, font_color: font_color, placement: placement, padding: padding, linespacing: linespacing}
}

func (m *TextDraw) textBoxes(draw *gg.Context, size []uint32) ([4]int, [][4]int) {
	total_bbox, boxes := m.relativeTextBoxes(draw)
	return m.placeBoxes(total_bbox, boxes, size)
}

func (m *TextDraw) draw(ggctx *gg.Context, size []uint32) {
	total_bbox, boxes := m.textBoxes(ggctx, size)
	if m.bg_color != nil {
		ggctx.SetColor(m.bg_color)
		w, h := total_bbox[2]-total_bbox[0], total_bbox[3]-total_bbox[1]
		ggctx.DrawRectangle(float64(total_bbox[0]),
			float64(total_bbox[1]),
			float64(w),
			float64(h))
		ggctx.Fill()
	}
	ggctx.SetColor(m.font_color)
	ggctx.SetFontFace(m.font)

	for i, text := range m.text {
		_, h := boxes[i][2]-boxes[i][0], boxes[i][3]-boxes[i][1]
		ggctx.DrawString(text, float64(boxes[i][0]), float64(boxes[i][1]+h))
	}
}

func (m *TextDraw) relativeTextBoxes(draw *gg.Context) ([4]int, [][4]int) {
	total_bbox := [4]int{1e9, 1e9, -1e9, -1e9}
	boxes := make([][4]int, 0)
	y_offset := 0
	for _, line := range m.text {
		text_w, text_h := draw.MeasureString(line)
		text_box := [4]int{0, y_offset, int(text_w), int(text_h) + y_offset}
		boxes = append(boxes, text_box)
		total_bbox = [4]int{
			geo.MinInt(total_bbox[0], text_box[0]),
			geo.MinInt(total_bbox[1], text_box[1]),
			geo.MaxInt(total_bbox[2], text_box[2]),
			geo.MaxInt(total_bbox[3], text_box[3]),
		}

		y_offset += int(text_h) + m.linespacing
	}
	return total_bbox, boxes
}

func (m *TextDraw) placeBoxes(total_bbox [4]int, boxes [][4]int, size []uint32) ([4]int, [][4]int) {
	x_offset := -1
	y_offset := -1
	text_size := [2]int{(total_bbox[2] - total_bbox[0]), (total_bbox[3] - total_bbox[1])}

	if m.placement[0] == 'u' {
		y_offset = m.padding
	} else if m.placement[0] == 'l' {
		y_offset = int(size[1]) - m.padding - text_size[1]
	} else if m.placement[0] == 'c' {
		y_offset = int(size[1]/2) - int(text_size[1]/2)
	}

	if m.placement[1] == 'l' {
		x_offset = m.padding
	}
	if m.placement[1] == 'L' {
		x_offset = -int(text_size[0] / 2)
	} else if m.placement[1] == 'r' {
		x_offset = int(size[0]) - m.padding - text_size[0]
	} else if m.placement[1] == 'R' {
		x_offset = int(size[0]) - int(text_size[0]/2)
	} else if m.placement[1] == 'c' {
		x_offset = int(size[0]/2) - (text_size[0] / 2)
	}

	if x_offset == -1 || y_offset == -1 {
		panic(fmt.Sprintf("placement %s not supported", m.placement))
	}

	offsets := [2]int{x_offset, y_offset}
	return m.moveBBoxes([][4]int{total_bbox}, offsets)[0], m.moveBBoxes(boxes, offsets)
}

func (m *TextDraw) moveBBoxes(boxes [][4]int, offsets [2]int) [][4]int {
	result := make([][4]int, 0)
	for _, box := range boxes {
		box = [4]int{box[0] + offsets[0], box[1] + offsets[1], box[2] + offsets[0], box[3] + offsets[1]}
		result = append(result, box)
	}
	return result
}

func fontFile(font string) string {
	font = strings.Replace(font, " ", "", -1)
	if !strings.HasSuffix(font, ".ttf") {
		font += ".ttf"
	}
	return path.Join(font_paths, font)
}

func _luminance(color color.Color) float64 {
	r, g, b, _ := color.RGBA()
	return float64(r)*299/1000 + float64(g)*587/1000 + float64(b)*114/1000
}
