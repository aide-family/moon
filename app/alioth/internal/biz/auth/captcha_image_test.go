package auth

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/image/font"
)

// mathQuestionPattern 匹配 newMathCaptcha 产出的算式，用于核对题目和答案是否一致。
var mathQuestionPattern = regexp.MustCompile(`^([0-9]{1,3}) ([+\-*/]) ([0-9]{1,3}) = \?$`)

func TestMathCaptchaQuestionMatchesAnswer(t *testing.T) {
	for i := 0; i < 200; i++ {
		question, answer, err := newMathCaptcha()
		if err != nil {
			t.Fatalf("第 %d 次生成数学题: %v", i, err)
		}
		match := mathQuestionPattern.FindStringSubmatch(question)
		if match == nil {
			t.Fatalf("第 %d 次生成的题目格式不对: %q", i, question)
		}
		left, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatalf("第 %d 次生成左操作数: %v", i, err)
		}
		right, err := strconv.Atoi(match[3])
		if err != nil {
			t.Fatalf("第 %d 次生成右操作数: %v", i, err)
		}
		if left > 100 || right > 100 {
			t.Fatalf("第 %d 次操作数超出 100: %q", i, question)
		}
		var want int
		switch match[2] {
		case "+":
			want = left + right
		case "-":
			if left <= right {
				t.Fatalf("第 %d 次减法出现非正结果: %q", i, question)
			}
			want = left - right
		case "*":
			want = left * right
		case "/":
			if right == 0 || left%right != 0 {
				t.Fatalf("第 %d 次除法不能整除: %q", i, question)
			}
			want = left / right
		default:
			t.Fatalf("第 %d 次出现未知运算符: %q", i, question)
		}
		if want < 0 || want > 100 {
			t.Fatalf("第 %d 次结果 %d 超出 0..100: %q", i, want, question)
		}
		if answer != strconv.Itoa(want) {
			t.Fatalf("第 %d 次题目 %q 的答案 = %q，应为 %d", i, question, answer, want)
		}
	}
}

// TestMathCaptchaOperators 确认 200 次生成里四种运算符都出现过。概率上连续 200 次缺同一种可以忽略。
func TestMathCaptchaOperators(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		question, _, err := newMathCaptcha()
		if err != nil {
			t.Fatalf("第 %d 次生成数学题: %v", i, err)
		}
		match := mathQuestionPattern.FindStringSubmatch(question)
		if match == nil {
			t.Fatalf("第 %d 次生成的题目格式不对: %q", i, question)
		}
		seen[match[2]] = true
	}
	for _, operator := range []string{"+", "-", "*", "/"} {
		if !seen[operator] {
			t.Fatalf("200 次生成没有出现运算符 %q，实际为 %v", operator, seen)
		}
	}
}

func TestNumericCaptcha(t *testing.T) {
	for i := 0; i < 100; i++ {
		question, answer, err := newNumericCaptcha()
		if err != nil {
			t.Fatalf("第 %d 次生成数字码: %v", i, err)
		}
		if len(question) != 6 {
			t.Fatalf("第 %d 次数字码长度 = %d，应为 6: %q", i, len(question), question)
		}
		if strings.HasPrefix(question, "0") {
			t.Fatalf("第 %d 次数字码以 0 开头: %q", i, question)
		}
		if answer != question {
			t.Fatalf("第 %d 次数字码 %q 的答案 = %q，应与题目一致", i, question, answer)
		}
	}
}

// TestNewCaptchaCoversBothKinds 确认两种题型都会被生成，避免题型开关写死。
func TestNewCaptchaCoversBothKinds(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		question, answer, kind, err := newCaptcha()
		if err != nil {
			t.Fatalf("第 %d 次生成图形验证码: %v", i, err)
		}
		if question == "" || answer == "" {
			t.Fatalf("第 %d 次生成空题目或空答案", i)
		}
		seen[kind] = true
	}
	if !seen[captchaKindNumeric] || !seen[captchaKindMath] {
		t.Fatalf("200 次生成只出现了题型 %v", seen)
	}
}

// TestCaptchaQuestionFitsImage 确认最长的题目也能完整画进图片，不会被裁掉。
func TestCaptchaQuestionFitsImage(t *testing.T) {
	face, err := captchaFace()
	if err != nil {
		t.Fatalf("captchaFace: %v", err)
	}
	for _, question := range []string{"100 / 10 = ?", "100 - 99 = ?", "100 + 0 = ?", "10 * 10 = ?", "999999"} {
		if width := font.MeasureString(face, question).Ceil(); width > captchaWidth-16 {
			t.Fatalf("题目 %q 宽 %dpx，超出图片可用宽度 %dpx", question, width, captchaWidth-16)
		}
	}
}

func TestDrawCaptcha(t *testing.T) {
	imageURL, err := drawCaptcha("7 + 5 = ?")
	if err != nil {
		t.Fatalf("drawCaptcha: %v", err)
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(imageURL, prefix) {
		t.Fatalf("缺少 data URL 前缀")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(imageURL, prefix))
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("png.DecodeConfig: %v", err)
	}
	if cfg.Width != captchaWidth || cfg.Height != captchaHeight {
		t.Fatalf("size = %dx%d, want %dx%d", cfg.Width, cfg.Height, captchaWidth, captchaHeight)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	if img.Bounds() != image.Rect(0, 0, captchaWidth, captchaHeight) {
		t.Fatalf("bounds = %v", img.Bounds())
	}
	if !captchaHasText(img) {
		t.Fatal("图片里没有画出题目")
	}
}

// captchaHasText 粗略判断图片里是否画了深色文字。噪点和干扰线是浅灰，
// 文字用 20..100 的深色，因此深色像素数量明显更多。
func captchaHasText(img image.Image) bool {
	dark := 0
	for y := 0; y < captchaHeight; y++ {
		for x := 0; x < captchaWidth; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r>>8 < 120 && g>>8 < 120 && b>>8 < 120 {
				dark++
			}
		}
	}
	return dark > 50
}
