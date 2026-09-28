package setting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validServiceJSON = `{
  "services": [
    {
      "id": "svc1",
      "grids": {
        "g1": {"srs": "EPSG:3857", "bbox": [-20037508.34, -20037508.34, 20037508.34, 20037508.34], "tile_size": [256, 256], "num_levels": 5}
      },
      "sources": {
        "s1": {"type": "tile", "url_template": "https://example.com/{z}/{x}/{y}.png", "grid": "g1", "request_format": "png", "options": {"type": "image", "format": "png"}}
      },
      "caches": {
        "c1": {"grid": "g1", "sources": ["s1"], "format": "png", "request_format": "png", "cache": {"directory_layout": "tms", "directory": "__TMP__"}, "tile_options": {"type": "image", "format": "png"}}
      },
      "service": {"type": "tms", "title": "svc1", "layers": [{"name": "l1", "title": "Layer 1", "source": "c1"}]}
    }
  ]
}`

func validConfig(t *testing.T) []byte {
	t.Helper()
	dir := t.TempDir()
	return []byte(strings.ReplaceAll(validServiceJSON, "__TMP__", filepath.ToSlash(dir)))
}

func TestLoadProxyServicesFromJSON_Multi(t *testing.T) {
	services, err := LoadProxyServicesFromJSON(validConfig(t))
	if err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	svc := services[0]
	if svc.Id != "svc1" {
		t.Fatalf("unexpected id %q", svc.Id)
	}

	if _, ok := svc.Service.(*TMSService); !ok {
		t.Fatalf("service not converted to *TMSService: %T", svc.Service)
	}

	src, ok := svc.Sources["s1"].(*TileSource)
	if !ok {
		t.Fatalf("source not converted to *TileSource: %T", svc.Sources["s1"])
	}
	if src.URLTemplate == "" || src.Grid != "g1" {
		t.Fatalf("tile source fields lost: %+v", src)
	}
	if _, ok := src.Options.(*ImageOpts); !ok {
		t.Fatalf("source options not converted: %T", src.Options)
	}

	cache, ok := svc.Caches["c1"].(*CacheSource)
	if !ok {
		t.Fatalf("cache not converted to *CacheSource: %T", svc.Caches["c1"])
	}
	if cache.Grid != "g1" || len(cache.Sources) != 1 || cache.Sources[0] != "s1" {
		t.Fatalf("cache fields lost: %+v", cache)
	}
	if _, ok := cache.TileOptions.(*ImageOpts); !ok {
		t.Fatalf("cache tile_options not converted: %T", cache.TileOptions)
	}
}

func TestLoadProxyServicesFromJSON_SingleService(t *testing.T) {
	multi := validConfig(t)
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(multi, &doc); err != nil {
		t.Fatal(err)
	}
	// strip the surrounding array to get a bare single service object
	single := strings.TrimSpace(string(doc["services"]))
	single = strings.TrimPrefix(single, "[")
	single = strings.TrimSuffix(single, "]")

	services, err := LoadProxyServicesFromJSON([]byte(single))
	if err != nil {
		t.Fatalf("single service config rejected: %v", err)
	}
	if len(services) != 1 || services[0].Id != "svc1" {
		t.Fatalf("unexpected services: %v", services)
	}
}

func TestLoadProxyServicesFromJSON_RejectsBrokenConfig(t *testing.T) {
	base := string(validConfig(t))

	// grid without srs must be rejected
	broken := strings.Replace(base, `"srs": "EPSG:3857", `, "", 1)
	if _, err := LoadProxyServicesFromJSON([]byte(broken)); err == nil {
		t.Fatal("config with broken grid should be rejected")
	}

	// duplicate service ids must be rejected
	svcObj := strings.TrimSpace(base)
	svcObj = strings.TrimPrefix(svcObj, `{"services": [`)
	svcObj = strings.TrimSuffix(svcObj, `]}`)
	dup := `{"services": [` + svcObj + `, ` + svcObj + `]}`
	if _, err := LoadProxyServicesFromJSON([]byte(dup)); err == nil {
		t.Fatal("duplicate service ids should be rejected")
	}

	// empty documents must be rejected
	if _, err := LoadProxyServicesFromJSON([]byte(`{"services": []}`)); err == nil {
		t.Fatal("empty services should be rejected")
	}
	if _, err := LoadProxyServicesFromJSON([]byte(`{}`)); err == nil {
		t.Fatal("empty document should be rejected")
	}
}

func TestCovertService_MissingTypeDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CovertService panicked on missing type: %v", r)
		}
	}()
	out := CovertService(map[string]interface{}{"title": "no type here"})
	if _, ok := out.(map[string]interface{}); !ok {
		t.Fatalf("expected raw map to be returned untouched, got %T", out)
	}
}

func TestLoadProxyServicesFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, validConfig(t), 0644); err != nil {
		t.Fatal(err)
	}

	services, err := LoadProxyServicesFromFile(path)
	if err != nil {
		t.Fatalf("file config rejected: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	// malformed file must produce an error, not panic
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProxyServicesFromFile(path); err == nil {
		t.Fatal("malformed file should be rejected")
	}
}
