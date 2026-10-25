package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	captchaWidth    = 280
	captchaHeight   = 72
	captchaFontSize = 36

	// captchaKindNumeric 是随机六位数字码。
	captchaKindNumeric = "NUMERIC"
	// captchaKindMath 是 100 以内四则运算题。
	captchaKindMath = "MATH"
)

// newCaptcha 随机生成数字码或算式题。题型只用于前端切换输入提示，
// 题目和答案都不出现在接口响应里：题目画进图片，答案用摘要落库。
func newCaptcha() (question, answer, kind string, err error) {
	pick, err := randomInt(0, 1)
	if err != nil {
		return "", "", "", err
	}
	if pick == 0 {
		question, answer, err = newNumericCaptcha()
		return question, answer, captchaKindNumeric, err
	}
	question, answer, err = newMathCaptcha()
	return question, answer, captchaKindMath, err
}

// newNumericCaptcha 生成六位数字码。首位不为 0，答案就是题目本身。
func newNumericCaptcha() (string, string, error) {
	value, err := randomInt(100000, 999999)
	if err != nil {
		return "", "", err
	}
	question := strconv.Itoa(value)
	return question, question, nil
}

// newMathCaptcha 生成 100 以内的四则运算题和对应答案。
// 所有结果都是整数：减法保证被减数大于减数，乘法乘积不超过 100，
// 除法先定除数和商再乘出被除数，因此不会出现小数或除不尽。
func newMathCaptcha() (string, string, error) {
	operator, err := randomInt(0, 3)
	if err != nil {
		return "", "", err
	}
	switch operator {
	case 0:
		left, err := randomInt(1, 99)
		if err != nil {
			return "", "", err
		}
		right, err := randomInt(1, 100-left)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%d + %d = ?", left, right), strconv.Itoa(left + right), nil
	case 1:
		left, err := randomInt(2, 100)
		if err != nil {
			return "", "", err
		}
		right, err := randomInt(1, left-1)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%d - %d = ?", left, right), strconv.Itoa(left - right), nil
	case 2:
		left, err := randomInt(2, 10)
		if err != nil {
			return "", "", err
		}
		right, err := randomInt(2, 100/left)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%d * %d = ?", left, right), strconv.Itoa(left * right), nil
	default:
		// 先定除数和商，再乘出被除数，保证整除且被除数不超过 100。
		divisor, err := randomInt(2, 10)
		if err != nil {
			return "", "", err
		}
		quotient, err := randomInt(2, 100/divisor)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%d / %d = ?", divisor*quotient, divisor), strconv.Itoa(quotient), nil
	}
}

// drawCaptcha 把题目画成单行居中的 PNG，带少量噪点和干扰线，返回 data URL。
func drawCaptcha(question string) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, captchaWidth, captchaHeight))
	fillCaptchaBackground(img)
	drawCaptchaNoise(img)
	drawCaptchaLines(img)
	if err := drawCaptchaText(img, question); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", merr.Server("request failed")
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func fillCaptchaBackground(img *image.RGBA) {
	bg := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
}

// drawCaptchaNoise 只撒少量噪点，避免影响一位数算式的辨识。
func drawCaptchaNoise(img *image.RGBA) {
	bounds := img.Bounds()
	for range 120 {
		x := randInt(bounds.Dx())
		y := randInt(bounds.Dy())
		img.Set(x, y, randomGray(80, 200))
	}
}

func drawCaptchaLines(img *image.RGBA) {
	for range 2 + randInt(2) {
		c := randomGray(60, 170)
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
		return nil, merr.Server("request failed")
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    captchaFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, merr.Server("request failed")
	}
	return face, nil
}

// drawCaptchaText 把整道算式作为一行画在图片正中。
// 基线 = 垂直居中后的顶部 + 字体的 ascent，保证不同算式都居中。
func drawCaptchaText(dst *image.RGBA, text string) error {
	face, err := captchaFace()
	if err != nil {
		return err
	}
	textWidth := font.MeasureString(face, text).Ceil()
	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	x := (captchaWidth - textWidth) / 2
	y := (captchaHeight-(ascent+descent))/2 + ascent
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(randomTextColor()),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(text)
	return nil
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

// randomInt 返回 [minV, maxV] 内的随机整数。随机源失败时返回服务端错误，
// 不退回固定值，否则数学题可能变成同一个算式。
func randomInt(minV, maxV int) (int, error) {
	if maxV <= minV {
		return minV, nil
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(maxV-minV+1)))
	if err != nil {
		return 0, merr.Server("request failed")
	}
	return minV + int(value.Int64()), nil
}

func randInt(n int) int {
	if n <= 0 {
		return 0
	}
	value, err := randomInt(0, n-1)
	if err != nil {
		return 0
	}
	return value
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
