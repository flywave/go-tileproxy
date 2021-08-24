package setting

type ProxySetting struct {
	Globals   Globals
	Coverages map[string]Coverage
	Services  map[string]interface{}
	Layers    map[string]interface{}
	Grids     map[string]interface{}
	Caches    map[string]interface{}
	Sources   map[string]interface{}
	Seeds     map[string]Seed
	Cleanups  map[string]Cleanup
}

func NewProxySetting(g Globals) *ProxySetting {
	return &ProxySetting{Globals: g, Coverages: make(map[string]Coverage), Services: make(map[string]interface{}), Layers: make(map[string]interface{}), Grids: make(map[string]interface{}), Caches: make(map[string]interface{}), Sources: make(map[string]interface{})}
}

func (l *ProxySetting) AddCoverage(name string, cov Coverage) {
	l.Coverages[name] = cov
}

func (l *ProxySetting) AddGrid(name string, g GridOpts) {
	l.Grids[name] = g
}

func (l *ProxySetting) AddLocalCache(name string, c LocalCache) {
	l.Caches[name] = c
}

func (l *ProxySetting) AddS3Cache(name string, c S3Cache) {
	l.Caches[name] = c
}

func (l *ProxySetting) AddTMSService(name string, ser TMSService) {
	l.Services[name] = ser
}

func (l *ProxySetting) AddMapboxService(name string, ser MapboxService) {
	l.Services[name] = ser
}

func (l *ProxySetting) AddWMTSService(name string, ser WMTSService) {
	l.Services[name] = ser
}

func (l *ProxySetting) AddWMSService(name string, ser WMSService) {
	l.Services[name] = ser
}

func (l *ProxySetting) AddTileLayer(name string, layer TileLayer) {
	l.Layers[name] = layer
}

func (l *ProxySetting) AddWMSLayer(name string, layer WMSLayer) {
	l.Layers[name] = layer
}

func (l *ProxySetting) AddMapboxTileLayer(name string, layer MapboxTileLayer) {
	l.Layers[name] = layer
}

func (l *ProxySetting) AddWMSSource(name string, src WMSSource) {
	l.Sources[name] = src
}

func (l *ProxySetting) AddTileSource(name string, src TileSource) {
	l.Sources[name] = src
}

func (l *ProxySetting) AddMapboxTileSource(name string, src MapboxTileSource) {
	l.Sources[name] = src
}

func (l *ProxySetting) AddLuokuangTileSource(name string, src LuokuangTileSource) {
	l.Sources[name] = src
}

func (l *ProxySetting) AddArcgisSource(name string, src ArcgisSource) {
	l.Sources[name] = src
}

func (l *ProxySetting) AddSeed(name string, s Seed) {
	l.Sources[name] = s
}

func (l *ProxySetting) AddCleanup(name string, c Cleanup) {
	l.Sources[name] = c
}
