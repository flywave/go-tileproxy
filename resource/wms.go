package resource

import (
	"crypto/md5"
	"fmt"

	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

type LegendCache struct {
	store Store
}

func (c *LegendCache) Save(r Resource) error {
	return c.store.Save(r)
}

func (c *LegendCache) Load(r Resource) error {
	return c.store.Load(r)
}

func NewLegendCache(store Store) *LegendCache {
	return &LegendCache{store: store}
}

type Legend struct {
	BaseResource
	Source  tile.Source
	Scale   int
	Options *imagery.ImageOptions
}

func (l *Legend) GetExtension() string {
	return l.Options.Format.Extension()
}

func (r *Legend) GetFileName() string {
	return fmt.Sprintf("legend-%d", r.Scale)
}

func (l *Legend) GetData() []byte {
	if l.Source != nil {
		return l.Source.GetBuffer(nil, nil)
	}
	return []byte{}
}

func (l *Legend) SetData(data []byte) {
	l.Source = imagery.CreateImageSourceFromBufer(data, l.Options)
}

func (l *Legend) Hash() []byte {
	m := md5.New()
	m.Write([]byte(l.StoreID))
	m.Write([]byte(fmt.Sprintf("%d", l.Scale)))
	return m.Sum(nil)
}
