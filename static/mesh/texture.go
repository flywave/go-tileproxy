package mesh

import (
	"image"

	"github.com/flywave/go-tileproxy/static/draw"
)

type TextureSource struct {
}

func (s *TextureSource) GetTexture() image.Image {
	return nil
}

type DrawTextureSource struct {
	ctx draw.Context
}

func (s *DrawTextureSource) GetTexture() image.Image {
	return nil
}
