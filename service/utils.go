package service

import (
	"net/http"
	"strings"

	"github.com/flywave/go-tileproxy/utils"
)

// servicePathPrefix returns the first path segment of a mounted service,
// e.g. "/wms" for "/wms/service.cgi?request=GetCapabilities".
func servicePathPrefix(path string) string {
	if path == "" || path[0] != '/' {
		return ""
	}
	if i := strings.Index(path[1:], "/"); i >= 0 {
		return path[:i+1]
	}
	return path
}

// serviceBaseURL builds the externally visible base URL of a service,
// including scheme, host and mount path.
func serviceBaseURL(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "http://localhost"
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	if host == "" {
		host = "localhost"
	}
	return scheme + "://" + host + servicePathPrefix(r.URL.Path)
}

func WMS100Format(format string) string {
	sub_type := strings.Split(format, "/")[1]
	sub_type = strings.ToUpper(sub_type)
	if utils.ContainsString([]string{"PNG", "TIFF", "GIF", "JPEG"}, sub_type) {
		return sub_type
	} else {
		return ""
	}
}

func WMS100InfoFormat(format string) string {
	if format == "application/vnd.ogc.gml" || format == "text/xml" {
		return "GML.1"
	}
	return "MIME"
}

func WMS111MetaDataType(type_ string) string {
	if type_ == "ISO19115:2003" {
		return "TC211"
	}
	if type_ == "FGDC:1998" {
		return "FGDC"
	}
	return ""
}
