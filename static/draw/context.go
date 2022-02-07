package draw

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"sync"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/static"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
)

type Context struct {
	width               int
	height              int
	zoom                *int
	boundingBox         *vec2d.Rect
	boundingBoxSrs      geo.Proj
	background          color.Color
	objects             []MapObject
	overlays            []static.TileProvider
	tileProvider        static.TileProvider
	overrideAttribution *string
	grid                *geo.TileGrid
}

func NewContext() *Context {
	t := new(Context)
	t.width = 512
	t.height = 512
	t.background = nil
	t.tileProvider = nil
	return t
}

func (m *Context) SetTileProvider(t static.TileProvider) {
	m.tileProvider = t
}

func (m *Context) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Context) SetZoom(zoom int) {
	m.zoom = &zoom
}

func (m *Context) SetBoundingBox(bbox vec2d.Rect, srs geo.Proj) {
	m.boundingBox = &bbox
	m.boundingBoxSrs = srs
}

func (m *Context) SetBackground(col color.Color) {
	m.background = col
}

func (m *Context) AddMarker(marker *Marker) {
	m.AddObject(marker)
}

func (m *Context) ClearMarkers() {
	filtered := []MapObject{}
	for _, object := range m.objects {
		switch object.(type) {
		case *Marker:
			// skip
		default:
			filtered = append(filtered, object)
		}
	}
	m.objects = filtered
}

func (m *Context) AddPath(path *Path) {
	m.AddObject(path)
}

func (m *Context) ClearPaths() {
	filtered := []MapObject{}
	for _, object := range m.objects {
		switch object.(type) {
		case *Path:
			// skip
		default:
			filtered = append(filtered, object)
		}
	}
	m.objects = filtered
}

func (m *Context) AddArea(area *Area) {
	m.AddObject(area)
}

func (m *Context) ClearAreas() {
	filtered := []MapObject{}
	for _, object := range m.objects {
		switch object.(type) {
		case *Area:
			// skip
		default:
			filtered = append(filtered, object)
		}
	}
	m.objects = filtered
}

func (m *Context) AddCircle(circle *Circle) {
	m.AddObject(circle)
}

func (m *Context) ClearCircles() {
	filtered := []MapObject{}
	for _, object := range m.objects {
		switch object.(type) {
		case *Circle:
			// skip
		default:
			filtered = append(filtered, object)
		}
	}
	m.objects = filtered
}

func (m *Context) AddObject(object MapObject) {
	m.objects = append(m.objects, object)
}

func (m *Context) ClearObjects() {
	m.objects = nil
}

func (m *Context) AddOverlay(overlay static.TileProvider) {
	m.overlays = append(m.overlays, overlay)
}

func (m *Context) ClearOverlays() {
	m.overlays = nil
}

func (m *Context) OverrideAttribution(attribution string) {
	m.overrideAttribution = &attribution
}

func (m *Context) Attribution() string {
	if m.overrideAttribution != nil {
		return *m.overrideAttribution
	}
	return m.tileProvider.Attribution()
}

func (m *Context) determineBounds() (vec2d.Rect, geo.Proj) {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	var srs geo.Proj
	for _, object := range m.objects {
		bb := object.Bounds()
		if srs == nil {
			srs = object.SrsProj()
		} else if !srs.Eq(object.SrsProj()) {
			bb = object.SrsProj().TransformRectTo(srs, bb, 16)
		}

		r.Join(&bb)
	}
	return r, srs
}

func (m *Context) determineExtraMarginPixels() (float64, float64, float64, float64) {
	maxL := 0.0
	maxT := 0.0
	maxR := 0.0
	maxB := 0.0
	if m.Attribution() != "" {
		maxB = 12.0
	}
	for _, object := range m.objects {
		l, t, r, b := object.ExtraMarginPixels()
		maxL = math.Max(maxL, l)
		maxT = math.Max(maxT, t)
		maxR = math.Max(maxR, r)
		maxB = math.Max(maxB, b)
	}
	return maxL, maxT, maxR, maxB
}

func (m *Context) determineZoom(bounds vec2d.Rect, center vec2d.T) int {
	bounds.Extend(&center)
	if bounds.Area() == 0 {
		return 15
	}

	b := bounds

	marginL, marginT, marginR, marginB := m.determineExtraMarginPixels()
	w := (float64(m.width) - marginL - marginR) / float64(m.grid.TileSize[0])
	h := (float64(m.height) - marginT - marginB) / float64(m.grid.TileSize[1])

	if w <= 0 || h <= 0 {
		w = float64(m.width) / float64(m.grid.TileSize[0])
		h = float64(m.height) / float64(m.grid.TileSize[1])
	}

	minX := (b.Min[1] + 180.0) / 360.0
	maxX := (b.Max[1] + 180.0) / 360.0
	minY := (1.0 - math.Log(math.Tan(DegreesToRadians(b.Min[0]))+(1.0/math.Cos(b.Min[1])))/math.Pi) / 2.0
	maxY := (1.0 - math.Log(math.Tan(DegreesToRadians(b.Max[0]))+(1.0/math.Cos(b.Max[1])))/math.Pi) / 2.0

	dx := maxX - minX
	for dx < 0 {
		dx = dx + 1
	}
	for dx > 1 {
		dx = dx - 1
	}
	dy := math.Abs(maxY - minY)

	zoom := 1
	for zoom < 30 {
		tiles := float64(uint(1) << uint(zoom))
		if dx*tiles > w || dy*tiles > h {
			return zoom - 1
		}
		zoom = zoom + 1
	}

	return 15
}

func (m *Context) determineCenter(bounds vec2d.Rect) vec2d.T {
	latLo := DegreesToRadians(bounds.Min[0])
	latHi := DegreesToRadians(bounds.Max[0])
	yLo := math.Log((1+math.Sin(latLo))/(1-math.Sin(latLo))) / 2
	yHi := math.Log((1+math.Sin(latHi))/(1-math.Sin(latHi))) / 2
	lat := RadiansToDegrees(math.Atan(math.Sinh((yLo + yHi) / 2)))
	lng := (bounds.Min[1] + bounds.Max[1]) / 2
	return vec2d.T{lat, lng}
}

func (m *Context) adjustCenter(center vec2d.T, srs geo.Proj, zoom int) vec2d.T {
	if m.objects == nil || len(m.objects) == 0 {
		return center
	}

	transformer := newTransformer(m.width, m.height, zoom, center, srs, m.grid)

	first := true
	minX := 0.0
	maxX := 0.0
	minY := 0.0
	maxY := 0.0
	for _, object := range m.objects {
		bounds := object.Bounds()
		nwX, nwY := transformer.LatLngToXY(bounds.Min, srs)
		seX, seY := transformer.LatLngToXY(bounds.Max, srs)
		l, t, r, b := object.ExtraMarginPixels()
		if first {
			minX = nwX - l
			maxX = seX + r
			minY = nwY - t
			maxY = seY + b
			first = false
		} else {
			minX = math.Min(minX, nwX-l)
			maxX = math.Max(maxX, seX+r)
			minY = math.Min(minY, nwY-t)
			maxY = math.Max(maxY, seY+b)
		}
	}

	if (maxX-minX) > float64(m.width) || (maxY-minY) > float64(m.height) {
		return center
	}

	centerX := (maxX + minX) * 0.5
	centerY := (maxY + minY) * 0.5

	return transformer.XYToLatLng(centerX, centerY, srs)
}

func (m *Context) determineZoomCenter() (int, vec2d.T, geo.Proj, error) {
	if m.boundingBox != nil {
		center := m.determineCenter(*m.boundingBox)
		return m.determineZoom(*m.boundingBox, center), center, m.boundingBoxSrs, nil
	}

	bounds, srs := m.determineBounds()
	if m.boundingBox == nil {
		center := m.determineCenter(bounds)
		zoom := m.zoom
		if zoom == nil {
			z := m.determineZoom(bounds, center)
			zoom = &z
		}
		return *zoom, m.adjustCenter(center, srs, *zoom), srs, nil
	}

	return 0, vec2d.T{}, nil, errors.New("cannot determine map extent: no center coordinates given, no bounding box given, no content (markers, paths, areas) given")
}

type Transformer struct {
	zoom               int
	numTiles           float64
	pWidth, pHeight    int
	pCenterX, pCenterY int
	tCountX, tCountY   int
	tCenterX, tCenterY float64
	tOriginX, tOriginY int
	pMinX, pMaxX       int
	grid               *geo.TileGrid
}

func (m *Context) Transformer() (*Transformer, error) {
	zoom, center, srs, err := m.determineZoomCenter()
	if err != nil {
		return nil, err
	}

	return newTransformer(m.width, m.height, zoom, center, srs, m.grid), nil
}

func newTransformer(width int, height int, zoom int, llCenter vec2d.T, srs geo.Proj, grid *geo.TileGrid) *Transformer {
	t := new(Transformer)

	t.zoom = zoom
	t.numTiles = math.Exp2(float64(t.zoom))
	t.grid = grid

	t.tCenterX, t.tCenterY = t.ll2t(llCenter, srs)

	ww := float64(width) / float64(grid.TileSize[0])
	hh := float64(height) / float64(grid.TileSize[1])

	t.tOriginX = int(math.Floor(t.tCenterX - 0.5*ww))
	t.tOriginY = int(math.Floor(t.tCenterY - 0.5*hh))

	t.tCountX = 1 + int(math.Floor(t.tCenterX+0.5*ww)) - t.tOriginX
	t.tCountY = 1 + int(math.Floor(t.tCenterY+0.5*hh)) - t.tOriginY

	t.pWidth = t.tCountX * int(grid.TileSize[0])
	t.pHeight = t.tCountY * int(grid.TileSize[1])

	t.pCenterX = int((t.tCenterX - float64(t.tOriginX)) * float64(grid.TileSize[0]))
	t.pCenterY = int((t.tCenterY - float64(t.tOriginY)) * float64(grid.TileSize[1]))

	t.pMinX = t.pCenterX - width/2
	t.pMaxX = t.pMinX + width

	return t
}

func (t *Transformer) ll2t(ll vec2d.T, srs geo.Proj) (float64, float64) {
	p := t.grid.Srs.TransformTo(srs, []vec2d.T{ll})
	return t.numTiles * (p[0][0] + 0.5), t.numTiles * (1 - (p[0][1] + 0.5))
}

func (t *Transformer) LatLngToXY(ll vec2d.T, srs geo.Proj) (float64, float64) {
	x, y := t.ll2t(ll, srs)

	x = float64(t.pCenterX) + (x-t.tCenterX)*float64(t.grid.TileSize[0])
	y = float64(t.pCenterY) + (y-t.tCenterY)*float64(t.grid.TileSize[1])

	offset := t.numTiles * float64(t.grid.TileSize[0])
	if x < float64(t.pMinX) {
		for x < float64(t.pMinX) {
			x = x + offset
		}
	} else if x >= float64(t.pMaxX) {
		for x >= float64(t.pMaxX) {
			x = x - offset
		}
	}
	return x, y
}

func (t *Transformer) XYToLatLng(x float64, y float64, srs geo.Proj) vec2d.T {
	xx := ((((x - float64(t.pCenterX)) / float64(t.grid.TileSize[0])) + t.tCenterX) / t.numTiles) - 0.5
	yy := 0.5 - (((y-float64(t.pCenterY))/float64(t.grid.TileSize[1]))+t.tCenterY)/t.numTiles
	return t.grid.Srs.TransformTo(srs, []vec2d.T{{xx, yy}})[0]
}

func (t *Transformer) Rect() (bbox vec2d.Rect) {
	invNumTiles := 1.0 / t.numTiles

	n := math.Pi - 2.0*math.Pi*float64(t.tOriginY)*invNumTiles
	bbox.Max[0] = RadiansToDegrees(math.Atan(0.5 * (math.Exp(n) - math.Exp(-n))))
	n = math.Pi - 2.0*math.Pi*float64(t.tOriginY+t.tCountY)*invNumTiles
	bbox.Min[0] = RadiansToDegrees(math.Atan(0.5 * (math.Exp(n) - math.Exp(-n))))

	bbox.Min[1] = RadiansToDegrees(float64(t.tOriginX)*invNumTiles*2.0*math.Pi - math.Pi)
	bbox.Max[1] = RadiansToDegrees(float64(t.tOriginX+t.tCountX)*invNumTiles*2.0*math.Pi - math.Pi)

	return bbox
}

func (m *Context) Render() (image.Image, error) {
	zoom, center, srs, err := m.determineZoomCenter()
	if err != nil {
		return nil, err
	}

	trans := newTransformer(m.width, m.height, zoom, center, srs, m.grid)
	img := image.NewRGBA(image.Rect(0, 0, trans.pWidth, trans.pHeight))
	gc := gg.NewContextForRGBA(img)
	if m.background != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{m.background}, image.Point{}, draw.Src)
	}

	layers := []static.TileProvider{m.tileProvider}
	if m.overlays != nil {
		layers = append(layers, m.overlays...)
	}

	for _, layer := range layers {
		if err := m.renderLayer(gc, zoom, trans, layer); err != nil {
			return nil, err
		}
	}

	for _, object := range m.objects {
		object.Draw(gc, trans)
	}

	croppedImg := image.NewRGBA(image.Rect(0, 0, int(m.width), int(m.height)))
	draw.Draw(croppedImg, image.Rect(0, 0, int(m.width), int(m.height)),
		img, image.Point{trans.pCenterX - int(m.width)/2, trans.pCenterY - int(m.height)/2},
		draw.Src)

	attribution := m.Attribution()
	if attribution == "" {
		return croppedImg, nil
	}
	_, textHeight := gc.MeasureString(attribution)
	boxHeight := textHeight + 4.0
	gc = gg.NewContextForRGBA(croppedImg)
	gc.SetRGBA(0.0, 0.0, 0.0, 0.5)
	gc.DrawRectangle(0.0, float64(m.height)-boxHeight, float64(m.width), boxHeight)
	gc.Fill()
	gc.SetRGBA(1.0, 1.0, 1.0, 0.75)
	gc.DrawString(attribution, 4.0, float64(m.height)-4.0)

	return croppedImg, nil
}

func (m *Context) RenderWithTransformer() (image.Image, *Transformer, error) {
	zoom, center, srs, err := m.determineZoomCenter()
	if err != nil {
		return nil, nil, err
	}

	trans := newTransformer(m.width, m.height, zoom, center, srs, m.grid)
	img := image.NewRGBA(image.Rect(0, 0, trans.pWidth, trans.pHeight))
	gc := gg.NewContextForRGBA(img)
	if m.background != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{m.background}, image.Point{}, draw.Src)
	}

	layers := []static.TileProvider{m.tileProvider}
	if m.overlays != nil {
		layers = append(layers, m.overlays...)
	}

	for _, layer := range layers {
		if err := m.renderLayer(gc, zoom, trans, layer); err != nil {
			return nil, nil, err
		}
	}

	for _, object := range m.objects {
		object.Draw(gc, trans)
	}

	if m.tileProvider.Attribution() == "" {
		return img, trans, nil
	}
	_, textHeight := gc.MeasureString(m.tileProvider.Attribution())
	boxHeight := textHeight + 4.0
	gc.SetRGBA(0.0, 0.0, 0.0, 0.5)
	gc.DrawRectangle(0.0, float64(trans.pHeight)-boxHeight, float64(trans.pWidth), boxHeight)
	gc.Fill()
	gc.SetRGBA(1.0, 1.0, 1.0, 0.75)
	gc.DrawString(m.tileProvider.Attribution(), 4.0, float64(m.height)-4.0)

	return img, trans, nil
}

func (m *Context) RenderWithBounds() (image.Image, vec2d.Rect, error) {
	img, trans, err := m.RenderWithTransformer()
	if err != nil {
		return nil, vec2d.Rect{}, err

	}
	return img, trans.Rect(), nil
}

type tile struct {
	t    *cache.Tile
	offx int
	offy int
}

func (m *Context) renderLayer(gc *gg.Context, zoom int, trans *Transformer, provider static.TileProvider) error {
	var wg sync.WaitGroup
	tiles := (1 << uint(zoom))
	fetchedTiles := make(chan *tile)
	f := static.NewTileFetcher(provider)

	go func() {
		for xx := 0; xx < trans.tCountX; xx++ {
			x := trans.tOriginX + xx
			if x < 0 {
				x = x + tiles
			} else if x >= tiles {
				x = x - tiles
			}
			if x < 0 || x >= tiles {
				continue
			}
			for yy := 0; yy < trans.tCountY; yy++ {
				y := trans.tOriginY + yy
				if y < 0 || y >= tiles {
					continue
				}
				wg.Add(1)
				coord := [3]int{x, y, zoom}
				go func(wg *sync.WaitGroup, c [3]int, xx, yy int) {
					defer wg.Done()
					if ti, err := f.Fetch(c); err == nil {
						t := &tile{t: ti}
						t.offx = xx * int(m.grid.TileSize[0])
						t.offy = yy * int(m.grid.TileSize[1])
						fetchedTiles <- t
					} else if err != nil {
						log.Printf("Error downloading tile file: %s (Ignored)", err)
					}
				}(&wg, coord, xx, yy)
			}
		}
		wg.Wait()
		close(fetchedTiles)
	}()

	for tile := range fetchedTiles {
		image := tile.t.Source.(*imagery.ImageSource).GetImage()
		gc.DrawImage(image, tile.offx, tile.offy)
	}

	return nil
}
