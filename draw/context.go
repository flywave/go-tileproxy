package draw

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"sync"

	"github.com/flywave/gg"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type Context struct {
	width  int
	height int

	hasZoom bool
	zoom    int

	hasCenter bool
	center    s2.LatLng

	hasBoundingBox bool
	boundingBox    s2.Rect

	background color.Color

	objects  []MapObject
	overlays []*TileProvider

	tileProvider *TileProvider

	overrideAttribution *string
}

func NewContext() *Context {
	t := new(Context)
	t.width = 512
	t.height = 512
	t.hasZoom = false
	t.hasCenter = false
	t.hasBoundingBox = false
	t.background = nil
	t.userAgent = ""
	t.online = true
	t.tileProvider = NewTileProviderOpenStreetMaps()
	return t
}

func (m *Context) SetTileProvider(t *TileProvider) {
	m.tileProvider = t
}

func (m *Context) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Context) SetZoom(zoom int) {
	m.zoom = zoom
	m.hasZoom = true
}

func (m *Context) SetCenter(center s2.LatLng) {
	m.center = center
	m.hasCenter = true
}

func (m *Context) SetBoundingBox(bbox s2.Rect) {
	m.boundingBox = bbox
	m.hasBoundingBox = true
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

func (m *Context) AddOverlay(overlay *TileProvider) {
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
	return m.tileProvider.Attribution
}

func (m *Context) determineBounds() s2.Rect {
	r := s2.EmptyRect()
	for _, object := range m.objects {
		r = r.Union(object.Bounds())
	}
	return r
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

func (m *Context) determineZoom(bounds s2.Rect, center s2.LatLng) int {
	b := bounds.AddPoint(center)
	if b.IsEmpty() || b.IsPoint() {
		return 15
	}

	tileSize := m.tileProvider.TileSize
	marginL, marginT, marginR, marginB := m.determineExtraMarginPixels()
	w := (float64(m.width) - marginL - marginR) / float64(tileSize)
	h := (float64(m.height) - marginT - marginB) / float64(tileSize)
	if w <= 0 || h <= 0 {
		log.Printf("Object margins are bigger than the target image size => ignoring object margins for calculation of the zoom level")
		w = float64(m.width) / float64(tileSize)
		h = float64(m.height) / float64(tileSize)
	}
	minX := (b.Lo().Lng.Degrees() + 180.0) / 360.0
	maxX := (b.Hi().Lng.Degrees() + 180.0) / 360.0
	minY := (1.0 - math.Log(math.Tan(b.Lo().Lat.Radians())+(1.0/math.Cos(b.Lo().Lat.Radians())))/math.Pi) / 2.0
	maxY := (1.0 - math.Log(math.Tan(b.Hi().Lat.Radians())+(1.0/math.Cos(b.Hi().Lat.Radians())))/math.Pi) / 2.0

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

func (m *Context) determineCenter(bounds s2.Rect) s2.LatLng {
	latLo := bounds.Lo().Lat.Radians()
	latHi := bounds.Hi().Lat.Radians()
	yLo := math.Log((1+math.Sin(latLo))/(1-math.Sin(latLo))) / 2
	yHi := math.Log((1+math.Sin(latHi))/(1-math.Sin(latHi))) / 2
	lat := s1.Angle(math.Atan(math.Sinh((yLo + yHi) / 2)))
	lng := bounds.Center().Lng
	return s2.LatLng{Lat: lat, Lng: lng}
}

func (m *Context) adjustCenter(center s2.LatLng, zoom int) s2.LatLng {
	if m.objects == nil || len(m.objects) == 0 {
		return center
	}

	transformer := newTransformer(m.width, m.height, zoom, center, m.tileProvider.TileSize)

	first := true
	minX := 0.0
	maxX := 0.0
	minY := 0.0
	maxY := 0.0
	for _, object := range m.objects {
		bounds := object.Bounds()
		nwX, nwY := transformer.LatLngToXY(bounds.Vertex(3))
		seX, seY := transformer.LatLngToXY(bounds.Vertex(1))
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
		log.Printf("Object margins are bigger than the target image size => ignoring object margins for adjusting the center")
		return center
	}

	centerX := (maxX + minX) * 0.5
	centerY := (maxY + minY) * 0.5

	return transformer.XYToLatLng(centerX, centerY)
}

func (m *Context) determineZoomCenter() (int, s2.LatLng, error) {
	if m.hasBoundingBox && !m.boundingBox.IsEmpty() {
		center := m.determineCenter(m.boundingBox)
		return m.determineZoom(m.boundingBox, center), center, nil
	}

	if m.hasCenter {
		if m.hasZoom {
			return m.zoom, m.center, nil
		}
		return m.determineZoom(m.determineBounds(), m.center), m.center, nil
	}

	bounds := m.determineBounds()
	if !bounds.IsEmpty() {
		center := m.determineCenter(bounds)
		zoom := m.zoom
		if !m.hasZoom {
			zoom = m.determineZoom(bounds, center)
		}
		return zoom, m.adjustCenter(center, zoom), nil
	}

	return 0, s2.LatLngFromDegrees(0, 0), errors.New("cannot determine map extent: no center coordinates given, no bounding box given, no content (markers, paths, areas) given")
}

type Transformer struct {
	zoom               int
	numTiles           float64
	tileSize           int
	pWidth, pHeight    int
	pCenterX, pCenterY int
	tCountX, tCountY   int
	tCenterX, tCenterY float64
	tOriginX, tOriginY int
	pMinX, pMaxX       int
	proj               s2.Projection
}

func (m *Context) Transformer() (*Transformer, error) {
	zoom, center, err := m.determineZoomCenter()
	if err != nil {
		return nil, err
	}

	return newTransformer(m.width, m.height, zoom, center, m.tileProvider.TileSize), nil
}

func newTransformer(width int, height int, zoom int, llCenter s2.LatLng, tileSize int) *Transformer {
	t := new(Transformer)

	t.zoom = zoom
	t.numTiles = math.Exp2(float64(t.zoom))
	t.tileSize = tileSize

	t.proj = s2.NewMercatorProjection(0.5)

	t.tCenterX, t.tCenterY = t.ll2t(llCenter)

	ww := float64(width) / float64(tileSize)
	hh := float64(height) / float64(tileSize)

	t.tOriginX = int(math.Floor(t.tCenterX - 0.5*ww))
	t.tOriginY = int(math.Floor(t.tCenterY - 0.5*hh))

	t.tCountX = 1 + int(math.Floor(t.tCenterX+0.5*ww)) - t.tOriginX
	t.tCountY = 1 + int(math.Floor(t.tCenterY+0.5*hh)) - t.tOriginY

	t.pWidth = t.tCountX * tileSize
	t.pHeight = t.tCountY * tileSize

	t.pCenterX = int((t.tCenterX - float64(t.tOriginX)) * float64(tileSize))
	t.pCenterY = int((t.tCenterY - float64(t.tOriginY)) * float64(tileSize))

	t.pMinX = t.pCenterX - width/2
	t.pMaxX = t.pMinX + width

	return t
}

func (t *Transformer) ll2t(ll s2.LatLng) (float64, float64) {
	p := t.proj.FromLatLng(ll)
	return t.numTiles * (p.X + 0.5), t.numTiles * (1 - (p.Y + 0.5))
}

func (t *Transformer) LatLngToXY(ll s2.LatLng) (float64, float64) {
	x, y := t.ll2t(ll)
	x = float64(t.pCenterX) + (x-t.tCenterX)*float64(t.tileSize)
	y = float64(t.pCenterY) + (y-t.tCenterY)*float64(t.tileSize)

	offset := t.numTiles * float64(t.tileSize)
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

func (t *Transformer) XYToLatLng(x float64, y float64) s2.LatLng {
	xx := ((((x - float64(t.pCenterX)) / float64(t.tileSize)) + t.tCenterX) / t.numTiles) - 0.5
	yy := 0.5 - (((y-float64(t.pCenterY))/float64(t.tileSize))+t.tCenterY)/t.numTiles
	return t.proj.ToLatLng(r2.Point{X: xx, Y: yy})
}

func (t *Transformer) Rect() (bbox s2.Rect) {
	invNumTiles := 1.0 / t.numTiles

	n := math.Pi - 2.0*math.Pi*float64(t.tOriginY)*invNumTiles
	bbox.Lat.Hi = math.Atan(0.5 * (math.Exp(n) - math.Exp(-n)))
	n = math.Pi - 2.0*math.Pi*float64(t.tOriginY+t.tCountY)*invNumTiles
	bbox.Lat.Lo = math.Atan(0.5 * (math.Exp(n) - math.Exp(-n)))

	bbox.Lng.Lo = float64(t.tOriginX)*invNumTiles*2.0*math.Pi - math.Pi
	bbox.Lng.Hi = float64(t.tOriginX+t.tCountX)*invNumTiles*2.0*math.Pi - math.Pi
	return bbox
}

func (m *Context) Render() (image.Image, error) {
	zoom, center, err := m.determineZoomCenter()
	if err != nil {
		return nil, err
	}

	tileSize := m.tileProvider.TileSize
	trans := newTransformer(m.width, m.height, zoom, center, tileSize)
	img := image.NewRGBA(image.Rect(0, 0, trans.pWidth, trans.pHeight))
	gc := gg.NewContextForRGBA(img)
	if m.background != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{m.background}, image.Point{}, draw.Src)
	}

	layers := []*TileProvider{m.tileProvider}
	if m.overlays != nil {
		layers = append(layers, m.overlays...)
	}

	for _, layer := range layers {
		if err := m.renderLayer(gc, zoom, trans, tileSize, layer); err != nil {
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
	zoom, center, err := m.determineZoomCenter()
	if err != nil {
		return nil, nil, err
	}

	tileSize := m.tileProvider.TileSize
	trans := newTransformer(m.width, m.height, zoom, center, tileSize)
	img := image.NewRGBA(image.Rect(0, 0, trans.pWidth, trans.pHeight))
	gc := gg.NewContextForRGBA(img)
	if m.background != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{m.background}, image.Point{}, draw.Src)
	}

	layers := []*TileProvider{m.tileProvider}
	if m.overlays != nil {
		layers = append(layers, m.overlays...)
	}

	for _, layer := range layers {
		if err := m.renderLayer(gc, zoom, trans, tileSize, layer); err != nil {
			return nil, nil, err
		}
	}

	for _, object := range m.objects {
		object.Draw(gc, trans)
	}

	if m.tileProvider.Attribution == "" {
		return img, trans, nil
	}
	_, textHeight := gc.MeasureString(m.tileProvider.Attribution)
	boxHeight := textHeight + 4.0
	gc.SetRGBA(0.0, 0.0, 0.0, 0.5)
	gc.DrawRectangle(0.0, float64(trans.pHeight)-boxHeight, float64(trans.pWidth), boxHeight)
	gc.Fill()
	gc.SetRGBA(1.0, 1.0, 1.0, 0.75)
	gc.DrawString(m.tileProvider.Attribution, 4.0, float64(m.height)-4.0)

	return img, trans, nil
}

func (m *Context) RenderWithBounds() (image.Image, s2.Rect, error) {
	img, trans, err := m.RenderWithTransformer()
	if err != nil {
		return nil, s2.Rect{}, err

	}
	return img, trans.Rect(), nil
}

func (m *Context) renderLayer(gc *gg.Context, zoom int, trans *Transformer, tileSize int, provider *TileProvider) error {
	var wg sync.WaitGroup
	tiles := (1 << uint(zoom))
	fetchedTiles := make(chan *Tile)
	t := NewTileFetcher(provider)

	go func() {
		for xx := 0; xx < trans.tCountX; xx++ {
			x := trans.tOriginX + xx
			if x < 0 {
				x = x + tiles
			} else if x >= tiles {
				x = x - tiles
			}
			if x < 0 || x >= tiles {
				log.Printf("Skipping out of bounds tile column %d/?", x)
				continue
			}
			for yy := 0; yy < trans.tCountY; yy++ {
				y := trans.tOriginY + yy
				if y < 0 || y >= tiles {
					log.Printf("Skipping out of bounds tile %d/%d", x, y)
					continue
				}
				wg.Add(1)
				tile := &Tile{Zoom: zoom, X: x, Y: y}
				go func(wg *sync.WaitGroup, tile *Tile, xx, yy int) {
					defer wg.Done()
					if err := t.Fetch(tile); err == nil {
						tile.X = xx * tileSize
						tile.Y = yy * tileSize
						fetchedTiles <- tile
					} else if err == errTileNotFound && provider.IgnoreNotFound {
						log.Printf("Error downloading tile file: %s (Ignored)", err)
					} else {
						log.Printf("Error downloading tile file: %s", err)
					}
				}(&wg, tile, xx, yy)
			}
		}
		wg.Wait()
		close(fetchedTiles)
	}()

	for tile := range fetchedTiles {
		gc.DrawImage(tile.Img, tile.X, tile.Y)
	}

	return nil
}
