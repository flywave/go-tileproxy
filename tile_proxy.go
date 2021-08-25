package tileproxy

import (
	"net/http"
	"sync"

	"github.com/flywave/go-tileproxy/setting"
)

type TileProxy struct {
	m        sync.RWMutex
	Datasets map[string]Dataset
}

func (t *TileProxy) AddDataset(dataset string, d *setting.ProxyDataset) {

}

func (t *TileProxy) UpdateDataset(dataset string, d *setting.ProxyDataset) {

}

func (t *TileProxy) HasDataset(dataset string) bool {
	return false
}

func (t *TileProxy) RemoveDataset(dataset string) {

}

func (t *TileProxy) CleanDataset(dataset string) {

}

func (t *TileProxy) UpdateAll(proxy []*setting.ProxyDataset) {

}

func NewTileProxy(proxy []*setting.ProxyDataset) *TileProxy {
	return nil
}

func (s *TileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
