package auth

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"strings"
	"testing"
)

func TestDrawCaptcha(t *testing.T) {
	imageURL, err := drawCaptcha("XK7M")
	if err != nil {
		t.Fatalf("drawCaptcha: %v", err)
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(imageURL, prefix) {
		t.Fatalf("missing data url prefix")
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
}
