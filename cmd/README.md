# TileProxy CLI

## Overview

TileProxy CLI provides command-line interface for starting tile proxy servers from configuration files.

## Installation

```bash
# Build CLI
cd cmd
go build -o tileproxy

# Or install system-wide
go install ./cmd/...
```

## Usage

```bash
tileproxy [command] [flags]
```

## Commands

### `serve` - Start tile proxy server

Start a tile proxy server using a configuration file.

**Flags:**
- `-c, --config string` - Path to configuration file (default: "config.json")
- `-p, --port int` - Server port (default: 8000)
- `-H, --host string` - Server host (default: "0.0.0.0")
- `-l, --log-level string` - Log level: debug, info, warn, error (default: "info")

**Examples:**

```bash
# Start with default config (config.json)
tileproxy serve

# Start with custom config
tileproxy serve --config config/wms.json

# Start with custom port
tileproxy serve --port 8080

# Start with custom host and port
tileproxy serve --host 127.0.0.1 --port 9000
```

## Configuration Files

### Supported Formats

- **JSON** - Fully supported with `CreateProxyServiceFromJSON()`
- **YAML** - Planned (requires conversion or direct YAML parser)

### Configuration Structure

```json
{
  "id": "service-id",
  "service": { ... },
  "grids": { ... },
  "sources": { ... },
  "caches": { ... }
}
```

### Example Configurations

#### TMS Service (config/examples/tms.json)

```json
{
  "id": "my-tileproxy",
  "service": {
    "type": "tms",
    "layers": [
      {
        "source": "osm_cache",
        "name": "osm"
      }
    ]
  },
  "grids": {
    "global_webmercator": {
      "name": "GLOBAL_WEB_MERCATOR",
      "srs": "EPSG:3857",
      "origin": "ul"
    }
  },
  "sources": {
    "osm": {
      "type": "tile",
      "req": {
        "url": "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        "grid": "global_webmercator"
      }
    }
  },
  "caches": {
    "osm_cache": {
      "sources": ["osm"],
      "name": "osm_cache",
      "grid": "global_webmercator",
      "format": "png",
      "request_format": "png",
      "cache": {
        "directory": "./cache/osm",
        "directory_layout": "tms"
      }
    }
  }
}
```

#### WMS Service (config/examples/wms.json)

```json
{
  "id": "wms-proxy",
  "service": {
    "type": "wms",
    "srs": ["EPSG:3857", "EPSG:4326"],
    "image_formats": ["image/png", "image/jpeg"],
    "layers": [ ... ]
  },
  ...
}
```

#### Mapbox Service (config/examples/mapbox.json)

```json
{
  "id": "mapbox-proxy",
  "service": {
    "type": "mapbox",
    "layers": [ ... ]
  },
  ...
}
```

## Service Types

### TMS (Tile Map Service)

Type: `tms`

Features:
- XYZ tile format
- Multiple layers
- Caching support

### WMS (Web Map Service)

Type: `wms`

Features:
- OGC WMS 1.1.1
- GetMap, GetFeatureInfo, GetLegendGraphic
- Multiple SRS support

### WMTS (Web Map Tile Service)

Type: `wmts`

Features:
- OGC WMTS 1.0.0
- RESTful and KVP
- Tile matrix sets

### Mapbox

Type: `mapbox`

Features:
- Vector tiles (MVT/PBF)
- Raster tiles
- TileJSON format

### Cesium

Type: `cesium`

Features:
- Quantized mesh terrain
- Layer JSON
- Tile data format

## Grid Systems

### Global Web Mercator

```json
{
  "global_webmercator": {
    "name": "GLOBAL_WEB_MERCATOR",
    "srs": "EPSG:3857",
    "origin": "ul"
  }
}
```

### Global Geodetic

```json
{
  "global_geodetic": {
    "name": "GLOBAL_GEODETIC",
    "srs": "EPSG:4326",
    "origin": "sw"
  }
}
```

## Source Types

### Tile Source

```json
{
  "type": "tile",
  "req": {
    "url": "https://example.com/{z}/{x}/{y}.png",
    "grid": "global_webmercator",
    "subdomains": ["0", "1", "2", "3"]
  }
}
```

### WMS Source

```json
{
  "type": "wms",
  "req": {
    "url": "https://wms.example.com/service",
    "layers": "layer_name",
    "srs": ["EPSG:3857"],
    "transparent": true,
    "format": "image/png"
  }
}
```

### Mapbox Source

```json
{
  "type": "mapbox",
  "req": {
    "url": "https://api.mapbox.com/v4/mapbox.streets-v8/{z}/{x}/{y}.vector.pbf",
    "grid": "global_webmercator",
    "format": "mvt"
  },
  "transparent": true
}
```

## Cache Configuration

```json
{
  "cache_name": {
    "sources": ["source_name"],
    "name": "cache_name",
    "grid": "global_webmercator",
    "format": "png",
    "request_format": "png",
    "cache": {
      "directory": "./cache/path",
      "directory_layout": "tms"
    }
  }
}
```

### Directory Layouts

- `tms` - Standard TMS layout (`{z}/{x}/{y}.ext`)
- `quadkey` - Bing Maps quadkey layout
- `arcgis` - ArcGIS tile server layout
- `mp` - MapProxy layout
- `tc` - TileCache layout

## Health Check

All services provide a health check endpoint:

```bash
curl http://localhost:8000/health
```

Response:
```json
{
  "health": {
    "status": "healthy"
  }
}
```

## Usage Examples

### Start TMS Service

```bash
# Copy example config
cp config/examples/tms.json my-config.json

# Edit config with your sources
vim my-config.json

# Start server
tileproxy serve --config my-config.json --port 8000
```

### Start WMS Service

```bash
# Use WMS example
tileproxy serve --config config/examples/wms.json --port 8001
```

### Start Mapbox Service

```bash
# Use Mapbox example (requires API key)
tileproxy serve --config config/examples/mapbox.json --port 8002
```

## URL Patterns

### TMS

```
http://localhost:8000/{layer}/{z}/{x}/{y}.png
```

### WMS

```
http://localhost:8000/?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetMap&...
```

### Mapbox

```
http://localhost:8000/{layer}/source.json
http://localhost:8000/{layer}/{z}/{x}/{y}.vector.pbf
```

### Cesium

```
http://localhost:8000/{layer}/layer.json
http://localhost:8000/{layer}/{z}/{x}/{y}.terrain
```

## Production Deployment

### Systemd Service

```ini
# /etc/systemd/system/tileproxy.service
[Unit]
Description=TileProxy Server
After=network.target

[Service]
Type=simple
User=tileproxy
WorkingDirectory=/opt/tileproxy
ExecStart=/usr/local/bin/tileproxy serve --config /etc/tileproxy/config.json
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN cd cmd && go build -o tileproxy

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/cmd/tileproxy /app/tileproxy
COPY config /app/config
EXPOSE 8000
CMD ["./tileproxy", "serve", "--config", "config/tms.json"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tileproxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: tileproxy
  template:
    metadata:
      labels:
        app: tileproxy
    spec:
      containers:
      - name: tileproxy
        image: tileproxy:latest
        ports:
        - containerPort: 8000
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 10
        volumeMounts:
        - name: config
          mountPath: /config
        - name: cache
          mountPath: /cache
      volumes:
      - name: config
        configMap:
          name: tileproxy-config
      - name: cache
        persistentVolumeClaim:
          claimName: tileproxy-cache
```

## Troubleshooting

### Config File Not Found

```bash
# Check if file exists
ls -la config.json

# Check file path
tileproxy serve --config ./config.json
```

### Port Already in Use

```bash
# Find process using port
lsof -i :8000

# Kill process
kill -9 <PID>

# Or use different port
tileproxy serve --port 8080
```

### Service Won't Start

```bash
# Validate config JSON
cat config.json | jq empty

# Check logs
journalctl -u tileproxy -f
```

## Contributing

To add new features:
1. Update cmd/*.go files
2. Add tests
3. Update documentation

## License

See project LICENSE file.
