package request

import (
	"regexp"
	"testing"
)

var (
	authUrl = "https://api.cesium.com/v1/assets/1/endpoint?access_token="

	TEST_LAYERJSON    = "/1/layer.json"
	TEST_TERRAIN_TILE = "/1/0/1/0.terrain?extensions=octvertexnormals-metadata&v=1.2.0"
)

func TestCesiumTilejsonRequest(t *testing.T) {
	reg := regexp.MustCompile(`^/(?P<asset_id>[^/]+)/layer.json`)
	match := reg.FindStringSubmatch(TEST_LAYERJSON)
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

func TestCesiumTileRequest(t *testing.T) {
	reg := regexp.MustCompile(`^/(?P<asset_id>[^/]+)/(?P<zoom>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)\.(?P<format>\w+)`)
	match := reg.FindStringSubmatch(TEST_TERRAIN_TILE)
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
