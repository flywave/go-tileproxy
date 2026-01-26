package setting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ConfigValidator interface {
	Validate() error
	GetWarnings() []string
}

func (ps *ProxyService) Validate() error {
	if ps == nil {
		return fmt.Errorf("proxy service is nil")
	}

	if ps.Id == "" {
		return fmt.Errorf("service id is required")
	}

	if ps.Service == nil {
		return fmt.Errorf("service configuration is required")
	}

	warnings := ps.GetWarnings()

	if err := ps.validateGrids(); err != nil {
		return err
	}

	if err := ps.validateSources(); err != nil {
		return err
	}

	if err := ps.validateCaches(); err != nil {
		return err
	}

	if err := ps.validateCoverages(); err != nil {
		return err
	}

	if len(warnings) > 0 {
		fmt.Printf("Configuration warnings for service '%s':\n", ps.Id)
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	return nil
}

func (ps *ProxyService) GetWarnings() []string {
	var warnings []string

	if ps.Service != nil {
		if warnings = append(warnings, ps.validateServiceConfig()...); len(warnings) > 0 {
			return warnings
		}
	}

	if ps.Sources != nil {
		for name, src := range ps.Sources {
			if warnings = append(warnings, ps.validateSourceConfig(name, src)...); len(warnings) > 0 {
				return warnings
			}
		}
	}

	return warnings
}

func (ps *ProxyService) validateGrids() error {
	if ps.Grids == nil {
		return fmt.Errorf("grids configuration is required")
	}

	if len(ps.Grids) == 0 {
		return fmt.Errorf("at least one grid must be defined")
	}

	for name, grid := range ps.Grids {
		if grid.Srs == "" {
			return fmt.Errorf("grid '%s' missing required field: srs", name)
		}

		if grid.BBox == nil && grid.Resolutions == nil {
			return fmt.Errorf("grid '%s' must have either bbox or resolutions defined", name)
		}

		if grid.Resolutions != nil && len(grid.Resolutions) == 0 {
			return fmt.Errorf("grid '%s' resolutions cannot be empty", name)
		}
	}

	return nil
}

func (ps *ProxyService) validateSources() error {
	if ps.Sources == nil {
		return fmt.Errorf("sources configuration is required")
	}

	if len(ps.Sources) == 0 {
		return fmt.Errorf("at least one source must be defined")
	}

	for name, src := range ps.Sources {
		if warnings := ps.validateSourceConfig(name, src); len(warnings) > 0 {
			return fmt.Errorf("source '%s' has validation warnings: %v", name, warnings)
		}

		if err := ps.validateSourceGrid(name, src); err != nil {
			return err
		}
	}

	return nil
}

func (ps *ProxyService) validateSourceGrid(name string, src interface{}) error {
	var gridName string

	switch s := src.(type) {
	case *TileSource:
		gridName = s.Grid
	case *MapboxTileSource:
		gridName = s.Grid
	case *CesiumTileSource:
		gridName = s.Grid
	case *CacheSource:
		gridName = s.Grid
	default:
		return nil
	}

	if gridName != "" {
		if _, ok := ps.Grids[gridName]; !ok {
			return fmt.Errorf("source '%s' references undefined grid: %s", name, gridName)
		}
	}

	return nil
}

func (ps *ProxyService) validateSourceConfig(name string, src interface{}) []string {
	var warnings []string

	switch s := src.(type) {
	case *WMSSource:
		if s.Url == "" {
			return []string{fmt.Sprintf("WMS source '%s' has empty URL", name)}
		}
		if len(s.Layers) == 0 {
			warnings = append(warnings, fmt.Sprintf("WMS source '%s' has no layers defined", name))
		}

	case *TileSource:
		if s.URLTemplate == "" {
			return []string{fmt.Sprintf("Tile source '%s' has empty url_template", name)}
		}
		if s.Grid == "" {
			warnings = append(warnings, fmt.Sprintf("Tile source '%s' has no grid specified", name))
		}

	case *MapboxTileSource:
		if s.Url == "" && len(s.Tiles) == 0 {
			return []string{fmt.Sprintf("Mapbox source '%s' has no URL or tiles defined", name)}
		}
		if s.AccessToken == "" {
			warnings = append(warnings, fmt.Sprintf("Mapbox source '%s' has no access token", name))
		}

	case *CesiumTileSource:
		if s.Url == "" {
			return []string{fmt.Sprintf("Cesium source '%s' has empty URL", name)}
		}
		if s.AccessToken == "" {
			warnings = append(warnings, fmt.Sprintf("Cesium source '%s' has no access token", name))
		}

	case *ArcGISSource:
		if s.Url == "" {
			return []string{fmt.Sprintf("ArcGIS source '%s' has empty URL", name)}
		}
		if len(s.Layers) == 0 {
			warnings = append(warnings, fmt.Sprintf("ArcGIS source '%s' has no layers defined", name))
		}

	case *CacheSource:
		if len(s.Sources) == 0 {
			return []string{fmt.Sprintf("Cache source '%s' has no sources defined", name)}
		}
		if s.Grid == "" {
			warnings = append(warnings, fmt.Sprintf("Cache source '%s' has no grid specified", name))
		}
		for _, srcName := range s.Sources {
			if _, ok := ps.Sources[srcName]; !ok {
				warnings = append(warnings, fmt.Sprintf("Cache source '%s' references undefined source: %s", name, srcName))
			}
		}
	}

	return warnings
}

func (ps *ProxyService) validateServiceConfig() []string {
	var warnings []string

	switch srv := ps.Service.(type) {
	case *WMSService:
		if srv.Title == "" {
			warnings = append(warnings, "WMS service has no title")
		}
		if len(srv.ImageFormats) == 0 {
			warnings = append(warnings, "WMS service has no image formats defined")
		}

	case *TMSService:
		if srv.Title == "" {
			warnings = append(warnings, "TMS service has no title")
		}

	case *WMTSService:
		if srv.Title == "" {
			warnings = append(warnings, "WMTS service has no title")
		}

	case *MapboxService:
		if len(srv.Layers) == 0 {
			warnings = append(warnings, "Mapbox service has no layers defined")
		}

	case *CesiumService:
		if len(srv.Layers) == 0 {
			warnings = append(warnings, "Cesium service has no layers defined")
		}
	}

	return warnings
}

func (ps *ProxyService) validateCaches() error {
	if ps.Caches == nil {
		return fmt.Errorf("caches configuration is required")
	}

	for name, cache := range ps.Caches {
		if c, ok := cache.(*CacheSource); ok {
			if c.CacheInfo != nil && c.CacheInfo.Directory != "" {
				if err := ps.validateCachePath(c.CacheInfo.Directory); err != nil {
					return fmt.Errorf("cache '%s': %v", name, err)
				}
			}

			if c.CacheInfo != nil && c.CacheInfo.DirectoryLayout != "" {
				validLayouts := []string{
					DIRECTORY_LAYOUT_TC,
					DIRECTORY_LAYOUT_MP,
					DIRECTORY_LAYOUT_TMS,
					DIRECTORY_LAYOUT_RE_TMS,
					DIRECTORY_LAYOUT_QUADKEY,
					DIRECTORY_LAYOUT_ARCGIS,
				}
				valid := false
				for _, l := range validLayouts {
					if c.CacheInfo.DirectoryLayout == l {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("cache '%s' has invalid directory_layout: %s", name, c.CacheInfo.DirectoryLayout)
				}
			}

			if c.LockDir != "" {
				if err := ps.validateCachePath(c.LockDir); err != nil {
					return fmt.Errorf("cache '%s': %v", name, err)
				}
			}
		}
	}

	return nil
}

func (ps *ProxyService) validateCachePath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("cache path must be absolute: %s", path)
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %v", err)
		}
	}

	return nil
}

func (ps *ProxyService) validateCoverages() error {
	if ps.Coverages == nil {
		return nil
	}

	for name, cov := range ps.Coverages {
		if cov.Polygons != "" && cov.Geometry != "" {
			return fmt.Errorf("coverage '%s' cannot have both polygons and geometry", name)
		}

		if cov.Polygons != "" && cov.PolygonsSrs == "" {
			return fmt.Errorf("coverage '%s' with polygons must specify polygons_srs", name)
		}

		if cov.Geometry != "" && cov.GeometrySrs == "" {
			return fmt.Errorf("coverage '%s' with geometry must specify geometry_srs", name)
		}

		if cov.BBox != nil && cov.BBoxSrs == "" {
			return fmt.Errorf("coverage '%s' with bbox must specify bbox_srs", name)
		}

		if cov.Union != nil && cov.Intersection != nil {
			return fmt.Errorf("coverage '%s' cannot have both union and intersection", name)
		}

		if (cov.Union != nil || cov.Intersection != nil) && cov.Difference != nil {
			return fmt.Errorf("coverage '%s' cannot have difference with union or intersection", name)
		}
	}

	return nil
}

func IsValidSRS(srs string) bool {
	if srs == "" {
		return false
	}

	srs = strings.ToUpper(strings.TrimSpace(srs))

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

	validPrefixes := []string{
		"CRS:",
		"SR-ORG:",
		"IAN:",
		"IGNF:",
		"OGC:",
		"OSGEO:",
		"ESRI:",
		"IAU2000:",
		"PROJ4:",
		"WKT:",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(srs, prefix) {
			return true
		}
	}

	return false
}
