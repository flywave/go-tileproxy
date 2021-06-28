package downloader

import (
	"testing"

	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

func TestDownloadHtml(t *testing.T) {
	var req *request.Request
	req = request.NewRequest("http://finance.sina.com.cn/7x24/", "html", "", "GET", "", nil, nil, nil, nil)

	var dl Downloader
	dl = NewHttpDownloader()

	var p tile.Tile
	p = dl.Download(req)

	print(string(p.GetBody()))
}
