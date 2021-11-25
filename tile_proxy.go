package tileproxy

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/flywave/go-tileproxy/setting"
)

type TileProxy struct {
	m               sync.RWMutex
	Services        map[string]*Service
	basePath        string
	globals         *setting.GlobalsSetting
	serviceReqRegex *regexp.Regexp
}

func (t *TileProxy) UpdateService(uuid string, d *setting.ProxyService, fac setting.CacheFactory) {
	t.m.Lock()
	t.Services[uuid] = NewService(d, t.basePath, t.globals, fac)
	t.m.Unlock()
}

func (t *TileProxy) RemoveService(uuid string) {
	var d *Service
	t.m.Lock()
	if d_, ok := t.Services[uuid]; ok {
		d = d_
	}
	delete(t.Services, uuid)
	t.m.Unlock()
	d.Clean()
}

func (t *TileProxy) Reload(proxy []*setting.ProxyService, fac setting.CacheFactory) {
	t.m.Lock()
	t.Services = make(map[string]*Service)
	for i := range proxy {
		t.Services[proxy[i].UUID] = NewService(proxy[i], t.basePath, t.globals, fac)
	}
	t.m.Unlock()
}

func (t *TileProxy) parseServiceId(r *http.Request) string {
	match := t.serviceReqRegex.FindStringSubmatch(r.URL.Path)
	groupNames := t.serviceReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(match) == 0 {
		return ""
	}

	return result["service"]
}

func (s *TileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceId := s.parseServiceId(r)
	if serviceId == "" {
		w.WriteHeader(404)
		return
	}
	s.m.Lock()
	if d, ok := s.Services[serviceId]; ok {
		s.m.Unlock()
		d.ServeHTTP(w, r)
	} else {
		s.m.Unlock()
		w.WriteHeader(404)
	}
}

func NewTileProxy(basePath string, globals *setting.GlobalsSetting, proxys []*setting.ProxyService, fac setting.CacheFactory) *TileProxy {
	proxy := &TileProxy{basePath: basePath, globals: globals}
	proxy.Reload(proxys, fac)
	return proxy
}
