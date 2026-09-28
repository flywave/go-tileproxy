package setting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const MaxJSONSize = 10 * 1024 * 1024 // 10MB

type ProxyService struct {
	Id        string                 `json:"id,omitempty"`
	Service   interface{}            `json:"service,omitempty"`
	Coverages map[string]Coverage    `json:"coverages,omitempty"`
	Grids     map[string]GridOpts    `json:"grids,omitempty"`
	Sources   map[string]interface{} `json:"sources,omitempty"`
	Caches    map[string]interface{} `json:"caches,omitempty"`
}

func NewProxyService(id string) *ProxyService {
	return &ProxyService{
		Id:        id,
		Coverages: make(map[string]Coverage),
		Sources:   make(map[string]interface{}),
		Caches:    make(map[string]interface{}),
		Grids:     make(map[string]GridOpts),
	}
}

func CreateProxyServiceFromJSON(content []byte) (*ProxyService, error) {
	if len(content) > MaxJSONSize {
		return nil, errors.New("JSON size exceeds maximum limit")
	}
	set := &ProxyService{}
	reader := bytes.NewBuffer(content)
	dec := json.NewDecoder(reader)
	dec.DisallowUnknownFields()
	if err := dec.Decode(set); err != nil {
		return nil, err
	}
	set.Covert()
	return set, nil
}

// Covert converts the loosely typed JSON members (service, sources, caches)
// into their concrete configuration structs so that the proxy can consume
// them.
func (ps *ProxyService) Covert() {
	ps.Service = CovertService(ps.Service)
	ps.Sources = CovertSources(ps.Sources)
	ps.Caches = CovertCaches(ps.Caches)
}

// CovertService converts the raw service member into a typed service
// configuration based on its "type" field.
func CovertService(ser interface{}) interface{} {
	s, ok := ser.(map[string]interface{})
	if !ok {
		return ser
	}
	ty, ok := s["type"].(string)
	if !ok {
		return ser
	}
	data, err := json.Marshal(ser)
	if err != nil {
		return ser
	}
	switch ty {
	case string(CESIUM_SERVICE):
		sv := &CesiumService{}
		if err := json.Unmarshal(data, sv); err != nil {
			return ser
		}
		return sv
	case string(MAPBOX_SERVICE):
		sv := &MapboxService{}
		if err := json.Unmarshal(data, sv); err != nil {
			return ser
		}
		return sv
	case string(WMS_SERVICE):
		sv := &WMSService{}
		if err := json.Unmarshal(data, sv); err != nil {
			return ser
		}
		return sv
	case string(WMTS_SERVICE):
		sv := &WMTSService{}
		if err := json.Unmarshal(data, sv); err != nil {
			return ser
		}
		return sv
	case string(TMS_SERVICE):
		sv := &TMSService{}
		if err := json.Unmarshal(data, sv); err != nil {
			return ser
		}
		return sv
	}
	return ser
}

// CovertSources converts each raw source member into a typed source
// configuration based on its "type" field. Members with an unknown type are
// left untouched so that validation can report them.
func CovertSources(sources map[string]interface{}) map[string]interface{} {
	if sources == nil {
		return nil
	}
	for name, src := range sources {
		mp, ok := src.(map[string]interface{})
		if !ok {
			continue
		}
		ty, _ := mp["type"].(string)
		data, err := json.Marshal(src)
		if err != nil {
			continue
		}
		switch SourceType(ty) {
		case WMS_SOURCE:
			sv := &WMSSource{}
			if err := json.Unmarshal(data, sv); err == nil {
				sources[name] = sv
			}
		case TILE_SOURCE:
			sv := &TileSource{}
			if err := sv.FromJson(data); err == nil {
				sources[name] = sv
			}
		case MAPBOXTILE_SOURCE:
			sv := &MapboxTileSource{}
			if err := sv.FromJson(data); err == nil {
				sources[name] = sv
			}
		case CESIUMTILE_SOURCE:
			sv := &CesiumTileSource{}
			if err := sv.FromJson(data); err == nil {
				sources[name] = sv
			}
		case ARCGIS_SOURCE:
			sv := &ArcGISSource{}
			if err := json.Unmarshal(data, sv); err == nil {
				sources[name] = sv
			}
		}
	}
	return sources
}

// CovertCaches converts each raw cache member into a typed CacheSource.
func CovertCaches(caches map[string]interface{}) map[string]interface{} {
	if caches == nil {
		return nil
	}
	for name, c := range caches {
		data, err := json.Marshal(c)
		if err != nil {
			continue
		}
		sv := &CacheSource{}
		if err := sv.FromJson(data); err == nil {
			caches[name] = sv
		}
	}
	return caches
}

// LoadProxyServicesFromFile reads and parses a proxy configuration file.
func LoadProxyServicesFromFile(path string) ([]*ProxyService, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}
	if info.Size() > MaxJSONSize {
		return nil, fmt.Errorf("config file too large (max %d bytes)", MaxJSONSize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return LoadProxyServicesFromJSON(data)
}

// LoadProxyServicesFromJSON parses a proxy configuration document. It
// accepts either a single service object or a {"services": [...]} document.
// Every service is converted to its typed form and validated; the returned
// error aggregates all problems found so that the caller can keep the
// running configuration when the new one is broken.
func LoadProxyServicesFromJSON(data []byte) ([]*ProxyService, error) {
	if len(data) > MaxJSONSize {
		return nil, errors.New("config too large (max 10MB)")
	}
	if !isJSONObject(data) {
		return nil, errors.New("invalid config: expected a JSON object")
	}

	var multi multiServiceConfig
	if err := json.Unmarshal(data, &multi); err == nil && multi.Services != nil {
		return convertAndValidateServices(multi.Services)
	}

	var single ProxyService
	if err := json.Unmarshal(data, &single); err != nil {
		return nil, fmt.Errorf("invalid config: expected {\"services\":[...]} or a single service object: %w", err)
	}
	services, err := convertAndValidateServices([]*ProxyService{&single})
	if err != nil {
		return nil, err
	}
	return services, nil
}

type multiServiceConfig struct {
	Services []*ProxyService `json:"services"`
}

func convertAndValidateServices(raw []*ProxyService) ([]*ProxyService, error) {
	if len(raw) == 0 {
		return nil, errors.New("invalid config: no services defined")
	}
	services := make([]*ProxyService, 0, len(raw))
	var errs []error
	seen := make(map[string]bool, len(raw))
	for i, ps := range raw {
		if ps == nil {
			errs = append(errs, fmt.Errorf("service #%d is null", i))
			continue
		}
		ps.Covert()
		if err := ps.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("service %q: %w", ps.Id, err))
			continue
		}
		if seen[ps.Id] {
			errs = append(errs, fmt.Errorf("service %q: duplicate id", ps.Id))
			continue
		}
		seen[ps.Id] = true
		services = append(services, ps)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return services, nil
}

func isJSONObject(data []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(data, &js) == nil
}

func (gs *ProxyService) ToJSON() []byte {
	var bt []byte
	wr := bytes.NewBuffer(bt)
	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	enc.Encode(gs)
	return wr.Bytes()
}
