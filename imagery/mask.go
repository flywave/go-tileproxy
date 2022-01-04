package imagery

import (
	"image"
	"image/color"
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geos"
	"github.com/flywave/go-tileproxy/tile"

	"github.com/flywave/gg"
)

func drawPolygon(dc *gg.Context, pg *geos.Geometry, transf func([]float64) []float64) {
	if pg.IsEmpty() {
		return
	}
	{
		dc.SetColor(color.Opaque)
		coords := pg.GetExteriorRing().GetCoords()
		dc.NewSubPath()
		cp := coords[0]

		p := transf([]float64{cp.X, cp.Y})
		dc.MoveTo(p[0], p[1])

		for _, cp := range coords[1:] {
			p = transf([]float64{cp.X, cp.Y})
			dc.LineTo(p[0], p[1])
		}
		dc.ClosePath()
		dc.Fill()
	}
	nr := pg.GetNumInteriorRings()
	for i := 0; i < nr; i++ {
		coords := pg.GetInteriorRingN(i).GetCoords()
		dc.SetColor(color.Transparent)
		dc.NewSubPath()
		cp := coords[0]
		p := transf([]float64{cp.X, cp.Y})
		dc.MoveTo(p[0], p[1])

		for _, cp := range coords[1:] {
			p = transf([]float64{cp.X, cp.Y})
			dc.LineTo(p[0], p[1])
		}
		dc.ClosePath()
		dc.Fill()
	}
}

func imageMaskFromGeom(size [2]int, bbox vec2d.Rect, polygons []*geos.Geometry) *image.Alpha {
	dc := gg.NewContext(size[0], size[1])
	if len(polygons) == 0 {
		return dc.AsMask()
	}

	transf := geo.MakeLinTransf(bbox, vec2d.Rect{Min: vec2d.T{float64(0), float64(0)}, Max: vec2d.T{float64(size[0]), float64(size[1])}})

	buffer := -0.1 * math.Min((bbox.Max[0]-bbox.Min[0])/float64(size[0]), (bbox.Max[1]-bbox.Min[1])/float64(size[1]))

	for _, p := range polygons {
		buffered := p.BufferWithStyle(buffer, 1, geos.CAP_ROUND, geos.JOIN_MITRE, 5.0)

		if buffered.IsEmpty() {
			continue
		}

		if buffered.GetType() == geos.MULTIPOLYGON {
			ng := p.GetNumGeometries()
			for i := 0; i < ng; i++ {
				drawPolygon(dc, p.GetGeometryN(i), transf)
			}
		} else {
			drawPolygon(dc, buffered, transf)
		}
	}

	return dc.AsMask()
}

func flattenToPolygons(geometry *geos.Geometry) []*geos.Geometry {
	if geometry.GetType() == geos.POLYGON {
		return []*geos.Geometry{geometry}
	}

	if geometry.GetType() == geos.MULTIPOLYGON {
		num := geometry.GetNumGeometries()
		geoms := make([]*geos.Geometry, num)
		for i := 0; i < num; i++ {
			geoms[i] = geometry.GetGeometryN(i)
		}
		return geoms
	}

	if geometry.GetType() == geos.GEOMETRYCOLLECTION {
		geoms := make([]*geos.Geometry, 0)
		num := geometry.GetNumGeometries()
		for i := 0; i < num; i++ {
			g := geometry.GetGeometryN(i)
			if g.GetType() == geos.POLYGON {
				geoms[i] = g
			}
		}
		return geoms
	}

	return nil
}

func maskPolygons(bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) []*geos.Geometry {
	coverage = coverage.TransformTo(bbox_srs)
	coverage = coverage.Intersection(bbox, bbox_srs)
	return flattenToPolygons(coverage.GetGeom())
}

func maskImage(img image.Image, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) (image.Image, *image.Alpha) {
	geom := maskPolygons(bbox, bbox_srs, coverage)
	size := [2]int{img.Bounds().Dx(), img.Bounds().Dy()}
	mask := imageMaskFromGeom(size, bbox, geom)
	return img, mask
}

func MaskImageSourceFromCoverage(img_source tile.Source, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage, image_opts *ImageOptions) tile.Source {
	var opts *ImageOptions
	if image_opts == nil {
		opts = img_source.GetTileOptions().(*ImageOptions)
	} else {
		opts = image_opts
	}
	var mask *image.Alpha
	img := img_source.GetTile().(image.Image)
	img, mask = maskImage(img, bbox, bbox_srs, coverage)

	result := CreateImage([2]uint32{uint32(img.Bounds().Dx()), uint32(img.Bounds().Dy())}, image_opts)

	dc := gg.NewContextForImage(result)

	dc.SetMask(mask)
	dc.DrawImage(img, 0, 0)

	return &ImageSource{image: dc.Image(), Options: opts, cacheable: nil}
}
