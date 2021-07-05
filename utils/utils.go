package utils

import "strings"

func SplitMimeType(mime_type string) string {
	if strings.Contains(mime_type, "/") {
		strs := strings.Split(mime_type, "/")
		return strs[0]
	}
	return ""
}
