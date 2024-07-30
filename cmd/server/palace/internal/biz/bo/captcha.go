package bo

import (
	"encoding"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/captcha"
)

// redis缓存实现
var _ encoding.BinaryMarshaler = (*ValidateCaptchaItem)(nil)
var _ encoding.BinaryUnmarshaler = (*ValidateCaptchaItem)(nil)

// ValidateCaptchaItem 验证码缓存明细
type ValidateCaptchaItem struct {
	ValidateCaptchaParams
	ExpireAt int64 `json:"expireAt"`
}

// UnmarshalBinary 实现encoding.UnmarshalBinary
func (l *ValidateCaptchaItem) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}

// MarshalBinary 实现encoding.MarshalBinary
func (l *ValidateCaptchaItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(l)
}

// GenerateCaptchaParams 生成验证码参数
type GenerateCaptchaParams struct {
	Type  captcha.Type
	Theme captcha.Theme
	Size  []int
}

// CaptchaItem 验证码明细
type CaptchaItem struct {
	ValidateCaptchaItem
	Base64s string `json:"base64s"`
}

// ValidateCaptchaParams 验证码参数
type ValidateCaptchaParams struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
