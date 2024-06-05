package request

import (
	"net/url"
	"regexp"
	"testing"
)

var (
	TEST_TILEJSON         = "https://api.mapbox.com/v4/examples.civir98a801cq2oo6w6mk1aor-9msik.json?access_token={token}"
	TEST_VECTOR_TILE      = "https://api.mapbox.com/v4/mapbox.mapbox-streets-v8/1/0/0.mvt"
	TEST_RASTER_TILE      = "https://api.mapbox.com/v4/mapbox.satellite/1/0/0@2x.jpg90"
	TEST_STYLE_URL        = "https://api.mapbox.com/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7"
	TEST_SPRITE_JSON_URL  = "https://api.mapbox.com/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7/sprite@3x"
	TEST_SPRITE_IMAGE_URL = "https://api.mapbox.com/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7/sprite.png"
	TEST_GLYPHS_URL       = "https://api.mapbox.com/fonts/v1/examples/Arial%20Unicode%20MS%20Regular/0-255.pbf"
)

func TestMapboxTilejsonRequest(t *testing.T) {
	reg := regexp.MustCompile(`/(?P<tileset_name>[^/]+)/(?P<file_name>[^/]+).json`)
	match := reg.FindStringSubmatch("http://flywave.com/api/v1/services/sxsxsx/layers/xxx/tilestats.json")
	groupNames := reg.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(result) == 0 {
		t.FailNow()
	}
}

func TestMapboxTileRequest(t *testing.T) {
	reg := regexp.MustCompile(`(?P<begin>[^/]+)/(?P<version>[^/]+)/(?P<tileset_id>[^/]+)/(?P<zoom>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)(@(?P<retina>[^/]+)x)?\.(?P<format>\w+)`)
	match := reg.FindStringSubmatch(TEST_RASTER_TILE)
	groupNames := reg.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(result) == 0 {
		t.FailNow()
	}
}

func TestMapboxStyleRequest(t *testing.T) {
	reg := regexp.MustCompile(`(?P<begin>[^/]+)/styles/(?P<version>[^/]+)/(?P<username>[^/]+)/(?P<style_id>[^/]+)`)
	match := reg.FindStringSubmatch(TEST_STYLE_URL)
	groupNames := reg.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(result) == 0 {
		t.FailNow()
	}
}

func TestMapboxSpriteRequest(t *testing.T) {
	reg := regexp.MustCompile(`(?P<begin>[^/]+)/styles/(?P<version>[^/]+)/(?P<username>[^/]+)/(?P<style_id>[^/]+)/sprite(@(?P<retina>[^/]+)x)?\.?(?P<format>\w+)?`)
	match := reg.FindStringSubmatch(TEST_SPRITE_IMAGE_URL)
	groupNames := reg.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(result) == 0 {
		t.FailNow()
	}
}

func TestMapboxGlyphsRequest(t *testing.T) {
	reg := regexp.MustCompile(`(?P<begin>[^/]+)/fonts/(?P<version>[^/]+)/(?P<username>[^/]+)/(?P<font>[^/]+)/(?P<start>-?\d+)-(?P<end>-?\d+)\.(?P<format>\w+)`)
	match := reg.FindStringSubmatch(TEST_GLYPHS_URL)
	groupNames := reg.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			if name == "font" {
				str, _ := url.PathUnescape(match[i])
				result[name] = str
			} else {
				result[name] = match[i]
			}
		}
	}

	if len(result) == 0 {
		t.FailNow()
	}
}
