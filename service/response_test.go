package service

import (
	"testing"
	"time"
)

func TestResponse(t *testing.T) {
	response := []byte{0}

	rep := NewResponse(response, 200, DefaultContentType)

	ct := rep.GetContentType()

	if ct == "" {
		t.FailNow()
	}

	if rep.GetStatus() != 200 {
		t.FailNow()
	}

	rep.SetETag(etagFor([]byte("test")))

	etag := rep.GetETag()

	if etag != etagFor([]byte("test")) {
		t.FailNow()
	}

	maxTileAge := time.Hour

	now := time.Now()

	rep.cacheHeaders(&now, []string{"test"}, int(maxTileAge.Seconds()))

	rep.SetLastModified(now)
}
