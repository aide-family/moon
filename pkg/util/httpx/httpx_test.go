package httpx

import (
	"testing"
)

const baidu = "https://www.baidu.com"

func TestNewHttpX(t *testing.T) {
	h := NewHttpX()
	t.Log(h.POST(baidu, nil))
}
