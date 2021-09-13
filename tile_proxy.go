package tileproxy

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/flywave/go-tileproxy/setting"
)

type TileProxy struct {
	m               sync.RWMutex
	Datasets        map[string]*Dataset
	basePath        string
	globals         *setting.GlobalsSetting
	datasetReqRegex *regexp.Regexp
}

func (t *TileProxy) UpdateDataset(dataset string, d *setting.ProxyDataset) {
	t.m.Lock()
	t.Datasets[dataset] = NewDataset(d, t.basePath, t.globals)
	t.m.Unlock()
}

func (t *TileProxy) RemoveDataset(dataset string) {
	var d *Dataset
	t.m.Lock()
	if d_, ok := t.Datasets[dataset]; ok {
		d = d_
	}
	delete(t.Datasets, dataset)
	t.m.Unlock()
	d.Clean()
}

func (t *TileProxy) Reload(proxy []*setting.ProxyDataset) {
	t.m.Lock()
	t.Datasets = make(map[string]*Dataset)
	for i := range proxy {
		t.Datasets[proxy[i].Identifier] = NewDataset(proxy[i], t.basePath, t.globals)
	}
	t.m.Unlock()
}

func (t *TileProxy) parseServiceId(r *http.Request) string {
	match := t.datasetReqRegex.FindStringSubmatch(r.URL.Path)
	groupNames := t.datasetReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if len(match) == 0 {
		return ""
	}

	return result["service"]
}

func (s *TileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceId := s.parseServiceId(r)
	if serviceId == "" {
		w.WriteHeader(404)
		return
	}
	s.m.Lock()
	if d, ok := s.Datasets[serviceId]; ok {
		s.m.Unlock()
		d.ServeHTTP(w, r)
	} else {
		s.m.Unlock()
		w.WriteHeader(404)
	}
}

func NewTileProxy(basePath string, globals *setting.GlobalsSetting, proxys []*setting.ProxyDataset) *TileProxy {
	proxy := &TileProxy{basePath: basePath, globals: globals}
	proxy.Reload(proxys)
	return proxy
}
