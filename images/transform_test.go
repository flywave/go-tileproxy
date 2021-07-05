package images

import (
	"fmt"
	"testing"

	"github.com/flywave/go-tileproxy/geo"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/imaging"
)

func TestImageTransform(t *testing.T) {
	img, _ := imaging.Open("./flowers.png")
	img_opts := *PNG_FORMAT
	img_opts.Transparent = geo.NewBool(false)

	src_img := CreateImageSourceFromImage(img, &img_opts)
	src_srs := geo.NewSRSProj4("EPSG:31467")
	dst_size := [2]uint32{100, 150}
	dst_srs := geo.NewSRSProj4("EPSG:4326")
	dst_bbox := vec2d.Rect{Min: vec2d.T{0.2, 45.1}, Max: vec2d.T{8.3, 53.2}}
	src_bbox := dst_srs.TransformRectTo(src_srs, dst_bbox, 16)

	transformer := &ImageTransformer{SrcSRS: src_srs, DstSRS: dst_srs}

	img_opts = *PNG_FORMAT
	img_opts.Resampling = "nearest"

	result := transformer.Transform(
		src_img,
		src_bbox,
		dst_size,
		dst_bbox,
		&img_opts)

	imaging.Save(result.GetImage(), "./transform.png")
	errs := []float64{0.2, 0.5, 1, 2, 4, 6, 8, 12, 16}
	for _, err := range errs {
		transformer := &ImageTransformer{SrcSRS: src_srs, DstSRS: dst_srs, MaxPxErr: err}
		result = transformer.Transform(
			src_img,
			src_bbox,
			dst_size,
			dst_bbox,
			&img_opts)

		imaging.Save(result.GetImage(), fmt.Sprintf("./transform_%d.png", int(err*10)))
	}
}
