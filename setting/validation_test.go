package setting

import (
	"encoding/json"
	"strings"
	"testing"
)

const fullValidConfig = `{
  "services": [
    {
      "id": "svc1",
      "grids": {"g1": {"srs": "EPSG:3857", "bbox": [-20037508.34, -20037508.34, 20037508.34, 20037508.34], "tile_size": [256, 256]}},
      "sources": {
        "s1": {"type": "tile", "url_template": "https://example.com/{z}/{x}/{y}.png", "grid": "g1", "request_format": "png", "options": {"type": "image", "format": "png"}}
      },
      "caches": {
        "c1": {"grid": "g1", "sources": ["s1"], "format": "png", "request_format": "png", "cache": {"directory_layout": "tms", "directory": "/tmp/c1"}, "tile_options": {"type": "image", "format": "png"}}
      },
      "service": {"type": "tms", "title": "svc1", "layers": [{"name": "l1", "title": "L1", "source": "c1"}]}
    }
  ]
}`

func TestValidateConfigJSON_Valid(t *testing.T) {
	report := ValidateConfigJSON([]byte(fullValidConfig))
	if !report.Valid {
		t.Fatalf("valid config reported invalid: %v", report.Error())
	}
	if len(report.Services) != 1 {
		t.Fatalf("expected 1 service report, got %d", len(report.Services))
	}
	svc := report.Services[0]
	if svc.Id != "svc1" || !svc.Valid {
		t.Fatalf("unexpected service report: %+v", svc)
	}
}

func TestValidateConfigJSON_MissingFields(t *testing.T) {
	doc := strings.Replace(fullValidConfig, `"url_template": "https://example.com/{z}/{x}/{y}.png", `, "", 1)
	doc = strings.Replace(doc, `"directory_layout": "tms", `, "", 1)

	report := ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("config with missing fields must be invalid")
	}
	svc := report.Services[0]
	if svc.Valid {
		t.Fatal("service with missing fields must be invalid")
	}

	body, _ := json.Marshal(report)
	joined := string(body)
	if !strings.Contains(joined, "url_template") || !strings.Contains(joined, "url_template is required") {
		t.Fatalf("missing url_template not reported: %s", joined)
	}
	if !strings.Contains(joined, "directory_layout") {
		t.Fatalf("missing directory_layout not reported: %s", joined)
	}
}

func TestValidateConfigJSON_UnknownTypes(t *testing.T) {
	// unknown source type must be reported instead of being silently ignored
	doc := strings.Replace(fullValidConfig, `"type": "tile"`, `"type": "what-is-this"`, 1)
	report := ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("unknown source type must be invalid")
	}
	body, _ := json.Marshal(report)
	if !strings.Contains(string(body), "unknown or missing source type") {
		t.Fatalf("unknown source type not reported: %s", body)
	}

	// unknown service type must be reported
	doc = strings.Replace(fullValidConfig, `"type": "tms"`, `"type": "nope"`, 1)
	report = ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("unknown service type must be invalid")
	}
	body, _ = json.Marshal(report)
	if !strings.Contains(string(body), "unknown or missing service type") {
		t.Fatalf("unknown service type not reported: %s", body)
	}
}

func TestValidateConfigJSON_BrokenReferences(t *testing.T) {
	// layer referencing a cache that does not exist
	doc := strings.Replace(fullValidConfig, `"source": "c1"`, `"source": "nope"`, 1)
	report := ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("dangling layer reference must be invalid")
	}
	body, _ := json.Marshal(report)
	if !strings.Contains(string(body), "references undefined cache or source") {
		t.Fatalf("dangling reference not reported: %s", body)
	}

	// source referencing an undefined grid
	doc = strings.Replace(fullValidConfig, `"grid": "g1", "request_format": "png", "options"`, `"grid": "missing", "request_format": "png", "options"`, 1)
	report = ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("dangling grid reference must be invalid")
	}
	body, _ = json.Marshal(report)
	if !strings.Contains(string(body), "references undefined grid") {
		t.Fatalf("dangling grid reference not reported: %s", body)
	}
}

func TestValidateConfigJSON_BadServiceId(t *testing.T) {
	doc := strings.Replace(fullValidConfig, `"id": "svc1"`, `"id": "../evil"`, 1)
	report := ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("path-like service id must be invalid")
	}
	body, _ := json.Marshal(report)
	if !strings.Contains(string(body), "service id must match") {
		t.Fatalf("bad id not reported: %s", body)
	}

	// the reserved health id must be rejected
	doc = strings.Replace(fullValidConfig, `"id": "svc1"`, `"id": "health"`, 1)
	report = ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("reserved id health must be invalid")
	}
}

func TestValidateConfigJSON_DuplicateIds(t *testing.T) {
	base := strings.TrimSpace(fullValidConfig)
	base = strings.TrimPrefix(base, `{"services": [`)
	base = strings.TrimSuffix(base, `]}`)
	doc := `{"services": [` + base + `, ` + base + `]}`

	report := ValidateConfigJSON([]byte(doc))
	if report.Valid {
		t.Fatal("duplicate ids must be invalid")
	}
	body, _ := json.Marshal(report)
	if !strings.Contains(string(body), "duplicate service id") {
		t.Fatalf("duplicate id not reported: %s", body)
	}
}

func TestValidateConfigJSON_SyntaxError(t *testing.T) {
	report := ValidateConfigJSON([]byte(`{"services": [`))
	if report.Valid {
		t.Fatal("broken JSON must be invalid")
	}
	if len(report.Errors) == 0 {
		t.Fatal("syntax error should be reported at document level")
	}
}

func TestValidateProxyService_Warnings(t *testing.T) {
	report := ValidateConfigJSON([]byte(fullValidConfig))
	if len(report.Services[0].Warnings) != 0 {
		t.Fatalf("valid config should have no warnings: %+v", report.Services[0].Warnings)
	}

	// WMS service without image formats is a warning, not an error
	wms := `{"id": "wms1", "grids": {"g1": {"srs": "EPSG:3857", "bbox": [-1, -1, 1, 1]}}, "service": {"type": "wms", "srs": ["EPSG:3857"]}}`
	report = ValidateConfigJSON([]byte(wms))
	if !report.Valid {
		t.Fatalf("wms without image formats should stay valid: %v", report.Error())
	}
	if len(report.Services[0].Warnings) == 0 {
		t.Fatal("expected a warning for missing image formats")
	}

	// an invalid SRS however is an error
	wms = strings.Replace(wms, `"EPSG:3857"`, "42", 1)
	report = ValidateConfigJSON([]byte(wms))
	if report.Valid {
		t.Fatal("invalid SRS must be invalid")
	}
}

func TestValidateProxyService_RawPayload(t *testing.T) {
	// a raw, unconverted payload (as submitted by a front end) must be
	// validated and converted in one call
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fullValidConfig), &raw); err != nil {
		t.Fatal(err)
	}
	var services []json.RawMessage
	if err := json.Unmarshal(raw["services"], &services); err != nil {
		t.Fatal(err)
	}
	ps := &ProxyService{}
	if err := json.Unmarshal(services[0], ps); err != nil {
		t.Fatal(err)
	}
	rep := ValidateProxyService(ps)
	if !rep.Valid {
		t.Fatalf("raw payload should validate: %v", rep.Error())
	}
	// and it must be converted as a side effect so the proxy can build it
	if _, ok := ps.Service.(*TMSService); !ok {
		t.Fatalf("service not converted: %T", ps.Service)
	}
	if _, ok := ps.Sources["s1"].(*TileSource); !ok {
		t.Fatalf("source not converted: %T", ps.Sources["s1"])
	}
}

func TestValidateConfigJSON_ReportShape(t *testing.T) {
	// the report must be directly serializable for a validation endpoint
	report := ValidateConfigJSON([]byte(`{"services": [{"id": "x"}]}`))
	body, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var back struct {
		Valid    bool `json:"valid"`
		Services []struct {
			Id     string `json:"id"`
			Index  int    `json:"index"`
			Valid  bool   `json:"valid"`
			Errors []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"services"`
	}
	if err := json.Unmarshal(body, &back); err != nil {
		t.Fatalf("report not serializable: %v", err)
	}
	if back.Valid || len(back.Services) != 1 || back.Services[0].Valid {
		t.Fatalf("unexpected report: %s", body)
	}
	if len(back.Services[0].Errors) == 0 {
		t.Fatal("expected errors for empty service")
	}
}
