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
	globals         *setting.GlobalsSetting
	serviceReqRegex *regexp.Regexp
	serviceCache    map[string]*Service
	serviceCacheMu  sync.RWMutex
}

func (t *TileProxy) UpdateService(id string, d *setting.ProxyService, fac setting.CacheFactory) {
	t.m.Lock()
	t.Services[id] = NewService(d, t.globals, fac)
	t.m.Unlock()

	t.serviceCacheMu.Lock()
	t.serviceCache[id] = t.Services[id]
	t.serviceCacheMu.Unlock()
}

func (t *TileProxy) RemoveService(id string) {
	var d *Service
	t.m.Lock()
	if d_, ok := t.Services[id]; ok {
		d = d_
	}
	delete(t.Services, id)
	t.m.Unlock()

	t.serviceCacheMu.Lock()
	delete(t.serviceCache, id)
	t.serviceCacheMu.Unlock()

	d.Clean()
}

func (t *TileProxy) Reload(proxy []*setting.ProxyService, fac setting.CacheFactory) {
	t.m.Lock()
	t.Services = make(map[string]*Service)
	for i := range proxy {
		t.Services[proxy[i].Id] = NewService(proxy[i], t.globals, fac)
	}
	t.m.Unlock()

	t.serviceCacheMu.Lock()
	t.serviceCache = make(map[string]*Service)
	t.serviceCacheMu.Unlock()
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

	s.serviceCacheMu.RLock()
	d, ok := s.serviceCache[serviceId]
	s.serviceCacheMu.RUnlock()

	if !ok {
		s.m.RLock()
		d, ok = s.Services[serviceId]
		s.m.RUnlock()

		if ok {
			s.serviceCacheMu.Lock()
			s.serviceCache[serviceId] = d
			s.serviceCacheMu.Unlock()
		} else {
			w.WriteHeader(404)
			return
		}
	}

	d.ServeHTTP(w, r)
}

func NewTileProxy(globals *setting.GlobalsSetting, proxys []*setting.ProxyService, fac setting.CacheFactory) *TileProxy {
	proxy := &TileProxy{globals: globals, serviceCache: make(map[string]*Service)}
	proxy.Reload(proxys, fac)
	return proxy
}
