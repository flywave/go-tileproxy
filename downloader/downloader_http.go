package downloader

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"

	"golang.org/x/net/html/charset"
)

type HttpDownloader struct {
}

func NewHttpDownloader() *HttpDownloader {
	return &HttpDownloader{}
}

func (this *HttpDownloader) Download(req *request.Request) tile.Tile {
	var p = tile.NewTile(req)
	return this.downloadTile(p, req)
}

func (this *HttpDownloader) changeCharsetEncodingAuto(contentTypeStr string, sor io.ReadCloser) []byte {
	var err error
	destReader, err := charset.NewReader(sor, contentTypeStr)

	if err != nil {
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		//mlog.LogInst().LogError(err.Error())
	}

	return sorbody
}

func (this *HttpDownloader) changeCharsetEncodingAutoGzipSupport(contentTypeStr string, sor io.ReadCloser) []byte {
	var err error
	gzipReader, err := gzip.NewReader(sor)
	if err != nil {
		return nil
	}
	defer gzipReader.Close()
	destReader, err := charset.NewReader(gzipReader, contentTypeStr)

	if err != nil {
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		//mlog.LogInst().LogError(err.Error())
		// return ""
	}

	return sorbody
}

func connectByHttp(p tile.Tile, req *request.Request) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: req.GetRedirectFunc(),
	}

	httpreq, err := http.NewRequest(req.GetMethod(), req.GetUrl(), strings.NewReader(req.GetPostdata()))
	if header := req.GetHeader(); header != nil {
		httpreq.Header = req.GetHeader()
	}

	if cookies := req.GetCookies(); cookies != nil {
		for i := range cookies {
			httpreq.AddCookie(cookies[i])
		}
	}

	var resp *http.Response
	if resp, err = client.Do(httpreq); err != nil {
		if e, ok := err.(*url.Error); ok && e.Err != nil && e.Err.Error() == "normal" {
			//  normal
		} else {
			p.SetStatus(true, err.Error())
			return nil, err
		}
	}

	return resp, nil
}

func connectByHttpProxy(p tile.Tile, in_req *request.Request) (*http.Response, error) {
	request, _ := http.NewRequest("GET", in_req.GetUrl(), nil)
	proxy, err := url.Parse(in_req.GetProxyHost())
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func (this *HttpDownloader) downloadFile(p tile.Tile, req *request.Request) (tile.Tile, []byte) {
	var err error
	var urlstr string
	if urlstr = req.GetUrl(); len(urlstr) == 0 {
		p.SetStatus(true, "url is empty")
		return p, nil
	}

	var resp *http.Response

	if proxystr := req.GetProxyHost(); len(proxystr) != 0 {
		resp, err = connectByHttpProxy(p, req)
	} else {
		resp, err = connectByHttp(p, req)
	}

	if err != nil {
		return p, nil
	}

	p.SetHeader(resp.Header)
	p.SetCookies(resp.Cookies())

	var bodyStr []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		bodyStr = this.changeCharsetEncodingAutoGzipSupport(resp.Header.Get("Content-Type"), resp.Body)
	} else {
		bodyStr = this.changeCharsetEncodingAuto(resp.Header.Get("Content-Type"), resp.Body)
	}

	defer resp.Body.Close()
	return p, bodyStr
}

func (this *HttpDownloader) downloadTile(p tile.Tile, req *request.Request) tile.Tile {
	p, destbody := this.downloadFile(p, req)
	if !p.Success() {
		return p
	}

	p.SetBody(destbody).SetStatus(false, "")
	return p
}
