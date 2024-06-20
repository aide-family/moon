package captcha

import (
	"context"
	"testing"
)

func TestCreateCode(t *testing.T) {
	code, s, err := CreateCode(context.Background(), TypeDigit, "dark", 100, 200)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(code)
	t.Log(s)
	t.Log(GetCodeAnswer(code))
	if VerifyCaptcha(code, GetCodeAnswer(code)) {
		t.Log("验证成功")
		return
	}
	t.Error("验证失败")
}
