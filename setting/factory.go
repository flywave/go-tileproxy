package setting

import (
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/service"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/terrain"
	"github.com/flywave/go-tileproxy/vector"
)

func ConvertImageOptions(opt *ImageOpts) *imagery.ImageOptions {
	return nil
}

func ConvertRasterOptions(opt *RasterOpts) *terrain.RasterOptions {
	return nil
}

func ConvertTerrainOptions(opt *RasterOpts) *terrain.TerrainOptions {
	return nil
}

func ConvertVectorOptions(opt *VectorOpts) *vector.VectorOptions {
	return nil
}

func ConvertGridOpts(opt *GridOpts) *geo.TileGrid {
	return nil
}

func ConvertLocalCache(opt *LocalCache) *cache.LocalCache {
	return nil
}

func ConvertS3Cache(opt *S3Cache) *cache.S3Cache {
	return nil
}

func ConvertMapboxTileLayer(l *MapboxTileLayer) *service.MapboxTileProvider {
	return nil
}

func ConvertTileLayer(l *TileLayer) *service.TileProvider {
	return nil
}

func ConvertWMSLayer(l *WMSLayer) *service.WMSLayer {
	return nil
}

func LoadWMSRootLayer(l *WMSLayer) *service.WMSGroupLayer {
	return nil
}

func LoadWMSSource(s *WMSSource) *sources.WMSSource {
	return nil
}

func LoadTileSource(s *TileSource) *sources.TileSource {
	return nil
}

func LoadMapboxTileSource(s *MapboxTileSource) *sources.MapboxTileSource {
	return nil
}

func LoadLuokuangTileSource(s *LuokuangTileSource) *sources.LuoKuangTileSource {
	return nil
}

func LoadArcgisSource(s *ArcgisSource) *sources.ArcGISSource {
	return nil
}

func LoadMapboxService(s *MapboxService) *service.MapboxService {
	return nil
}

func LoadTMSService(s *TMSService) *service.TileService {
	return nil
}

func LoadWMTSService(s *WMTSService) *service.WMTSService {
	return nil
}

func LoadWMSService(s *WMSService) *service.WMSService {
	return nil
}
