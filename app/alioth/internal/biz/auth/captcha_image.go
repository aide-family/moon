package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/big"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	captchaWidth    = 420
	captchaHeight   = 140
	captchaFontSize = 42
)

func drawCaptcha(code string) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, captchaWidth, captchaHeight))
	fillCaptchaBackground(img)
	drawCaptchaNoise(img)
	drawCaptchaLines(img)
	if err := drawCaptchaChars(img, code); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func fillCaptchaBackground(img *image.RGBA) {
	bg := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
}

func drawCaptchaNoise(img *image.RGBA) {
	bounds := img.Bounds()
	for range 420 {
		x := randInt(bounds.Dx())
		y := randInt(bounds.Dy())
		img.Set(x, y, randomGray(80, 200))
	}
}

func drawCaptchaLines(img *image.RGBA) {
	n := 2 + randInt(3)
	for range n {
		c := randomGray(60, 160)
		x0 := randInt(captchaWidth / 4)
		y0 := randInt(captchaHeight)
		x1 := captchaWidth*3/4 + randInt(captchaWidth/4)
		y1 := randInt(captchaHeight)
		drawLine(img, x0, y0, x1, y1, c)
	}
}

func captchaFace() (font.Face, error) {
	parsed, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    captchaFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func drawCaptchaChars(dst *image.RGBA, code string) error {
	face, err := captchaFace()
	if err != nil {
		return err
	}
	runes := []rune(code)
	slot := captchaWidth / max(len(runes), 1)
	for i, char := range runes {
		glyph := image.NewRGBA(image.Rect(0, 0, 72, 84))
		drawer := &font.Drawer{
			Dst:  glyph,
			Src:  image.NewUniform(randomTextColor()),
			Face: face,
			Dot:  fixed.P(10, 62),
		}
		drawer.DrawString(string(char))

		angle := (randFloat() - 0.5) * 0.5 // about ±14 degrees
		rotated := rotateRGBA(glyph, angle)
		x := 12 + i*slot + randInt(8)
		y := 16 + randInt(16)
		draw.Draw(dst, image.Rect(x, y, x+rotated.Bounds().Dx(), y+rotated.Bounds().Dy()), rotated, rotated.Bounds().Min, draw.Over)
	}
	return nil
}

func rotateRGBA(src *image.RGBA, angle float64) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	// Bound of rotated rectangle.
	nw := int(math.Ceil(math.Abs(float64(w)*cosA) + math.Abs(float64(h)*sinA)))
	nh := int(math.Ceil(math.Abs(float64(w)*sinA) + math.Abs(float64(h)*cosA)))
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	out := image.NewRGBA(image.Rect(0, 0, nw, nh))
	cx, cy := float64(w)/2, float64(h)/2
	ncx, ncy := float64(nw)/2, float64(nh)/2
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			dx := float64(x) - ncx
			dy := float64(y) - ncy
			sx := int(math.Round(dx*cosA + dy*sinA + cx))
			sy := int(math.Round(-dx*sinA + dy*cosA + cy))
			if sx >= 0 && sx < w && sy >= 0 && sy < h {
				out.Set(x, y, src.At(bounds.Min.X+sx, bounds.Min.Y+sy))
			}
		}
	}
	return out
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func randomTextColor() color.RGBA {
	return color.RGBA{
		R: uint8(20 + randInt(80)),
		G: uint8(20 + randInt(80)),
		B: uint8(20 + randInt(80)),
		A: 255,
	}
}

func randomGray(minV, maxV int) color.RGBA {
	if maxV <= minV {
		maxV = minV + 1
	}
	v := uint8(minV + randInt(maxV-minV))
	return color.RGBA{R: v, G: v, B: v, A: 255}
}

func randInt(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(v.Int64())
}

func randFloat() float64 {
	return float64(randInt(10000)) / 10000
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
