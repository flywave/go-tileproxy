package images

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"testing"

	"github.com/flywave/imaging"
)

var (
	PNG_FORMAT  = &ImageOptions{Format: ImageFormat("image/png")}
	JPEG_FORMAT = &ImageOptions{Format: ImageFormat("image/jpeg")}
	TIFF_FORMAT = &ImageOptions{Format: ImageFormat("image/tiff")}
)

func createTmpImageFile(size [2]uint32) string {
	fd, err := ioutil.TempFile("", "test-*.png")
	if err != nil {
		fmt.Println(err)
	}
	r := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: int(size[0]), Y: int(size[1])}}
	img := image.NewNRGBA(r)
	imaging.Encode(fd, img, imaging.PNG)
	defer fd.Close()
	return fd.Name()
}

func TestImageSource(t *testing.T) {
	tmp_filename := createTmpImageFile([2]uint32{100, 100})
	defer os.Remove(tmp_filename)

	ir := &ImageSource{Options: *PNG_FORMAT}
	ir.SetSource(tmp_filename)

	if !isPng(string(ir.GetBuffer(nil, nil))) {
		t.FailNow()
	}

	rect := ir.GetImage().Bounds()

	if rect.Dx() != 100 || rect.Dy() != 100 {
		t.FailNow()
	}

	ir = &ImageSource{Options: *PNG_FORMAT}
	tmpfile, _ := os.Open(tmp_filename)
	ir.SetSource(tmpfile)

	if !isPng(string(ir.GetBuffer(nil, nil))) {
		t.FailNow()
	}

	rect = ir.GetImage().Bounds()

	if rect.Dx() != 100 || rect.Dy() != 100 {
		t.FailNow()
	}

}

func TestSubImageSource(t *testing.T) {
	sub_img := CreateImageSource([2]uint32{150, 150}, PNG_FORMAT)
	img := SubImageSource(
		sub_img, [2]uint32{100, 100}, []uint32{0, 0}, PNG_FORMAT, false).GetImage()

	rect := img.Bounds()

	if rect.Dx() != 100 || rect.Dy() != 100 {
		t.FailNow()
	}
}

func TestBlankImageSource(t *testing.T) {
	bi := NewBlankImageSource([2]uint32{100, 100}, PNG_FORMAT, false)

	if !isPng(string(bi.GetBuffer(nil, nil))) {
		t.FailNow()
	}
}

func TestImagingBlend(t *testing.T) {
	img_opts := *PNG_FORMAT
	img_opts.BgColor = color.Transparent

	image1, _ := imaging.Open("./flowers.png")

	image2 := image.NewRGBA(image.Rect(0, 0, 600, 400))

	for y := 0; y < 400; y++ {
		for x := 0; x < 600; x++ {
			image2.Set(x, y, color.RGBA{128, 128, 255, 255})
		}
	}

	result := ImagingBlend(image1, image2, 0.7)

	imaging.Save(result, "./test.png")

	defer os.Remove("./test.png")
}
