package request

import (
	"regexp"
	"testing"
)

var (
	TMS_URL               = "http://tms.osgeo.org/1.0.0/landsat2000/1/8500/8500.png"
	TMS_CAPABILITIES_URL  = "http://tms.osgeo.org/1.0.0/landsat2000"
	TMS_ROOT_RESOURCE_URL = "http://tms.osgeo.org/tms/"
)

func TestTileRequest(t *testing.T) {
	reg := regexp.MustCompile(`(?P<begin>[^/]+)/((?P<version>1\.0\.0)/)?(?P<layer>[^/]+)/((?P<layer_spec>[^/]+)/)?(?P<z>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)\.(?P<format>\w+)`)
	match := reg.FindStringSubmatch(TMS_URL)
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

func TestTMSRequest(t *testing.T) {
	reg := regexp.MustCompile(`^.*/1\.0\.0/?(/(?P<layer>[^/]+))?(/(?P<layer_spec>[^/]+))?$`)
	match := reg.FindStringSubmatch(TMS_CAPABILITIES_URL)
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

	root_match := rootRequestRegex.MatchString(TMS_ROOT_RESOURCE_URL)

	if !root_match {
		t.FailNow()
	}
}
