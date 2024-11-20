package httpx

import (
	"context"
	"testing"
)

const baidu = "https://www.baidu.com"

func TestNewHttpX(t *testing.T) {
	h := NewHTTPX()
	t.Log(h.POST(context.Background(), baidu, nil))
}
