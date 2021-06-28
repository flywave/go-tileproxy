package images

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/flywave/go-tileproxy/maths"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	"github.com/flywave/imaging"
	"github.com/fogleman/gg"
)

func stringIn(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func isJpeg(h string) bool {
	return h[6:10] == "JFIF"
}

func isPng(h string) bool {
	return h[:8] == "\211PNG\r\n\032\n"
}

func isGif(h string) bool {
	header := h[:6]
	return stringIn(header, []string{"GIF87a", "GIF89a"})
}

func isTiff(h string) bool {
	return stringIn(h[:2], []string{"MM", "II"})
}

func peekImageFormat(buf string) string {
	if isJpeg(buf) {
		return "jpeg"
	} else if isPng(buf) {
		return "png"
	} else if isGif(buf) {
		return "gif"
	} else if isTiff(buf) {
		return "tiff"
	}
	return ""
}

type GeoReference struct {
	bbox vec2d.Rect
	srs  string
}

func (g *GeoReference) TiePoints() [6]float64 {
	return [6]float64{
		0.0, 0.0, 0.0,
		g.bbox.Min[0], g.bbox.Max[1], 0.0,
	}
}

func (g *GeoReference) PixelScale(img_size [2]int) vec3d.T {
	width := g.bbox.Max[0] - g.bbox.Min[0]
	height := g.bbox.Max[1] - g.bbox.Min[1]
	return vec3d.T{width / float64(img_size[0]), height / float64(img_size[1]), 0.0}
}

type Source interface {
	GetSource() interface{}
	SetSource(src interface{})
	GetFileName() string
	GetSize() [2]uint32
	GetBuffer(format *ImageFormat, in_image_opts *ImageOptions) []byte
	GetImage() image.Image
	GetCacheable() bool
	SetCacheable(c bool)
	GetImageOptions() *ImageOptions
}

type ImageSource struct {
	Source
	image     image.Image
	buf       []byte
	fname     string
	Options   ImageOptions
	size      []uint32
	cacheable bool
	georef    *GeoReference
}

func CreateImageSource(si [2]uint32, opts *ImageOptions) *ImageSource {
	ret := &ImageSource{image: CreateImage(si, opts), Options: *opts}
	return ret
}

func CreateImageSourceFromImage(img image.Image, opts *ImageOptions) *ImageSource {
	ret := &ImageSource{image: img, Options: *opts}
	return ret
}

func (s *ImageSource) GetImageOptions() *ImageOptions {
	return &s.Options
}

func (s *ImageSource) GetCacheable() bool {
	return s.cacheable
}

func (s *ImageSource) SetCacheable(c bool) {
	s.cacheable = c
}

func (s *ImageSource) GetSource() interface{} {
	if s.image != nil {
		return s.image
	} else if len(s.fname) > 0 {
		return s.fname
	}
	return nil
}

func (s *ImageSource) SetSource(src interface{}) {
	s.image = nil
	s.buf = nil
	switch ss := src.(type) {
	case io.Reader:
		s.image, _ = imaging.Decode(ss)
	case image.Image:
		s.image = ss
	case string:
		s.fname = ss
	}
}

func (s *ImageSource) GetFileName() string {
	return s.fname
}

func (s *ImageSource) GetSize() [2]uint32 {
	if s.size == nil {
		s.size = make([]uint32, 2)
		s.size[0] = uint32(s.image.Bounds().Dx())
		s.size[1] = uint32(s.image.Bounds().Dy())
	}
	return [2]uint32{s.size[0], s.size[1]}
}

func (s *ImageSource) makeImageBuf() error {
	if len(s.fname) != 0 {
		f, err := os.Open(s.fname)
		if err != nil {
			return err
		}
		s.buf, err = ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("image name is empty!")
}

func imageToBuf(image image.Image, image_opts ImageOptions, georef *GeoReference) []byte {
	fname := image_opts.Format.Extension()
	buf := &bytes.Buffer{}
	encodeImage(fname, buf, image)
	return buf.Bytes()
}

func (s *ImageSource) GetBuffer(format *ImageFormat, in_image_opts *ImageOptions) []byte {
	var image_opts ImageOptions
	if in_image_opts != nil {
		image_opts = *in_image_opts
	} else {
		image_opts = s.Options
	}
	if format != nil {
		image_opts = s.Options
		image_opts.Format = *format
	}
	if s.buf == nil {
		s.buf = imageToBuf(s.GetImage(), image_opts, s.georef)
		if len(s.Options.Format) == 0 {
			s.Options.Format = ImageFormat(peekImageFormat(string(s.buf)))
		}
		if image_opts.Format != s.Options.Format {
			s.SetSource(s.GetImage())
			s.buf = nil
			s.Options = image_opts
			fname := s.fname
			s.fname = ""
			s.GetBuffer(nil, &image_opts)
			s.fname = fname
		}
	}

	return s.buf
}

func (s *ImageSource) GetImage() image.Image {
	if s.image == nil {
		err := s.makeImageBuf()
		if err != nil {
			return nil
		}
		buf := bytes.NewBuffer(s.buf)
		return decodeImage(s.fname, buf)
	}
	return s.image
}

func encodeImage(inputName string, writer io.Writer, rgba image.Image) {
	if strings.HasSuffix(inputName, "jpg") || strings.HasSuffix(inputName, "jpeg") {
		jpeg.Encode(writer, rgba, nil)
	} else if strings.HasSuffix(inputName, "png") {
		png.Encode(writer, rgba)
	} else if strings.HasSuffix(inputName, "gif") {
		gif.Encode(writer, rgba, nil)
	}
}

func decodeImage(inputName string, reader io.Reader) image.Image {
	if strings.HasSuffix(inputName, "jpg") || strings.HasSuffix(inputName, "jpeg") {
		img, err := jpeg.Decode(reader)
		if err != nil {
			return nil
		}
		return img
	} else if strings.HasSuffix(inputName, "png") {
		img, err := png.Decode(reader)
		if err != nil {
			return nil
		}
		return img
	} else if strings.HasSuffix(inputName, "gif") {
		img, err := gif.Decode(reader)
		if err != nil {
			return nil
		}
		return img
	}
	return nil
}

func SubImageSource(source *ImageSource, size [2]uint32, offset []uint32, image_opts *ImageOptions, cacheable bool) *ImageSource {
	new_image_opts := *image_opts
	new_image_opts.Transparent = newBool(true)
	img := CreateImage(size, &new_image_opts)

	subimg := source.GetImage()

	dc := gg.NewContextForImage(img)

	dc.DrawImage(subimg, int(offset[0]), int(offset[1]))

	return &ImageSource{image: dc.Image(), size: size[:], Options: new_image_opts, cacheable: cacheable}
}

type BlankImageSource struct {
	ImageSource
}

func NewBlankImageSource(size [2]uint32, image_opts *ImageOptions, cacheable bool) *BlankImageSource {
	return &BlankImageSource{ImageSource: ImageSource{size: size[:], Options: *image_opts, image: nil, cacheable: cacheable}}
}

func (s *BlankImageSource) GetFileName() string {
	return ""
}

func (s *BlankImageSource) GetImage() image.Image {
	if s.image == nil {
		s.image = CreateImage([2]uint32{s.size[0], s.size[1]}, &s.Options)
	}
	return s.image
}

func (s *BlankImageSource) GetBuffer(format *ImageFormat, in_image_opts *ImageOptions) []byte {
	if s.buf == nil {
		var image_opts ImageOptions
		if in_image_opts != nil {
			image_opts = *in_image_opts
		} else {
			image_opts = s.Options
		}
		if format != nil {
			image_opts = s.Options
			image_opts.Format = *format
		}
		image_opts.Colors = 0
		s.buf = imageToBuf(s.GetImage(), image_opts, nil)
	}
	return s.buf
}

func bboxPositionInImage(bbox vec2d.Rect, size [2]int, src_bbox vec2d.Rect) ([2]int, [2]int, []float64) {
	coordToPx := maths.MakeLinTransf(bbox, vec2d.Rect{Min: vec2d.T{float64(0), float64(0)}, Max: vec2d.T{float64(size[0]), float64(size[1])}})
	offsets := [4]int{0, size[1], size[0], 0}
	sub_bbox := []float64{bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1]}
	if src_bbox.Min[0] > bbox.Min[0] {
		sub_bbox[0] = src_bbox.Min[0]
		xy := coordToPx([]float64{src_bbox.Min[0], 0})
		offsets[0] = int(xy[0])
	}
	if src_bbox.Min[1] > bbox.Min[1] {
		sub_bbox[1] = src_bbox.Min[1]
		xy := coordToPx([]float64{0, src_bbox.Min[1]})
		offsets[1] = int(xy[1])
	}
	if src_bbox.Max[0] < bbox.Max[0] {
		sub_bbox[2] = src_bbox.Max[0]
		xy := coordToPx([]float64{src_bbox.Max[0], 0})
		offsets[2] = int(xy[0])
	}
	if src_bbox.Max[1] < bbox.Max[1] {
		sub_bbox[3] = src_bbox.Max[1]
		xy := coordToPx([]float64{0, src_bbox.Max[1]})
		offsets[3] = int(xy[1])
	}
	size_ := [2]int{maths.AbsInt(offsets[2] - offsets[0]), maths.AbsInt(offsets[1] - offsets[3])}
	return size_, [2]int{offsets[0], offsets[3]}, sub_bbox
}

var (
	image_filter = map[string]imaging.ResampleFilter{
		"nearest":  imaging.NearestNeighbor,
		"bilinear": imaging.Linear,
		"bicubic":  imaging.Box,
	}
)

func ImagingBlend(imIn1, imIn2 image.Image, alpha float64) image.Image {
	if imIn1.ColorModel() != imIn2.ColorModel() {
		return nil
	}

	if imIn1.Bounds().Dx() != imIn1.Bounds().Dx() || imIn1.Bounds().Dy() != imIn1.Bounds().Dy() {
		return nil
	}

	if alpha == 0.0 {
		return imaging.Clone(imIn1)
	} else if alpha == 1.0 {
		return imaging.Clone(imIn2)
	}

	result := imaging.Clone(imIn1)

	if alpha >= 0 && alpha <= 1.0 {
		switch im1 := imIn1.(type) {
		case *image.NRGBA:
			im2 := imIn2.(*image.NRGBA)
			for i := 0; i < len(im1.Pix); i++ {
				result.Pix[i] = (uint8)(float64(im1.Pix[i]) + alpha*(float64(im2.Pix[i])-float64(im1.Pix[i])))
			}
			break
		}
	} else {
		switch im1 := imIn1.(type) {
		case *image.NRGBA:
			im2 := imIn2.(*image.NRGBA)
			for i := 0; i < len(im1.Pix); i++ {
				temp := (float64(im1.Pix[i]) + alpha*(float64(im2.Pix[i])-float64(im1.Pix[i])))
				if temp <= 0.0 {
					result.Pix[i] = 0
				} else if temp >= 255.0 {
					result.Pix[i] = 255
				} else {
					result.Pix[i] = uint8(temp)
				}
			}
			break
		}
	}
	return result
}
