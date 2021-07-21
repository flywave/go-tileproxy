package service

import (
	"strings"

	"github.com/flywave/go-tileproxy/utils"
)

func wms100Format(format string) string {
	sub_type := strings.Split(format, "/")[1]
	sub_type = strings.ToUpper(sub_type)
	if utils.ContainsString([]string{"PNG", "TIFF", "GIF", "JPEG"}, sub_type) {
		return sub_type
	} else {
		return ""
	}
}

func wms100InfoFormat(format string) string {
	if format == "application/vnd.ogc.gml" || format == "text/xml" {
		return "GML.1"
	}
	return "MIME"
}

func wms111MetaDataType(type_ string) string {
	if type_ == "ISO19115:2003" {
		return "TC211"
	}
	if type_ == "FGDC:1998" {
		return "FGDC"
	}
	return ""
}
