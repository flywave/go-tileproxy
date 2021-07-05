package resource

import (
	"crypto/md5"
	"fmt"

	"github.com/flywave/go-tileproxy/images"
)

type StyleCache struct {
	LocalCache
}

func NewStyleCache(cache_dir string, file_ext string) *StyleCache {
	return &StyleCache{LocalCache: LocalCache{CacheDir: cache_dir, FileExt: file_ext}}
}

type SpriteCache struct {
	LocalCache
}

func NewSpriteCache(cache_dir string, file_ext string) *SpriteCache {
	return &SpriteCache{LocalCache: LocalCache{CacheDir: cache_dir, FileExt: file_ext}}
}

type GlyphsCache struct {
	LocalCache
}

func NewGlyphsCache(cache_dir string, file_ext string) *GlyphsCache {
	return &GlyphsCache{LocalCache: LocalCache{CacheDir: cache_dir, FileExt: file_ext}}
}

type Style struct {
	BaseResource
	style   string
	sprites []string
	glyphs  []string
}

func (l *Style) GetData() []byte {
	return []byte(l.style)
}

func (l *Style) SetData(buf []byte) {
	l.style = string(buf)
}

func (l *Style) Hash() []byte {
	m := md5.New()
	m.Write([]byte(l.ID))
	return m.Sum(nil)
}

type Sprite struct {
	BaseResource
	Source images.Source
	Scale  int
}

func (l *Sprite) GetData() []byte {
	if l.Source != nil {
		return l.Source.GetBuffer(nil, nil)
	}
	return []byte{}
}

func (l *Sprite) SetData(data []byte) {
	l.Source = images.CreateImageSourceFromBufer(data)
}

func (l *Sprite) Hash() []byte {
	m := md5.New()
	m.Write([]byte(l.ID))
	m.Write([]byte(fmt.Sprintf("%d", l.Scale)))
	return m.Sum(nil)
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
