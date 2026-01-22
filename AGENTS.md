# AGENTS.md

This file provides guidance for agentic coding tools working on this Go tileproxy project.

## Build, Lint, and Test Commands

### Build
```bash
go build ./...
```

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./cache
go test ./service
go test ./imagery

# Run a specific test function
go test -run TestLocalCache ./cache
go test -run TestPath ./cache
go test -v -run TestFunctionName ./package_name

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Lint/Format
```bash
go fmt ./...
go vet ./...
```

### Dependencies
```bash
go mod download
go mod tidy
```

## Code Style Guidelines

### Imports
- Group in three blocks: stdlib, external (grouped by source), internal
- Blank lines between groups, sorted alphabetically within groups

### Naming
- Packages: lowercase, single word (e.g., `cache`, `service`, `layer`)
- Exported: PascalCase (e.g., `NewService`, `GetMap`, `ImageSource`)
- Private: camelCase (e.g., `loadGrids`, `checkResRange`)
- Interfaces: descriptive with suffix (e.g., `InfoLayer`, `LegendLayer`, `MapLayer`)
- Receivers: short, first letter of type (e.g., `s *Service`, `l *Layer`)

### Types & Methods
- Constructor functions with `New` prefix: `NewService()`, `NewLocalCache()`
- Use pointer receivers for methods that modify state
- Embed types for composition when appropriate

### Error Handling
- Return errors as second value: `(result, error)`
- Check errors immediately after function calls
- Use simple error messages via `errors.New()`
- Return nil for success cases

### Formatting
- Use `go fmt`, tabs for indentation, no trailing whitespace
- Keep lines under 120 characters when practical
- Minimal comments - explain "why", not "what"

### Project Structure
- One package per directory
- Main packages: `cache`, `client`, `service`, `layer`, `imagery`, `terrain`, `vector`, `tile`, `sources`, `request`, `resource`, `task`, `imports`, `exports`, `utils`
- Test files: `*_test.go` in same package

### Testing
- Use standard `testing` package: `func TestFunctionName(t *testing.T)`
- Create mock/test structs in test files when needed
- Use table-driven tests for multiple test cases

### Common Patterns
- `sync.RWMutex` for concurrent access: `defer unlock` immediately after lock
- Return default nil for not-found cases from maps: `if g, ok := s.Grids[name]; ok { return g }; return nil`
- Use `make()` for maps/slices, pre-allocate capacity when size known
- Type assertions with comma-ok idiom: `if source, ok := src.(type); ok { ... }`

### Constants and Enums
- Use typed constants with iota for enums
```go
type ServiceType uint32
const (
	MapboxService ServiceType = 0
	WMSService    ServiceType = 1
)
```

### Map and Slice Initialization
- Pre-allocate capacity for maps: `m := make(map[string]Layer, len(items))`
- Pre-allocate slices: `s := make([]*Layer, 0, len(layers))`

### Struct Composition
- Embed structs to extend functionality: `type MapLayer struct { Layer; SupportMetaTiles bool }`
- Use interfaces for polymorphism where appropriate

### Resource Management
- Defer cleanup functions immediately after resource acquisition
- Check for nil before accessing struct fields
- Use context for request-scoped values when needed

### Package Guidelines
- Each package should have a clear, single responsibility
- Keep interfaces small and focused (Go best practice)
- Prefer returning concrete types over interfaces when possible
- Avoid circular dependencies between packages

## Domain-Specific Concepts

### Tile Types
- `TILE_IMAGERY`: Raster image tiles (PNG, JPEG, WebP)
- `TILE_VECTOR`: Vector tiles (MVT, PBF, GeoJSON)
- `TILE_DEM`: Digital elevation model tiles (quantized mesh, terrain)

### Service Types
- `MapboxService`: Mapbox Vector Tiles specification
- `WMSService`: Web Map Service (OGC standard)
- `WMTSService`: Web Map Tile Service (OGC standard)
- `CesiumService`: Cesium 3D Tiles specification
- `TileService`: XYZ/OSM tile services

### Layer Types
- `InfoLayer`: Provides feature info capabilities (WMS GetFeatureInfo)
- `LegendLayer`: Provides legend graphic capabilities
- `MapLayer`: Base map rendering layer with extent/resolution info

### Tile Formats
Common formats: `png`, `jpeg`, `webp`, `mvt`, `pbf`, `geojson`, `terrain`, `lerc`
Each format has corresponding MIME type mapping in `tile/format.go`

### Cache Layouts
- `tms`: `/z/x/y.ext` standard XYZ layout
- `quadkey`: Bing Maps quadkey layout
- `arcgis`: ArcGIS tile server layout
- `mp`/`tc`: MapProxy-style layouts

### Tile Coordinate System
- Coordinates are `[z, x, y]` where z=zoom level, x=column, y=row
- Extent-based coordinate systems use bounding box definitions
- Grid objects define projection and tile matrix
