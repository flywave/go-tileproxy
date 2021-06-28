package autorange

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flywave/go-tileproxy/backend"
	"github.com/flywave/go-tileproxy/filters"
	"github.com/flywave/go-tileproxy/utils"
)

const (
	filterName string = "autorange"
)

type Config struct {
	Sites          []string
	SupportFilters []string
	MaxSize        int
	BufSize        int
	Threads        int
}

func (c *Config) fromKVS(conf map[string]backend.KVS) error {
	return nil
}

type Filter struct {
	Config
	SiteMatcher    *utils.HostMatcher
	SupportFilters map[string]struct{}
	MaxSize        int
	BufSize        int
	Threads        int
}

func init() {
	filters.Register(filterName, func(conf map[string]backend.KVS) (filters.Filter, error) {
		config := new(Config)
		err := config.fromKVS(conf)
		if err != nil {
			//logger.Fatal(err, "storage.ReadJsonConfig failed")
		}
		return NewFilter(config)
	})
}

func NewFilter(config *Config) (filters.Filter, error) {
	f := &Filter{
		Config:         *config,
		SiteMatcher:    utils.NewHostMatcher(config.Sites),
		SupportFilters: make(map[string]struct{}),
		MaxSize:        config.MaxSize,
		BufSize:        config.BufSize,
		Threads:        config.Threads,
	}

	for _, name := range config.SupportFilters {
		f.SupportFilters[name] = struct{}{}
	}

	return f, nil
}

func (f *Filter) FilterName() string {
	return filterName
}

func (f *Filter) Request(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
	if req.Method != http.MethodGet || strings.Contains(req.URL.RawQuery, "range=") {
		return ctx, req, nil
	}

	if r := req.Header.Get("Range"); r == "" {
		switch {
		case f.SiteMatcher.Match(req.Host):
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", 0, f.MaxSize))
			ctx = filters.WithBool(ctx, "autorange.site", true)
		default:
		}
	} else {
		ctx = filters.WithBool(ctx, "autorange.default", true)
		parts := strings.Split(r, "=")
		switch parts[0] {
		case "bytes":
			parts1 := strings.Split(parts[1], "-")
			if start, err := strconv.Atoi(parts1[0]); err == nil {
				if end, err := strconv.Atoi(parts1[1]); err != nil || end-start > f.MaxSize {
					req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, start+f.MaxSize))
				}
			}
		default:
		}
	}

	return ctx, req, nil
}

func (f *Filter) Response(ctx context.Context, resp *http.Response) (context.Context, *http.Response, error) {
	if resp.StatusCode != http.StatusPartialContent || resp.Header.Get("Content-Length") == "" {
		return ctx, resp, nil
	}

	if ok1, ok := filters.Bool(ctx, "autorange.default"); ok && ok1 {
		return ctx, resp, nil
	}

	f1 := filters.GetRoundTripFilter(ctx)
	if f1 == nil {
		return ctx, resp, nil
	}
	if _, ok := f.SupportFilters[f1.FilterName()]; !ok {
		return ctx, resp, nil
	}

	parts := strings.Split(resp.Header.Get("Content-Range"), " ")
	if len(parts) != 2 || parts[0] != "bytes" {
		return ctx, resp, nil
	}

	parts1 := strings.Split(parts[1], "/")
	parts2 := strings.Split(parts1[0], "-")
	if len(parts1) != 2 || len(parts2) != 2 {
		return ctx, resp, nil
	}

	var end, length int64
	var err error

	end, err = strconv.ParseInt(parts2[1], 10, 64)
	if err != nil {
		return ctx, resp, nil
	}

	length, err = strconv.ParseInt(parts1[1], 10, 64)
	if err != nil {
		return ctx, resp, nil
	}

	resp.ContentLength = length
	resp.Header.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	resp.Header.Del("Content-Range")

	r, w := AutoPipe(f.Threads)

	go func(w *autoPipeWriter, filter filters.RoundTripFilter, req0 *http.Request, start, length int64) {

		req, err := http.NewRequest(req0.Method, req0.URL.String(), nil)
		if err != nil {
			return
		}

		for key, values := range req0.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		if err := w.WaitForReading(); err != nil {
			return
		}
		var index uint32
		for {
			if w.FatalErr() {
				break
			}
			if start > length-1 {
				w.Close()
				break
			}
			if w.Len() > 128*1024*1024 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			end := start + int64(1024<<10-1)
			if end > length-1 {
				end = length - 1
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

			_, resp, err := filter.RoundTrip(nil, req)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			if resp.StatusCode != http.StatusPartialContent {
				if resp.StatusCode >= http.StatusBadRequest {
					time.Sleep(1 * time.Second)
				}
				continue
			}

			w.ThreadHello()
			go func(index uint32, resp *http.Response) {
				defer resp.Body.Close()
				defer w.ThreadBye()

				piper := w.NewPiper(index)
				_, err := utils.IOCopy(piper, resp.Body)
				if err != nil {
					piper.EIndex()
				}
				piper.WClose()
			}(index, resp)

			start = end + 1
			index++
		}
	}(w, f1, resp.Request, end+1, length)

	resp.Body = utils.NewMultiReadCloser(resp.Body, r)

	return ctx, resp, nil
}
