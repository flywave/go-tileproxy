package request

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

var (
	TILE_URL               = "http://tms.osgeo.org/tiles/1.0.0/landsat2000/1/8500/8500.png"
	TILE_CAPABILITIES_URL  = "http://tms.osgeo.org/1.0.0/landsat2000"
	TILE_ROOT_RESOURCE_URL = "http://tms.osgeo.org/tms/"

	TEST_URL = "/tiles/tms_layer/5/19/13.png"
)

func TestOSMTileRequestParsing(t *testing.T) {
	reg := regexp.MustCompile(`(?P<begin>/[^/]+)/((?P<version>1\.0\.0)/)?(?P<layer>[^/]+)/((?P<layer_spec>[^/]+)/)?(?P<z>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)\.(?P<format>\w+)`)
	match := reg.FindStringSubmatch(TEST_URL)
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

func TestTileRequestParsing(t *testing.T) {
	reg := regexp.MustCompile(`(?P<begin>/[^/]+)/((?P<version>1\.0\.0)/)?(?P<layer>[^/]+)/((?P<layer_spec>[^/]+)/)?(?P<z>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)\.(?P<format>\w+)`)
	match := reg.FindStringSubmatch(TILE_URL)
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

func TestTMSRequestParsing(t *testing.T) {
	reg := regexp.MustCompile(`^.*/1\.0\.0/?(/(?P<layer>[^/]+))?(/(?P<layer_spec>[^/]+))?$`)
	match := reg.FindStringSubmatch(TILE_CAPABILITIES_URL)
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

	rootRequestRegex := regexp.MustCompile(`/tms/?$`)

	root_match := rootRequestRegex.MatchString(TILE_ROOT_RESOURCE_URL)

	if !root_match {
		t.FailNow()
	}
}

func TestTileRequest(t *testing.T) {
	req := &http.Request{}
	req.URL, _ = url.Parse(TILE_URL)

	tileReq := &TileRequest{Http: req}
	tileReq.init()

	if tileReq.Tile[0] != 8500 {
		t.FailNow()
	}
}

func TestTMSRequest(t *testing.T) {
	req := &http.Request{}
	req.URL, _ = url.Parse(TILE_CAPABILITIES_URL)

	tileReq := &TMSRequest{TileRequest: TileRequest{Http: req}}
	tileReq.init()

	if tileReq.RequestHandlerName != "tms_capabilities" {
		t.FailNow()
	}

	req = &http.Request{}
	req.URL, _ = url.Parse(TILE_ROOT_RESOURCE_URL)

	tileReq = &TMSRequest{TileRequest: TileRequest{Http: req}}
	tileReq.init()

	if tileReq.RequestHandlerName != "tms_root_resource" {
		t.FailNow()
	}
}
