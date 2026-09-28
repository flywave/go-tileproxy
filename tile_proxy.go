package tileproxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/flywave/go-tileproxy/setting"
)

var serviceReqPattern = regexp.MustCompile(`^/(?P<service>[^/]+)`)

// drainTimeout bounds how long a retired service instance waits for
// in-flight requests to finish before its resources are released.
const drainTimeout = 30 * time.Second

// defaultSlowRequestThreshold is the duration above which a request is
// logged; logging every request serializes hot paths on the log writer.
const defaultSlowRequestThreshold = 200 * time.Millisecond

type TileProxy struct {
	m               sync.RWMutex
	Services        map[string]*Service
	globals         *setting.GlobalsSetting
	serviceReqRegex *regexp.Regexp
	requestTimeout  time.Duration
	slowLog         time.Duration
}

// UpdateService builds and installs a single service, replacing an existing
// one with the same id. The running configuration is left untouched when the
// new service cannot be built.
func (t *TileProxy) UpdateService(id string, d *setting.ProxyService, fac setting.CacheFactory) error {
	svc, err := buildService(d, t.globals, fac)
	if err != nil {
		return fmt.Errorf("service %s: %w", id, err)
	}

	t.m.Lock()
	old, existed := t.Services[id]
	t.Services[id] = svc
	t.m.Unlock()

	if existed {
		go retire(id, old)
	}

	log.Printf("service updated: %s", id)
	return nil
}

func (t *TileProxy) RemoveService(id string) {
	t.m.Lock()
	d, ok := t.Services[id]
	delete(t.Services, id)
	t.m.Unlock()

	if ok {
		go retire(id, d)
		log.Printf("service removed: %s", id)
	}
}

// ReloadServices atomically replaces the whole set of services. The new
// configuration is validated and built first; if anything fails the current
// services keep serving and the problems are returned. Retired services are
// drained and cleaned up in the background so in-flight requests finish
// safely.
func (t *TileProxy) ReloadServices(services []*setting.ProxyService, fac setting.CacheFactory) error {
	if len(services) == 0 {
		return errors.New("empty service configuration")
	}

	built := make(map[string]*Service, len(services))
	var errs []error
	for _, ps := range services {
		if ps == nil {
			errs = append(errs, errors.New("service definition is null"))
			continue
		}
		if _, dup := built[ps.Id]; dup {
			errs = append(errs, fmt.Errorf("service %q: duplicate id", ps.Id))
			continue
		}
		svc, err := buildService(ps, t.globals, fac)
		if err != nil {
			errs = append(errs, fmt.Errorf("service %q: %w", ps.Id, err))
			continue
		}
		built[ps.Id] = svc
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("configuration rejected, keeping current services: %w", err)
	}

	t.m.Lock()
	old := t.Services
	t.Services = built
	t.m.Unlock()

	for id, s := range old {
		go retire(id, s)
	}

	ids := t.GetServiceIDs()
	log.Printf("reloaded %d services: %v", len(ids), ids)
	return nil
}

// Reload replaces all services, logging instead of returning failures. It is
// kept for backwards compatibility; prefer ReloadServices.
func (t *TileProxy) Reload(proxy []*setting.ProxyService, fac setting.CacheFactory) {
	if err := t.ReloadServices(proxy, fac); err != nil {
		log.Printf("reload failed: %v", err)
	}
}

// ReloadFromFile loads a configuration file and applies it without stopping
// the server. When the file is missing, malformed or fails validation the
// currently running services are left untouched and the error is returned.
func (t *TileProxy) ReloadFromFile(path string, fac setting.CacheFactory) error {
	services, err := setting.LoadProxyServicesFromFile(path)
	if err != nil {
		return fmt.Errorf("reload from %s: %w", path, err)
	}
	return t.ReloadServices(services, fac)
}

func buildService(d *setting.ProxyService, globals *setting.GlobalsSetting, fac setting.CacheFactory) (*Service, error) {
	if d == nil {
		return nil, errors.New("service definition is nil")
	}
	if d.Id == "" {
		return nil, errors.New("service id is required")
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	svc := NewService(d, globals, fac)
	if svc.Service == nil {
		if svc.LoadError() != nil {
			return nil, svc.LoadError()
		}
		return nil, errors.New("no supported service type configured")
	}
	return svc, nil
}

// retire drains in-flight requests of a replaced service instance and then
// releases its resources. It runs in its own goroutine so reloads never
// block the caller.
func retire(id string, s *Service) {
	s.Clean()
	log.Printf("service %s: old instance retired", id)
}

func (t *TileProxy) Shutdown(ctx context.Context) error {
	t.m.Lock()
	services := make([]*Service, 0, len(t.Services))
	for _, s := range t.Services {
		services = append(services, s)
	}
	t.Services = make(map[string]*Service)
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

// SetSlowLogThreshold sets the duration above which requests are logged.
// A negative value disables request logging entirely; zero restores the
// default of 200ms.
func (t *TileProxy) SetSlowLogThreshold(d time.Duration) {
	if d == 0 {
		d = defaultSlowRequestThreshold
	}
	t.slowLog = d
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
		// logging every request serializes concurrent handlers on the log
		// writer, so only slow requests (and the configured threshold)
		// produce output
		threshold := s.slowLog
		if threshold == 0 {
			threshold = defaultSlowRequestThreshold
		}
		if elapsed := time.Since(start); threshold >= 0 && elapsed >= threshold {
			log.Printf("slow request: %s %s -> service=%s duration=%v",
				r.Method, r.URL.Path, serviceId, elapsed)
		}
	}()

	s.m.RLock()
	d, ok := s.Services[serviceId]
	s.m.RUnlock()

	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if s.requestTimeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), s.requestTimeout)
		defer cancel()
		r = r.WithContext(ctx)
	}

	d.ServeHTTP(w, r)
}

func NewTileProxy(globals *setting.GlobalsSetting, proxys []*setting.ProxyService, fac setting.CacheFactory) *TileProxy {
	proxy := &TileProxy{
		globals:         globals,
		Services:        make(map[string]*Service),
		serviceReqRegex: serviceReqPattern,
		slowLog:         defaultSlowRequestThreshold,
	}
	if err := proxy.ReloadServices(proxys, fac); err != nil {
		log.Printf("initial service load incomplete: %v", err)
	}
	return proxy
}
