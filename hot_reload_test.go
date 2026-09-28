package tileproxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/setting"
)

// testUpstreamURL returns an upstream URL that fails instantly with
// connection refused, so cache misses never block the tests on network IO.
func testUpstreamURL(t *testing.T) string {
	t.Helper()
	return "http://127.0.0.1:1"
}

// buildTestServiceConfig creates a minimal, buildable TMS service backed by
// a local file cache in dir.
func buildTestServiceConfig(t *testing.T, id, dir string) *setting.ProxyService {
	ps := setting.NewProxyService(id)
	ps.Grids["g1"] = setting.GridOpts{
		Srs:      "EPSG:3857",
		BBox:     &[4]float64{-20037508.34, -20037508.34, 20037508.34, 20037508.34},
		TileSize: &[2]uint32{256, 256},
	}
	reqTimeout := 2 * time.Second
	threads := 16
	ps.Sources["s1"] = &setting.TileSource{
		SourceCommons: setting.SourceCommons{
			Http: &setting.HttpSetting{
				RequestTimeout: &reqTimeout,
				Threads:        &threads,
			},
		},
		URLTemplate:   testUpstreamURL(t) + "/{z}/{x}/{y}.png",
		Grid:          "g1",
		RequestFormat: "png",
		Options:       &setting.ImageOpts{Format: "png"},
	}
	ps.Caches["c1"] = &setting.CacheSource{
		Grid:          "g1",
		Sources:       []string{"s1"},
		Format:        "png",
		RequestFormat: "png",
		CacheInfo:     &setting.CacheInfo{DirectoryLayout: "tms", Directory: dir},
		TileOptions:   &setting.ImageOpts{Format: "png"},
	}
	ps.Service = &setting.TMSService{
		Title:  id,
		Origin: "nw",
		Layers: []setting.TileLayer{{Name: "l1", Title: "Layer 1", Source: "c1"}},
	}
	return ps
}

func buildBrokenServiceConfig(id string) *setting.ProxyService {
	ps := setting.NewProxyService(id)
	// service member of an unknown type can never be built
	ps.Service = map[string]interface{}{"type": "does-not-exist"}
	ps.Grids = map[string]setting.GridOpts{}
	ps.Sources = map[string]interface{}{}
	ps.Caches = map[string]interface{}{}
	return ps
}

// warmTileCaches pre-stores an empty tile at coordinate 0/0/0 in every
// service cache so that test requests are served from cache and never
// trigger upstream tile fetches (which block on network IO).
func warmTileCaches(t *testing.T, proxy *TileProxy) {
	t.Helper()
	for _, id := range proxy.GetServiceIDs() {
		svc := proxy.GetService(id)
		if svc == nil {
			continue
		}
		if mng := svc.GetCache("c1"); mng != nil {
			src := cache.GetEmptyTile([2]uint32{256, 256}, mng.GetTileOptions())
			if src == nil {
				t.Errorf("no tile options for cache of %s", id)
				return
			}
			tile := cache.NewTile([3]int{0, 0, 0})
			tile.Source = src
			if err := mng.StoreTile(tile); err != nil {
				t.Errorf("failed to warm cache of %s: %v", id, err)
				return
			}
		}
	}
}

func newTestProxy(t *testing.T, id string) (*TileProxy, string) {
	t.Helper()
	dir := t.TempDir()
	proxy := NewTileProxy(&setting.GlobalsSetting{}, []*setting.ProxyService{buildTestServiceConfig(t, id, filepath.Join(dir, "cache"))}, nil)
	if got := proxy.GetServiceIDs(); len(got) != 1 || got[0] != id {
		t.Fatalf("proxy did not load test service: %v", got)
	}
	warmTileCaches(t, proxy)
	return proxy, dir
}

func requestStatus(t *testing.T, proxy *TileProxy, path string) int {
	t.Helper()
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w.Code
}

func TestReloadServices_RejectsBrokenConfigAndKeepsServing(t *testing.T) {
	proxy, dir := newTestProxy(t, "svc1")

	err := proxy.ReloadServices([]*setting.ProxyService{
		buildTestServiceConfig(t, "svc2", filepath.Join(dir, "c2")),
		buildBrokenServiceConfig("svc3"),
	}, nil)
	if err == nil {
		t.Fatal("broken configuration must be rejected")
	}

	// the whole broken snapshot is rejected atomically: svc2 must NOT be
	// installed and svc1 must keep serving
	ids := proxy.GetServiceIDs()
	if len(ids) != 1 || ids[0] != "svc1" {
		t.Fatalf("services changed after rejected reload: %v", ids)
	}
	if code := requestStatus(t, proxy, "/svc1/tms/1.0.0/l1/0/0/0.png"); code != 200 {
		t.Fatalf("svc1 broken after rejected reload, status %d", code)
	}
}

func TestReloadServices_AtomicReplaceAndRetire(t *testing.T) {
	proxy, dir := newTestProxy(t, "svc1")
	old := proxy.GetService("svc1")
	if old == nil {
		t.Fatal("svc1 missing")
	}

	if err := proxy.ReloadServices([]*setting.ProxyService{
		buildTestServiceConfig(t, "svc2", filepath.Join(dir, "c2")),
	}, nil); err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	warmTileCaches(t, proxy)

	ids := proxy.GetServiceIDs()
	if len(ids) != 1 || ids[0] != "svc2" {
		t.Fatalf("expected only svc2 after reload, got %v", ids)
	}
	if code := requestStatus(t, proxy, "/svc2/tms/1.0.0/l1/0/0/0.png"); code != 200 {
		t.Fatalf("svc2 not serving after reload, status %d", code)
	}

	// the retired instance must eventually reach the stopped state
	deadline := time.Now().Add(5 * time.Second)
	for old.Status() != ServiceStatusStopped && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if old.Status() != ServiceStatusStopped {
		t.Fatalf("retired service still in state %v", old.Status())
	}
}

func TestUpdateService_FailureKeepsOldInstance(t *testing.T) {
	proxy, _ := newTestProxy(t, "svc1")

	if err := proxy.UpdateService("svc1", buildBrokenServiceConfig("svc1"), nil); err == nil {
		t.Fatal("broken update must fail")
	}
	if proxy.GetService("svc1") == nil {
		t.Fatal("svc1 was dropped by a failed update")
	}
	if code := requestStatus(t, proxy, "/svc1/tms/1.0.0/l1/0/0/0.png"); code != 200 {
		t.Fatalf("svc1 broken after failed update, status %d", code)
	}

	// unknown id with a broken config must not install anything
	if err := proxy.UpdateService("svc9", buildBrokenServiceConfig("svc9"), nil); err == nil {
		t.Fatal("broken update must fail")
	}
	if proxy.GetService("svc9") != nil {
		t.Fatal("svc9 was installed despite build failure")
	}
}

func TestReloadFromFile_ReloadsChangedConfig(t *testing.T) {
	proxy, dir := newTestProxy(t, "svc1")

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, buildMultiConfig(t, []string{"svc1", "svc2"}, dir), 0644); err != nil {
		t.Fatal(err)
	}
	if err := proxy.ReloadFromFile(path, nil); err != nil {
		t.Fatalf("reload from file failed: %v", err)
	}
	warmTileCaches(t, proxy)

	ids := proxy.GetServiceIDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 services after reload, got %v", ids)
	}

	// a broken file keeps the current services and reports the problem
	if err := os.WriteFile(path, []byte(`{"services": [{"id": "bad"}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := proxy.ReloadFromFile(path, nil); err == nil {
		t.Fatal("broken file must be rejected")
	}
	if got := len(proxy.GetServiceIDs()); got != 2 {
		t.Fatalf("services lost after rejected reload: %d", got)
	}
}

func TestConcurrentRequestsAndReloads(t *testing.T) {
	proxy, dir := newTestProxy(t, "svc1")

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// concurrent reloaders flipping between configurations
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				select {
				case <-stop:
					return
				default:
				}
				var err error
				if j%2 == 0 {
					err = proxy.ReloadServices([]*setting.ProxyService{
						buildTestServiceConfig(t, fmt.Sprintf("svc%d", i+1), filepath.Join(dir, fmt.Sprintf("c%d", i+1))),
					}, nil)
				} else {
					err = proxy.ReloadServices([]*setting.ProxyService{
						buildTestServiceConfig(t, "svc1", filepath.Join(dir, "c1")),
						buildTestServiceConfig(t, "svc2", filepath.Join(dir, "c2")),
					}, nil)
				}
				if err == nil {
					warmTileCaches(t, proxy)
				}
				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	// concurrent requesters hitting existing and unknown services
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			paths := []string{
				"/svc1/tms/1.0.0/l1/0/0/0.png",
				"/svc2/tms/1.0.0/l1/0/0/0.png",
				"/unknown/tms/1.0.0/l1/0/0/0.png",
			}
			for j := 0; j < 100; j++ {
				w := httptest.NewRecorder()
				proxy.ServeHTTP(w, httptest.NewRequest("GET", paths[j%len(paths)], nil))
				if w.Code == http.StatusInternalServerError {
					t.Errorf("unexpected 500 for %s", paths[j%len(paths)])
					return
				}
			}
		}(i)
	}

	// wait a bit then stop reloaders first so requests can finish cleanly
	time.Sleep(500 * time.Millisecond)
	close(stop)
	wg.Wait()
}

// fakeService blocks inside ServeHTTP until released, to observe the
// drain behaviour of Clean.
type fakeService struct {
	started chan struct{}
	release chan struct{}
}

func (f *fakeService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	close(f.started)
	<-f.release
	w.WriteHeader(http.StatusOK)
}

func (f *fakeService) RequestParser(r *http.Request) request.Request { return nil }

func TestServiceClean_DrainsInFlightRequests(t *testing.T) {
	fake := &fakeService{started: make(chan struct{}), release: make(chan struct{})}
	svc := &Service{Id: "fake", Service: fake}

	req := httptest.NewRequest("GET", "/fake/tile", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		svc.ServeHTTP(w, req)
		close(done)
	}()

	<-fake.started

	cleaned := make(chan struct{})
	go func() {
		svc.Clean()
		close(cleaned)
	}()

	// Clean must not finish while the request is still in flight
	select {
	case <-cleaned:
		t.Fatal("Clean finished while a request was still in flight")
	case <-time.After(100 * time.Millisecond):
	}

	// once the request completes, Clean proceeds
	close(fake.release)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("request never completed")
	}
	select {
	case <-cleaned:
	case <-time.After(2 * time.Second):
		t.Fatal("Clean never completed after request finished")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("in-flight request got status %d", w.Code)
	}
}

func TestConfigWatcher_HotReloadsOnChange(t *testing.T) {
	proxy, dir := newTestProxy(t, "svc1")

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, buildMultiConfig(t, []string{"svc1"}, dir), 0644); err != nil {
		t.Fatal(err)
	}

	watcher := NewConfigWatcher(proxy, path, nil, 20*time.Millisecond, 60*time.Millisecond)
	watcher.Start()
	defer watcher.Stop()

	if err := os.WriteFile(path, buildMultiConfig(t, []string{"svc1", "svc2"}, dir), 0644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		ids := proxy.GetServiceIDs()
		if len(ids) == 2 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("watcher did not hot-reload, services: %v", proxy.GetServiceIDs())
}

func buildMultiConfig(t *testing.T, ids []string, dir string) []byte {
	out := `{"services": [`
	for i, id := range ids {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf(`{"id": %q, "grids": {"g1": {"srs": "EPSG:3857", "bbox": [-20037508.34, -20037508.34, 20037508.34, 20037508.34], "tile_size": [256, 256]}}, `+
			`"sources": {"s1": {"type": "tile", "url_template": %q, "grid": "g1", "request_format": "png", "options": {"type": "image", "format": "png"}, "http": {"request_timeout": 2, "thread_size": 16}}}, `+
			`"caches": {"c1": {"grid": "g1", "sources": ["s1"], "format": "png", "request_format": "png", "cache": {"directory_layout": "tms", "directory": %q}, "tile_options": {"type": "image", "format": "png"}}}, `+
			`"service": {"type": "tms", "title": %q, "layers": [{"name": "l1", "title": "L1", "source": "c1"}]}}`,
			id, testUpstreamURL(t)+"/{z}/{x}/{y}.png", filepath.ToSlash(filepath.Join(dir, "cache-"+id)), id)
	}
	out += `]}`
	return []byte(out)
}
