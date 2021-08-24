package resource

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/flywave/go-mapbox/style"
	"github.com/flywave/go-tileproxy/imagery"
)

type StyleCache struct {
	store Store
}

func NewStyleCache(cache_dir string, file_ext string) *StyleCache {
	return &StyleCache{store: &LocalStore{CacheDir: cache_dir, FileExt: file_ext}}
}

func (c *StyleCache) Save(r Resource) error {
	return c.store.Save(r)
}

func (c *StyleCache) Load(r Resource) error {
	return c.Load(r)
}

type GlyphsCache struct {
	store Store
}

func (c *GlyphsCache) Save(r Resource) error {
	return c.store.Save(r)
}

func (c *GlyphsCache) Load(r Resource) error {
	return c.Load(r)
}

func NewGlyphsCache(cache_dir string, file_ext string) *GlyphsCache {
	return &GlyphsCache{store: &LocalStore{CacheDir: cache_dir, FileExt: file_ext}}
}

type Style struct {
	BaseResource
	style style.Style
}

func (l *Style) GetData() []byte {
	var jdata []byte
	buf := bytes.NewBuffer(jdata)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(l.style); err != nil {
		return nil
	}
	return buf.Bytes()
}

func (l *Style) SetData(buf []byte) {
	dec := json.NewDecoder(bytes.NewBuffer(buf))
	if err := dec.Decode(&l.style); err != nil {
		return
	}
}

func (l *Style) Hash() []byte {
	m := md5.New()
	m.Write([]byte(l.ID))
	return m.Sum(nil)
}

func CreateStyle(content []byte) *Style {
	s := &Style{}
	s.SetData(content)
	return s
}

type SpriteJSON struct {
	BaseResource
	Buffer []byte
}

func (l *SpriteJSON) GetData() []byte {
	return l.Buffer
}

func (l *SpriteJSON) SetData(data []byte) {
	l.Buffer = data
}

func (l *SpriteJSON) Hash() []byte {
	m := md5.New()
	m.Write([]byte(l.ID))
	return m.Sum(nil)
}

func CreateSpriteJSON(content []byte) *SpriteJSON {
	return &SpriteJSON{Buffer: content}
}

type Sprite struct {
	BaseResource
	Source  *imagery.ImageSource
	Scale   int
	Options *imagery.ImageOptions
}

func (l *Sprite) GetData() []byte {
	if l.Source != nil {
		return l.Source.GetBuffer(nil, nil)
	}
	return []byte{}
}

func (l *Sprite) SetData(data []byte) {
	l.Source = imagery.CreateImageSourceFromBufer(data, l.Options)
}

func (l *Sprite) Hash() []byte {
	m := md5.New()
	m.Write([]byte(l.ID))
	m.Write([]byte(fmt.Sprintf("%d", l.Scale)))
	return m.Sum(nil)
}

func CreateSprite(content []byte) *Sprite {
	s := &Sprite{}
	s.SetData(content)
	return s
}

type Glyphs struct {
	BaseResource
	Buffer []byte
}

func (l *Glyphs) GetData() []byte {
	return l.Buffer
}

func (l *Glyphs) SetData(data []byte) {
	l.Buffer = data
}

func (l *Glyphs) Hash() []byte {
	m := md5.New()
	m.Write([]byte(l.ID))
	return m.Sum(nil)
}

func CreateGlyphs(content []byte) *Glyphs {
	return &Glyphs{Buffer: content}
}
