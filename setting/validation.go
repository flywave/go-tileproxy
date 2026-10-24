package setting

import (
	"encoding/json"

	"errors"
	"fmt"
	cachepolicy "github.com/flywave/go-tileproxy/cache"
	"regexp"
	"strings"
)

// FieldIssue points at one concrete configuration problem in a way a front
// end can render directly: a field path inside the service object plus a
// human readable message.
type FieldIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (i FieldIssue) String() string {
	if i.Field == "" {
		return i.Message
	}
	return i.Field + ": " + i.Message
}

// ServiceValidationReport is the validation result for a single service.
type ServiceValidationReport struct {
	Id       string       `json:"id"`
	Index    int          `json:"index"`
	Valid    bool         `json:"valid"`
	Errors   []FieldIssue `json:"errors,omitempty"`
	Warnings []FieldIssue `json:"warnings,omitempty"`
}

// Error flattens the report's errors into a single error.
func (r *ServiceValidationReport) Error() error {
	if r.Valid {
		return nil
	}
	var errs []error
	for _, e := range r.Errors {
		errs = append(errs, errors.New(e.String()))
	}
	return errors.Join(errs...)
}

// ConfigValidationReport is the result of validating a whole configuration
// document. It is designed to be returned as-is by a validation endpoint:
// Valid reports whether the configuration may be applied, Errors holds
// document level problems (syntax, structure) and Services carries the
// per-service findings.
type ConfigValidationReport struct {
	Valid    bool                      `json:"valid"`
	Errors   []FieldIssue              `json:"errors,omitempty"`
	Services []ServiceValidationReport `json:"services"`
}

// Error flattens the report into a single error for callers that only need
// a success/failure signal.
func (r *ConfigValidationReport) Error() error {
	if r.Valid {
		return nil
	}
	var errs []error
	for _, e := range r.Errors {
		errs = append(errs, errors.New(e.String()))
	}
	for _, s := range r.Services {
		for _, e := range s.Errors {
			errs = append(errs, fmt.Errorf("service %q: %s", s.Id, e.String()))
		}
	}
	return errors.Join(errs...)
}

var serviceIdPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

type issueList struct {
	path     string
	errs     *[]FieldIssue
	warnings *[]FieldIssue
}

func (l issueList) errf(field, format string, args ...interface{}) {
	*l.errs = append(*l.errs, FieldIssue{Field: l.join(field), Message: fmt.Sprintf(format, args...)})
}

func (l issueList) warnf(field, format string, args ...interface{}) {
	*l.warnings = append(*l.warnings, FieldIssue{Field: l.join(field), Message: fmt.Sprintf(format, args...)})
}

func (l issueList) join(field string) string {
	if l.path == "" {
		return field
	}
	if field == "" {
		return l.path
	}
	return l.path + "." + field
}

// ValidateConfigJSON validates a raw configuration document: JSON syntax,
// document structure and the full per-service checks. It never panics and
// always returns a report, even for broken documents, so a front end can
// show every problem in one round trip.
func ValidateConfigJSON(data []byte) *ConfigValidationReport {
	report := &ConfigValidationReport{Valid: true, Services: []ServiceValidationReport{}}
	fail := func(format string, args ...interface{}) {
		report.Valid = false
		report.Errors = append(report.Errors, FieldIssue{Message: fmt.Sprintf(format, args...)})
	}

	if len(data) > MaxJSONSize {
		fail("config too large (max %d bytes)", MaxJSONSize)
		return report
	}

	var doc struct {
		Services []*ProxyService `json:"services"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		fail("invalid JSON document: %v", err)
		return report
	}

	raws := doc.Services
	if doc.Services == nil {
		var single ProxyService
		if err := json.Unmarshal(data, &single); err != nil {
			fail("expected {\"services\":[...]} or a single service object")
			return report
		}
		raws = []*ProxyService{&single}
	}
	if len(raws) == 0 {
		fail("no services defined")
		return report
	}

	seen := make(map[string]int, len(raws))
	for i, ps := range raws {
		rep := ValidateProxyService(ps)
		rep.Index = i
		if first, dup := seen[ps.Id]; dup {
			rep.Valid = false
			rep.Errors = append(rep.Errors, FieldIssue{
				Field:   "id",
				Message: fmt.Sprintf("duplicate service id %q (also used by service #%d)", ps.Id, first),
			})
		} else {
			seen[ps.Id] = i
		}
		if !rep.Valid {
			report.Valid = false
		}
		report.Services = append(report.Services, *rep)
	}
	return report
}

// ValidateProxyService performs the full static validation of one service:
// required fields, unknown types, references between grids, sources, caches
// and layers, and URL safety of the service id. The service is converted to
// its typed form first, so raw front-end payloads can be passed in directly.
// All findings are reported, not just the first one.
func ValidateProxyService(ps *ProxyService) *ServiceValidationReport {
	rep := &ServiceValidationReport{Valid: true, Errors: []FieldIssue{}, Warnings: []FieldIssue{}}

	if ps == nil {
		rep.Valid = false
		rep.Errors = append(rep.Errors, FieldIssue{Message: "service definition is null"})
		return rep
	}

	// normalize raw front-end payloads into typed structs; this also makes
	// the service immediately buildable for the proxy
	ps.Covert()

	rep.Id = ps.Id

	errs := &rep.Errors
	warns := &rep.Warnings

	// ---- id: it becomes a URL path segment, so it must be URL-safe ----
	if ps.Id == "" {
		*errs = append(*errs, FieldIssue{Field: "id", Message: "service id is required"})
		rep.Valid = false
	} else if !serviceIdPattern.MatchString(ps.Id) {
		*errs = append(*errs, FieldIssue{Field: "id", Message: "service id must match [a-zA-Z0-9][a-zA-Z0-9_-]{0,63}"})
		rep.Valid = false
	} else if strings.EqualFold(ps.Id, "health") {
		// "/health" is handled by the proxy itself
		*errs = append(*errs, FieldIssue{Field: "id", Message: "service id \"health\" is reserved"})
		rep.Valid = false
	}

	root := issueList{errs: errs, warnings: warns}

	// ---- service: known type required, then per-type layer references ----
	switch svc := ps.Service.(type) {
	case nil:
		root.errf("service", "service configuration is required")
		rep.Valid = false
	case map[string]interface{}:
		ty, _ := svc["type"].(string)
		root.errf("service.type", "unknown or missing service type %q (expected wms, tms, wmts, mapbox or cesium)", ty)
		rep.Valid = false
	case *WMSService:
		validateWMSService(ps, root)
	case *TMSService:
		validateLayerRefs(ps, "service.layers", layersOfTile(svc.Layers), root)
	case *WMTSService:
		validateLayerRefs(ps, "service.layers", layersOfTile(svc.Layers), root)
	case *MapboxService:
		refs := make([]string, 0, len(svc.Layers))
		for _, l := range svc.Layers {
			refs = append(refs, l.Source)
		}
		validateLayerRefs(ps, "service.layers", refs, root)
	case *CesiumService:
		refs := make([]string, 0, len(svc.Layers))
		for _, l := range svc.Layers {
			refs = append(refs, l.Source)
		}
		validateLayerRefs(ps, "service.layers", refs, root)
	default:
		root.errf("service", "unsupported service configuration of type %T", ps.Service)
		rep.Valid = false
	}

	validateGridsField(ps, root)
	validateSourcesField(ps, root)
	validateCachesField(ps, root)
	validateCoveragesField(ps, root)

	if len(rep.Errors) > 0 {
		rep.Valid = false
	}
	return rep
}

func layersOfTile(layers []TileLayer) []string {
	refs := make([]string, 0, len(layers))
	for _, l := range layers {
		refs = append(refs, l.Source)
	}
	return refs
}

// validateLayerRefs checks that every layer source reference resolves to a
// cache or source defined by this service.
func validateLayerRefs(ps *ProxyService, field string, refs []string, root issueList) {
	for i, ref := range refs {
		if ref == "" {
			root.errf(fmt.Sprintf("%s.%d.source", field, i), "layer source reference is required")
			continue
		}
		_, inCaches := ps.Caches[ref]
		_, inSources := ps.Sources[ref]
		if !inCaches && !inSources {
			root.errf(fmt.Sprintf("%s.%d.source", field, i), "references undefined cache or source %q", ref)
		}
	}
}

func validateWMSService(ps *ProxyService, root issueList) {
	svc := ps.Service.(*WMSService)
	if len(svc.ImageFormats) == 0 {
		root.warnf("service.image_formats", "no image formats defined; clients cannot request maps")
	}
	for i, srs := range svc.Srs {
		if !IsValidSRS(srs) {
			root.errf(fmt.Sprintf("service.srs.%d", i), "invalid SRS %q", srs)
		}
	}
	refs := make([]string, 0, len(svc.Layers)*3)
	for _, l := range svc.Layers {
		refs = append(refs, l.Sources...)
		refs = append(refs, l.FeatureinfoSources...)
		refs = append(refs, l.LegendSources...)
	}
	validateLayerRefs(ps, "service.layers", refs, root)
}

func validateGridsField(ps *ProxyService, root issueList) {
	if len(ps.Grids) == 0 {
		root.errf("grids", "at least one grid must be defined")
		return
	}
	for name, grid := range ps.Grids {
		path := issueList{path: "grids." + name, errs: root.errs, warnings: root.warnings}
		if grid.Srs == "" {
			path.errf("srs", "srs is required")
		} else if !IsValidSRS(grid.Srs) {
			path.errf("srs", "invalid SRS %q", grid.Srs)
		}
		if grid.BBox == nil && grid.Resolutions == nil {
			// go-geo falls back to a world-wide extent with computed
			// resolutions; warn instead of failing legacy configs
			path.warnf("bbox", "neither bbox nor resolutions defined; the world-wide default extent will be used")
		}
		if grid.BBox != nil {
			b := grid.BBox
			if b[0] >= b[2] || b[1] >= b[3] {
				path.errf("bbox", "invalid bbox [minx miny maxx maxy]: min values must be smaller than max values")
			}
		}
		if grid.Resolutions != nil && len(grid.Resolutions) == 0 {
			path.errf("res", "resolutions cannot be empty")
		}
		if grid.TileSize != nil && (grid.TileSize[0] == 0 || grid.TileSize[1] == 0) {
			path.errf("tile_size", "tile size must be greater than zero")
		}
	}
}

func validateSourcesField(ps *ProxyService, root issueList) {
	for name, src := range ps.Sources {
		path := issueList{path: "sources." + name, errs: root.errs, warnings: root.warnings}
		if mp, raw := src.(map[string]interface{}); raw {
			ty, _ := mp["type"].(string)
			path.errf("type", "unknown or missing source type %q (expected wms, tile, mapbox, cesium or arcgis)", ty)
			continue
		}
		checkSourceFields(ps, src, path)
	}
}

func checkSourceFields(ps *ProxyService, src interface{}, path issueList) {
	checkGridRef := func(grid string) {
		if grid == "" {
			path.errf("grid", "grid is required")
			return
		}
		if _, ok := ps.Grids[grid]; !ok {
			path.errf("grid", "references undefined grid %q", grid)
		}
	}
	switch s := src.(type) {
	case *TileSource:
		if s.URLTemplate == "" {
			path.errf("url_template", "url_template is required")
		}
		checkGridRef(s.Grid)
	case *MapboxTileSource:
		if s.Url == "" && len(s.Tiles) == 0 {
			path.errf("url", "either url or tiles must be defined")
		}
		checkGridRef(s.Grid)
	case *CesiumTileSource:
		if s.Url == "" {
			path.errf("url", "url is required")
		}
		checkGridRef(s.Grid)
	case *ArcGISSource:
		if s.Url == "" {
			path.errf("url", "url is required")
		}
		if len(s.Layers) == 0 {
			path.errf("layers", "at least one layer must be defined")
		}
	case *WMSSource:
		if s.Url == "" {
			path.errf("url", "url is required")
		}
		if len(s.Layers) == 0 {
			path.warnf("layers", "no layers defined")
		}
	}
}

func validateCachesField(ps *ProxyService, root issueList) {
	for name, c := range ps.Caches {
		path := issueList{path: "caches." + name, errs: root.errs, warnings: root.warnings}
		cs, ok := c.(*CacheSource)
		if !ok {
			path.errf("type", "unknown or missing cache type (expected \"cache\")")
			continue
		}
		if cs.Grid == "" {
			path.errf("grid", "grid is required")
		} else if _, ok := ps.Grids[cs.Grid]; !ok {
			path.errf("grid", "references undefined grid %q", cs.Grid)
		}
		if cs.CacheInfo == nil || cs.CacheInfo.Directory == "" {
			path.errf("cache.directory", "cache directory is required")
		} else if !validCacheLayout(cs.CacheInfo.DirectoryLayout) {
			path.errf("cache.directory_layout", "invalid directory_layout %q", cs.CacheInfo.DirectoryLayout)
		}
		if cs.TileOptions == nil {
			path.errf("tile_options", "tile_options with a known type (image, raster or vector) is required")
		}
		for _, ref := range cs.Sources {
			if _, ok := ps.Sources[ref]; !ok {
				path.warnf("sources", "references undefined source %q", ref)
			}
		}
		if cs.CachePolicy != nil {
			ttl := 0
			if cs.CachePolicy.TtlHours != nil {
				ttl = *cs.CachePolicy.TtlHours
			}
			if _, err := cachepolicy.ParseCachePolicy(cs.CachePolicy.Mode, ttl); err != nil {
				path.errf("cache_policy", "%v", err)
			}
		}
	}
}

func validCacheLayout(layout string) bool {
	switch layout {
	case DIRECTORY_LAYOUT_TC, DIRECTORY_LAYOUT_MP, DIRECTORY_LAYOUT_TMS,
		DIRECTORY_LAYOUT_RE_TMS, DIRECTORY_LAYOUT_QUADKEY, DIRECTORY_LAYOUT_ARCGIS:
		return true
	}
	return false
}

func validateCoveragesField(ps *ProxyService, root issueList) {
	for name, cov := range ps.Coverages {
		path := issueList{path: "coverages." + name, errs: root.errs, warnings: root.warnings}
		if cov.Polygons != "" && cov.Geometry != "" {
			path.errf("", "cannot have both polygons and geometry")
		}
		if cov.Polygons != "" && cov.PolygonsSrs == "" {
			path.errf("polygons_srs", "polygons_srs is required when polygons is set")
		}
		if cov.Geometry != "" && cov.GeometrySrs == "" {
			path.errf("geometry_srs", "geometry_srs is required when geometry is set")
		}
		if cov.BBox != nil && cov.BBoxSrs == "" {
			path.errf("bbox_srs", "bbox_srs is required when bbox is set")
		}
		if cov.Union != nil && cov.Intersection != nil {
			path.errf("", "cannot have both union and intersection")
		}
		if (cov.Union != nil || cov.Intersection != nil) && cov.Difference != nil {
			path.errf("", "cannot have difference together with union or intersection")
		}
	}
}

var validSrsPrefixes = []string{
	"CRS:", "SR-ORG:", "IAN:", "IGNF:", "OGC:", "OSGEO:", "ESRI:", "IAU2000:", "PROJ4:", "WKT:",
}

// IsValidSRS reports whether srs looks like a supported SRS identifier
// (either EPSG:<num> or one of the known prefixed namespaces).
func IsValidSRS(srs string) bool {
	if srs == "" {
		return false
	}
	srs = strings.ToUpper(strings.TrimSpace(srs))
	switch srs {
	case "EPSG:GCJ02", "GCJ02", "EPSG:GCJ02MC", "GCJ02MC", "EPSG:BD09", "BD09", "EPSG:BDMC", "EPSG:BD09MC", "BD09MC", "GCJ02:3857":
		return true
	}

	if strings.HasPrefix(srs, "EPSG:") {
		num := strings.TrimPrefix(srs, "EPSG:")
		if num == "" {
			return false
		}
		for _, c := range num {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	for _, prefix := range validSrsPrefixes {
		if strings.HasPrefix(srs, prefix) {
			return true
		}
	}
	return false
}
