package tileproxy

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/flywave/go-tileproxy/setting"
)

func TestNestedCacheDependencies(t *testing.T) {
	var encoded bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.SetRGBA(x, y, color.RGBA{35, 120, 190, 255})
		}
	}
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(encoded.Bytes())
	}))
	defer origin.Close()
	// Repeated construction exercises different map iteration orders. All caches
	// deliberately omit Name: references are the keys in the caches object.
	for attempt := 0; attempt < 24; attempt++ {
		dir := t.TempDir()
		cfg := buildTestServiceConfig(t, "tiles", filepath.Join(dir, "c1"))
		cfg.Sources["s1"].(*setting.TileSource).URLTemplate = origin.URL + "/{z}/{x}/{y}.png"
		for _, pair := range [][2]string{{"c2", "c1"}, {"c3", "c2"}} {
			child := *cfg.Caches["c1"].(*setting.CacheSource)
			child.Sources = []string{pair[1]}
			child.CacheInfo = &setting.CacheInfo{DirectoryLayout: "tms", Directory: filepath.Join(dir, pair[0])}
			cfg.Caches[pair[0]] = &child
		}
		cfg.Service.(*setting.TMSService).Layers[0].Source = "c3"
		proxy := NewTileProxy(&setting.GlobalsSetting{}, []*setting.ProxyService{cfg}, nil)
		svc := proxy.GetService("tiles")
		for _, name := range []string{"c1", "c2", "c3"} {
			manager := svc.GetCache(name)
			if manager == nil || len(manager.GetSources()) != 1 {
				t.Fatalf("attempt %d: unresolved cache %s", attempt, name)
			}
		}
		response := httptest.NewRecorder()
		proxy.ServeHTTP(response, httptest.NewRequest("GET", "http://localhost/tiles/l1/0/0/0.png", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
		result, err := png.Decode(bytes.NewReader(response.Body.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		red, green, blue, _ := result.At(128, 128).RGBA()
		if red != 35*257 || green != 120*257 || blue != 190*257 {
			t.Fatalf("empty or incorrect nested tile: %d,%d,%d", red, green, blue)
		}
		proxy.RemoveService("tiles")
	}
}
