package setting

type ProxyDataset struct {
	Coverages    map[string]Coverage
	Services     map[string]interface{}
	Grids        map[string]interface{}
	LegendLayers map[string]interface{}
	InfoLayers   map[string]interface{}
	MapLayers    map[string]interface{}
	Sources      map[string]interface{}
	Seeds        map[string]Seed
	Cleanups     map[string]Cleanup
}

func NewProxyDataset() *ProxyDataset {
	return &ProxyDataset{Coverages: make(map[string]Coverage), Services: make(map[string]interface{}), MapLayers: make(map[string]interface{}), InfoLayers: make(map[string]interface{}), LegendLayers: make(map[string]interface{}), Grids: make(map[string]interface{}), Sources: make(map[string]interface{})}
}

func (l *ProxyDataset) AddCoverage(name string, cov Coverage) {
	l.Coverages[name] = cov
}

func (l *ProxyDataset) AddGrid(name string, g GridOpts) {
	l.Grids[name] = g
}

func (l *ProxyDataset) AddTMSService(name string, ser TMSService) {
	l.Services[name] = ser
}

func (l *ProxyDataset) AddMapboxService(name string, ser MapboxService) {
	l.Services[name] = ser
}

func (l *ProxyDataset) AddWMTSService(name string, ser WMTSService) {
	l.Services[name] = ser
}

func (l *ProxyDataset) AddWMSService(name string, ser WMSService) {
	l.Services[name] = ser
}

func (l *ProxyDataset) AddTileLayer(name string, layer TileLayer) {
	l.MapLayers[name] = layer
}

func (l *ProxyDataset) AddWMSLayer(name string, layer WMSLayer) {
	l.MapLayers[name] = layer
}

func (l *ProxyDataset) AddMapboxTileLayer(name string, layer MapboxTileLayer) {
	l.MapLayers[name] = layer
}

func (l *ProxyDataset) AddWMSSource(name string, src WMSSource) {
	l.Sources[name] = src
}

func (l *ProxyDataset) AddTileSource(name string, src TileSource) {
	l.Sources[name] = src
}

func (l *ProxyDataset) AddMapboxTileSource(name string, src MapboxTileSource) {
	l.Sources[name] = src
}

func (l *ProxyDataset) AddLuokuangTileSource(name string, src LuokuangTileSource) {
	l.Sources[name] = src
}

func (l *ProxyDataset) AddArcgisSource(name string, src ArcgisSource) {
	l.Sources[name] = src
}

func (l *ProxyDataset) AddSeed(name string, s Seed) {
	l.Sources[name] = s
}

func (l *ProxyDataset) AddCleanup(name string, c Cleanup) {
	l.Sources[name] = c
}
