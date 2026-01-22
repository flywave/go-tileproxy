# Go-TileProxy Demo Collection

This directory contains multiple demo applications demonstrating various tile proxy configurations and services.

## Demo List

| Demo | Port | Type | Source | Description |
|------|------|------|--------|-------------|
| **osm** | 8001 | TMS | OpenStreetMap | Standard XYZ tile service |
| **wms** | 8000 | WMS | Omniscale | WMS 1.1.1 proxy service |
| **mapbox** | 8001 | Mapbox | Mapbox | Vector, Raster, and DEM tiles |
| **cesium** | 8001 | Cesium | Cesium Ion | Quantized mesh terrain |
| **mapbox_to_cesium** | 8001 | Mapbox→Cesium | Mapbox DEM | Convert Mapbox DEM to Cesium terrain |
| **wmts** | 8001 | WMTS | Basemap.at | WMTS restful service |
| **luokuang** | 8000 | Mapbox | Luokuang | Chinese vector tiles |
| **amap** | 8003 | TMS | Amap (高德) | Chinese map tiles |
| **google** | 8004 | TMS | Google Maps | Standard Google tiles |
| **bingmap** | 8005 | TMS | Bing Maps | Quadkey tile format |
| **baidu** | 8006 | TMS | Baidu (百度) | Chinese map with BD-09 coordinates |
| **tianditu** | 8007 | WMTS | TianDiTu (天地图) | Multi-layer Chinese WMTS |
| **arcgisonline** | 8008 | TMS | ArcGIS Online | Multiple ESRI map services |
| **template** | 8009 | Template | - | Developer starter template |

## Usage

### Running a Demo

Each demo is a standalone Go application. Navigate to the demo directory and run:

```bash
cd demo/{demo_name}
go run *.go
```

Or build first:

```bash
cd demo/{demo_name}
go build -o {demo_name}
./{demo_name}
```

### Accessing Tiles

**TMS Services** (osm, amap, google, bing, baidu, arcgisonline):
```
http://127.0.0.1:{port}/{layer}/{z}/{x}/{y}.png
```

Example:
```
http://127.0.0.1:8003/amap_layer/0/0/0.png
```

**WMS Services** (wms, tianditu):
```
http://127.0.0.1:{port}/?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetMap&LAYERS={layer}&...
```

**Mapbox Services** (mapbox, luokuang):
```
http://127.0.0.1:{port}/{layer}/source.json
http://127.0.0.1:{port}/{layer}/{z}/{x}/{y}.{format}
```

**Cesium Services** (cesium, mapbox_to_cesium):
```
http://127.0.0.1:{port}/{layer}/layer.json
http://127.0.0.1:{port}/{layer}/{z}/{x}/{y}.terrain
```

**WMTS Services** (wmts, tianditu):
```
http://127.0.0.1:{port}/?SERVICE=WMTS&REQUEST=GetTile&...
```

## Demo Details

### Phase 1: Core Services

#### 1. AMap (高德地图)
- **Port**: 8003
- **Source**: https://webrd{1-4}.is.autonavi.com/appmaptile
- **Features**:
  - Chinese street map
  - TMS tile format
  - Chinese language labels

#### 2. Google Maps
- **Port**: 8004
- **Source**: https://mt{0-3}.googleapis.com/vt
- **Features**:
  - Standard Google tiles
  - Multiple subdomains for load balancing
  - TMS tile format

#### 3. Bing Maps
- **Port**: 8005
- **Source**: https://ecn.t{0-3}.tiles.virtualearth.net/tiles
- **Features**:
  - Quadkey tile numbering system
  - JPEG format
  - Aerial imagery support

### Phase 2: Chinese Services

#### 4. Baidu Maps (百度地图)
- **Port**: 8006
- **Source**: http://online{0-3}.map.bdimg.com/onlinelabel
- **Features**:
  - BD-09 coordinate system
  - Chinese street map with POI labels
  - Standard TMS format

#### 5. TianDiTu (天地图)
- **Port**: 8007
- **Source**: http://t{0-6}.tianditu.com
- **Features**:
  - WMTS 1.0.0 service
  - Multiple layers:
    - `vec_w` - Vector map
    - `cva_w` - Vector annotation
    - `img_w` - Imagery
    - `cia_w` - Imagery annotation
    - `ter_w` - Terrain
    - `cta_w` - Terrain annotation
  - Supports EPSG:3857 and EPSG:4326

### Phase 3: Extensions

#### 6. ArcGIS Online
- **Port**: 8008
- **Source**: https://services.arcgisonline.com/arcgis/rest/services
- **Features**:
  - Multiple ESRI map services:
    - World Imagery
    - World Street Map
    - World Topo Map
    - NatGeo World Map
  - TMS tile format
  - Standard Web Mercator projection

#### 7. Template
- **Port**: 8009
- **Purpose**: Developer starter kit
- **Features**:
  - Complete demo structure
  - Detailed inline comments
  - Easy customization
  - Example configurations

## Configuration

### Cache Storage

All demos store cached tiles in `./cache_data/{demo_name}/` directory structure.

Supported layouts:
- `tms` - Standard TMS layout (`{z}/{x}/{y}.ext`)
- Custom layouts via configuration

### Grid Options

Available grids (defined in `grids.go`):
- `global_webmercator` - EPSG:3857 (Web Mercator)
- `global_geodetic` - EPSG:4326 (WGS84)
- `global_geodetic_cgcs2000` - EPSG:4490 (China)
- `global_mercator_cgcs2000` - EPSG:4479 (China)
- And more...

### Global Settings

Global settings are defined in `globals.go`:
- Projection data directory
- Geoid data directory
- Cache configuration
- Image processing options
- HTTP settings

## Creating New Demos

1. **Use the Template**:
   ```bash
   cp -r demo/template demo/my_new_demo
   ```

2. **Customize**:
   - Update `API_URL` constant
   - Modify service name
   - Change port number
   - Adjust cache directory

3. **Test**:
   ```bash
   cd demo/my_new_demo
   go run main.go
   ```

## Service Types

### TMS Service
```go
setting.TMSService{
    Layers: []setting.TileLayer{
        {Source: "cache_name", Name: "layer_name"},
    },
}
```

### WMS Service
```go
setting.WMSService{
    Srs:          []string{"EPSG:4326", "EPSG:3857"},
    ImageFormats:  []string{"image/png", "image/jpeg"},
    Layers:       []setting.WMSLayer{{...}},
    MaxOutputPixels: setting.NewInt(2000 * 2000),
}
```

### Mapbox Service
```go
setting.MapboxService{
    Layers: []setting.MapboxTileLayer{
        {
            Source:    "cache_name",
            Name:      "layer_name",
            TileJSON:  "tilejson_name",
            ZoomRange: &[2]int{0, 20},
        },
    },
}
```

### Cesium Service
```go
setting.CesiumService{
    Layers: []setting.CesiumTileLayer{
        {
            Source:    "cache_name",
            Name:      "layer_name",
            LayerJSON: "layerjson_name",
            ZoomRange: &[2]int{0, 20},
        },
    },
}
```

## Troubleshooting

### Port Already in Use
```bash
# Find process using the port
lsof -i :8003

# Kill the process
kill -9 <PID>
```

### Cache Directory Not Found
Ensure you run the demo from the correct directory or adjust cache paths in the code.

### Tile Not Found
- Check source URL is correct
- Verify network connectivity to tile server
- Check log output for error messages

## Contributing

To add new demos:

1. Create a new directory under `demo/`
2. Copy and modify the template
3. Follow the established naming conventions
4. Update this README with details

## License

See the main project LICENSE file.
