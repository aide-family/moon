package bo

import (
	"encoding"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/captcha"
)

// redis缓存实现
var _ encoding.BinaryMarshaler = (*ValidateCaptchaItem)(nil)
var _ encoding.BinaryUnmarshaler = (*ValidateCaptchaItem)(nil)

type ValidateCaptchaItem struct {
	ValidateCaptchaParams
	ExpireAt int64 `json:"expireAt"`
}

func (l *ValidateCaptchaItem) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}

func (l *ValidateCaptchaItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(l)
}

type GenerateCaptchaParams struct {
	Type  captcha.Type
	Theme captcha.Theme
	Size  []int
}

type CaptchaItem struct {
	ValidateCaptchaItem
	Base64s string `json:"base64s"`
}

type ValidateCaptchaParams struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}
