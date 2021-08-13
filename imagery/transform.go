package imagery

import (
	"image"
	"image/color"
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"

	"github.com/flywave/imaging"
)

type ImageTransformer struct {
	SrcSRS   geo.Proj
	DstSRS   geo.Proj
	DstBBox  *vec2d.Rect
	DstSize  *vec2d.Rect
	MaxPxErr float64
}

func (t *ImageTransformer) Transform(srcImg tile.Source, srcBBox vec2d.Rect, dstSize [2]uint32, dstBBox vec2d.Rect, imageOpts *ImageOptions) tile.Source {
	img := srcImg.GetTile().(image.Image)
	if t.noTransformationNeeded([2]uint32{uint32(img.Bounds().Dx()), uint32(img.Bounds().Dy())}, srcBBox, dstSize, dstBBox) {
		return srcImg
	}

	var result tile.Source
	if t.SrcSRS.Eq(t.DstSRS) {
		result = t.transformSimple(srcImg, srcBBox, dstSize, dstBBox,
			imageOpts)
	} else {
		result = t.transform(srcImg, srcBBox, dstSize, dstBBox, imageOpts)
	}
	result.SetCacheable(srcImg.GetCacheable())
	return result
}

func (t *ImageTransformer) noTransformationNeeded(srcSize [2]uint32, srcBBox vec2d.Rect, dstSize [2]uint32, dstBBox vec2d.Rect) bool {
	xres := (dstBBox.Max[0] - dstBBox.Min[0]) / float64(dstSize[0])
	yres := (dstBBox.Min[1] - dstBBox.Min[1]) / float64(dstSize[1])
	return (srcSize == dstSize &&
		t.SrcSRS.Eq(t.DstSRS) &&
		geo.BBoxEquals(srcBBox, dstBBox, xres/10, yres/10))
}

func (t *ImageTransformer) transformSimple(srcImg tile.Source, srcBBox vec2d.Rect, dstSize [2]uint32, dstBBox vec2d.Rect, imageOpts *ImageOptions) tile.Source {
	srcQuad := vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{float64(srcImg.GetSize()[0]), float64(srcImg.GetSize()[1])}}
	to_src_px := geo.MakeLinTransf(srcBBox, srcQuad)

	minxy := to_src_px([]float64{dstBBox.Min[0], dstBBox.Max[1]})
	maxxy := to_src_px([]float64{dstBBox.Max[0], dstBBox.Min[1]})

	src_res := []float64{(srcBBox.Min[0] - srcBBox.Max[0]) / float64(srcImg.GetSize()[0]),
		(srcBBox.Min[1] - srcBBox.Max[1]) / float64(srcImg.GetSize()[1])}
	dst_res := []float64{(dstBBox.Min[0] - dstBBox.Max[0]) / float64(dstSize[0]),
		(dstBBox.Min[1] - dstBBox.Max[1]) / float64(dstSize[1])}

	tenth_px_res := []float64{math.Abs(float64(dst_res[0]) / (float64(dstSize[0]) * 10)),
		math.Abs(float64(dst_res[1]) / (float64(dstSize[1]) * 10))}
	var result image.Image
	if math.Abs(src_res[0]-dst_res[0]) < tenth_px_res[0] &&
		math.Abs(src_res[1]-dst_res[1]) < tenth_px_res[1] {
		minx := int(math.Round(minxy[0]))
		miny := int(math.Round(minxy[1]))
		result = imaging.Crop(srcImg.GetTile().(image.Image), image.Rect(minx, miny,
			minx+int(dstSize[0]), miny+int(dstSize[1])))
	} else {
		result = imaging.Resize(srcImg.GetTile().(image.Image), int(maxxy[0]-minxy[0]), int(maxxy[1]-minxy[1]), image_filter[imageOpts.Resampling])
	}

	return &ImageSource{image: result, size: dstSize[:], Options: imageOpts}
}

func (t *ImageTransformer) transform(srcImg tile.Source, srcBBox vec2d.Rect, dstSize [2]uint32, dstBBox vec2d.Rect, imageOpts *ImageOptions) tile.Source {

	meshes := transformMeshes(
		srcImg.GetSize(),
		srcBBox,
		t.SrcSRS,
		dstSize,
		dstBBox,
		t.DstSRS,
		int32(t.MaxPxErr),
		false,
	)

	img := srcImg.GetTile().(image.Image)
	result := imaging.Transform(img, int(dstSize[0]), int(dstSize[1]), imaging.MESH, meshes,
		image_filter[imageOpts.Resampling], true, color.Black)

	return &ImageSource{image: result, size: dstSize[:], Options: imageOpts}

}

func NewImageTransformer(srcSrs geo.Proj, dstSrs geo.Proj, max_px_err *float64) *ImageTransformer {
	maxpxerr := 1.0
	if max_px_err != nil {
		maxpxerr = *max_px_err
	}
	return &ImageTransformer{SrcSRS: srcSrs, DstSRS: dstSrs, DstBBox: nil, DstSize: nil, MaxPxErr: maxpxerr}
}

func dstQuadToSrc(quad []float64, toDstW func([]float64) []float64, to_src_px func([]float64) []float64, srcSrs geo.Proj, dstSrs geo.Proj, px_offset float64) ([]float64, []float64) {
	srcQuad := make([]float64, 0)
	dest_pxs := [][]float64{{quad[0], quad[1]}, {quad[0], quad[3]},
		{quad[2], quad[3]}, {quad[2], quad[1]}}
	for _, dst_px := range dest_pxs {
		dst_w := toDstW([]float64{dst_px[0] + px_offset, dst_px[1] + px_offset})
		src_w := dstSrs.TransformTo(srcSrs, []vec2d.T{{dst_w[0], dst_w[1]}})
		src_px := to_src_px(src_w[0][:])
		srcQuad = append(srcQuad, src_px...)
	}
	return quad, srcQuad
}

func isGood(quad, srcQuad []float64, toDstW func([]float64) []float64, toSrcW func([]float64) []float64, srcSrs geo.Proj, dstSrs geo.Proj, maxErr float64) bool {
	w := quad[2] - quad[0]
	h := quad[3] - quad[1]

	if w < 50 || h < 50 {
		return true
	}

	xc := quad[0] + w/2.0 - 0.5
	yc := quad[1] + h/2.0 - 0.5

	dst_w := toDstW([]float64{xc, yc})

	src_px := centerQuadTransform(quad, srcQuad)
	to_p := toSrcW(src_px)
	real_dst_w := srcSrs.TransformTo(dstSrs, []vec2d.T{{to_p[0], to_p[1]}})

	err := math.Max(math.Abs(dst_w[0]-real_dst_w[0][0]), math.Abs(dst_w[1]-real_dst_w[0][1]))
	return err < maxErr
}

func addMeshes(quads [][]float64, toDstW func([]float64) []float64, to_src_px func([]float64) []float64, toSrcW func([]float64) []float64, srcSrs geo.Proj, dstSrs geo.Proj, px_offset float64, maxErr float64, meshes map[[4]float64][]float64) {
	for _, quad := range quads {
		quad, srcQuad := dstQuadToSrc(quad, toDstW, to_src_px, srcSrs, dstSrs, px_offset)
		key := [4]float64{quad[0], quad[1], quad[2], quad[3]}
		if isGood(quad, srcQuad, toDstW, toSrcW, srcSrs, dstSrs, maxErr) {
			meshes[key] = srcQuad
		} else {
			addMeshes(divideQuad(quad), toDstW, to_src_px, toSrcW, srcSrs, dstSrs, px_offset, maxErr, meshes)
		}
	}
}

func transformMeshes(srcSize [2]uint32, srcBBox vec2d.Rect, srcSrs geo.Proj, dstSize [2]uint32, dstBBox vec2d.Rect, dstSrs geo.Proj, maxPixelErr int32, useCenterPixel bool) interface{} {
	srcBBox = srcSrs.AlignBBox(srcBBox)
	dstBBox = dstSrs.AlignBBox(dstBBox)

	src_rect := vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{float64(srcSize[0]), float64(srcSize[1])}}
	dst_rect := vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{float64(dstSize[0]), float64(dstSize[1])}}

	to_src_px := geo.MakeLinTransf(srcBBox, src_rect)

	toSrcW := geo.MakeLinTransf(src_rect, srcBBox)
	toDstW := geo.MakeLinTransf(dst_rect, dstBBox)

	var px_offset float64

	if useCenterPixel {
		px_offset = 0.5
	} else {
		px_offset = 0.0
	}

	res := (dstBBox.Max[0] - dstBBox.Min[0]) / float64(dstSize[0])
	maxErr := float64(maxPixelErr) * res

	meshes := make(map[[4]float64][]float64)
	root := [][]float64{{0, 0, float64(dstSize[0]), float64(dstSize[1])}}

	addMeshes(root, toDstW, to_src_px, toSrcW, srcSrs, dstSrs, px_offset, maxErr, meshes)

	return meshes
}

func centerQuadTransform(quad []float64, srcQuad []float64) []float64 {
	w := quad[2] - quad[0]
	h := quad[3] - quad[1]

	nw := srcQuad[0:2]
	sw := srcQuad[2:4]
	se := srcQuad[4:6]
	ne := srcQuad[6:8]

	x0, y0 := nw[0], nw[1]

	As := float64(1.0 / w)
	At := float64(1.0 / h)

	a0 := x0
	a1 := (ne[0] - x0) * As
	a2 := (sw[0] - x0) * At
	a3 := (se[0] - sw[0] - ne[0] + x0) * As * At
	a4 := y0
	a5 := (ne[1] - y0) * As
	a6 := (sw[1] - y0) * At
	a7 := (se[1] - sw[1] - ne[1] + y0) * As * At

	x := float64(w)/2.0 - 0.5
	y := float64(h)/2.0 - 0.5

	return []float64{
		float64(a0) + float64(a1)*x + float64(a2)*y + float64(a3)*x*y,
		float64(a4) + float64(a5)*x + float64(a6)*y + float64(a7)*x*y,
	}
}

func divideQuad(quad []float64) [][]float64 {
	w := quad[2] - quad[0]
	h := quad[3] - quad[1]

	xc := float64(quad[0] + w/2)
	yc := float64(quad[1] + h/2)

	if w > 2*h {
		return [][]float64{
			{quad[0], quad[1], xc, quad[3]},
			{xc, quad[1], quad[2], quad[3]},
		}
	}
	if h > 2*w {
		return [][]float64{
			{quad[0], quad[1], quad[2], yc},
			{quad[0], yc, quad[2], quad[3]},
		}
	}

	return [][]float64{
		{quad[0], quad[1], xc, yc},
		{xc, quad[1], quad[2], yc},
		{quad[0], yc, xc, quad[3]},
		{xc, yc, quad[2], quad[3]},
	}
}
