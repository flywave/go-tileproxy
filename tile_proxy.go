package tileproxy

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/flywave/go-tileproxy/setting"
)

type TileProxy struct {
	m               sync.RWMutex
	Services        map[string]*Service
	globals         *setting.GlobalsSetting
	serviceReqRegex *regexp.Regexp
	serviceCache    map[string]*Service
	serviceCacheMu  sync.RWMutex
	requestTimeout  time.Duration
}

func (t *TileProxy) UpdateService(id string, d *setting.ProxyService, fac setting.CacheFactory) {
	t.m.Lock()
	old, exists := t.Services[id]
	t.Services[id] = NewService(d, t.globals, fac)
	t.m.Unlock()

	if exists {
		old.Clean()
	}

	t.serviceCacheMu.Lock()
	t.serviceCache[id] = t.Services[id]
	t.serviceCacheMu.Unlock()

	log.Printf("service updated: %s", id)
}

func (t *TileProxy) RemoveService(id string) {
	t.m.Lock()
	d, ok := t.Services[id]
	delete(t.Services, id)
	t.m.Unlock()

	t.serviceCacheMu.Lock()
	delete(t.serviceCache, id)
	t.serviceCacheMu.Unlock()

	if ok {
		d.Clean()
		log.Printf("service removed: %s", id)
	}
}

func (t *TileProxy) Reload(proxy []*setting.ProxyService, fac setting.CacheFactory) {
	t.m.Lock()
	oldServices := t.Services
	t.Services = make(map[string]*Service)
	for i := range proxy {
		t.Services[proxy[i].Id] = NewService(proxy[i], t.globals, fac)
	}
	t.m.Unlock()

	for _, s := range oldServices {
		s.Clean()
	}

	t.serviceCacheMu.Lock()
	t.serviceCache = make(map[string]*Service)
	for id, svc := range t.Services {
		t.serviceCache[id] = svc
	}
	t.serviceCacheMu.Unlock()

	log.Printf("reloaded %d services", len(t.Services))
}

func (t *TileProxy) Shutdown(ctx context.Context) error {
	t.m.Lock()
	services := make([]*Service, 0, len(t.Services))
	for _, s := range t.Services {
		services = append(services, s)
	}
	t.m.Unlock()

	done := make(chan struct{}, 1)
	go func() {
		for _, s := range services {
			s.Clean()
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
		log.Println("all services shut down")
		return nil
	case <-ctx.Done():
		log.Println("shutdown timed out")
		return ctx.Err()
	}
}

func (t *TileProxy) GetServiceIDs() []string {
	t.m.RLock()
	defer t.m.RUnlock()
	ids := make([]string, 0, len(t.Services))
	for id := range t.Services {
		ids = append(ids, id)
	}
	return ids
}

func (t *TileProxy) GetService(id string) *Service {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.Services[id]
}

func (t *TileProxy) SetRequestTimeout(timeout time.Duration) {
	t.requestTimeout = timeout
}

func (t *TileProxy) parseServiceId(r *http.Request) string {
	match := t.serviceReqRegex.FindStringSubmatch(r.URL.Path)
	if len(match) == 0 {
		return ""
	}

	groupNames := t.serviceReqRegex.SubexpNames()
	for i, name := range groupNames {
		if name != "" && i < len(match) && match[i] != "" {
			if name == "service" {
				return match[i]
			}
		}
	}

	return ""
}

func (s *TileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic serving %s: %v", r.URL.Path, rec)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	serviceId := s.parseServiceId(r)
	if serviceId == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	start := time.Now()
	defer func() {
		log.Printf("%s %s -> service=%s duration=%v",
			r.Method, r.URL.Path, serviceId, time.Since(start))
	}()

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
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}

	if s.requestTimeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), s.requestTimeout)
		defer cancel()
		r = r.WithContext(ctx)
	}

	d.ServeHTTP(w, r)
}

var serviceReqPattern = regexp.MustCompile(`^/(?P<service>[^/]+)`)

func NewTileProxy(globals *setting.GlobalsSetting, proxys []*setting.ProxyService, fac setting.CacheFactory) *TileProxy {
	proxy := &TileProxy{
		globals:         globals,
		serviceCache:    make(map[string]*Service),
		serviceReqRegex: serviceReqPattern,
	}
	proxy.Reload(proxys, fac)
	return proxy
}
