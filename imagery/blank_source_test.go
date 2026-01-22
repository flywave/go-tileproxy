package imagery

import (
	"image"
	"image/color"
	"testing"
)

func TestBlankImageSource(t *testing.T) {
	opts := &ImageOptions{
		BgColor: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Format:  "png",
		Mode:    RGBA,
	}

	// 创建BlankImageSource
	source := NewBlankImageSource([2]uint32{256, 256}, opts, nil)

	// 检查GetTile方法
	tile := source.GetTile()
	if tile == nil {
		t.Fatal("GetTile returned nil")
	}

	// 检查类型转换
	img, ok := tile.(image.Image)
	if !ok {
		t.Fatalf("GetTile returned %T, not image.Image", tile)
	}

	// 检查图像尺寸
	bounds := img.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 256 {
		t.Fatalf("Expected 256x256 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	t.Logf("BlankImageSource test passed, image type: %T", img)
}
