package biz

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func drawCaptcha(code string) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, 160, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 160; x++ {
			img.Set(x, y, color.RGBA{R: 245, G: 245, B: 245, A: 255})
		}
	}
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 20, G: 20, B: 20, A: 255}),
		Face: basicfont.Face7x13,
	}
	for i, char := range code {
		drawer.Dot = fixed.P(16+i*34, 30)
		drawer.DrawString(string(char))
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
