package dobo

import (
	"encoding"
	"encoding/json"
)

var _ encoding.BinaryMarshaler = (*CaptchaDO)(nil)
var _ encoding.BinaryUnmarshaler = (*CaptchaDO)(nil)

type (
	CaptchaBO struct {
		Id       string `json:"id"`
		Value    string `json:"value"`
		Image    string `json:"image"`
		ExpireAt int64  `json:"expireAt"`
	}
	CaptchaDO struct {
		Id       string `json:"id"`
		Value    string `json:"value"`
		Image    string `json:"image"`
		ExpireAt int64  `json:"expireAt"`
	}
)

// UnmarshalBinary 用于redis映射
func (c *CaptchaDO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary 用于redis映射
func (c *CaptchaDO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

func NewCaptchaBo(values ...*CaptchaBO) IBO[*CaptchaBO, *CaptchaDO] {
	return NewBO[*CaptchaBO, *CaptchaDO](
		BOWithValues[*CaptchaBO, *CaptchaDO](values...),
		BOWithDToB[*CaptchaBO, *CaptchaDO](captchaDOToBo),
		BOWithBToD[*CaptchaBO, *CaptchaDO](captchaBoToDO),
	)
}

func NewCaptchaDO(values ...*CaptchaDO) IDO[*CaptchaBO, *CaptchaDO] {
	return NewDO[*CaptchaBO, *CaptchaDO](
		DOWithValues[*CaptchaBO, *CaptchaDO](values...),
		DOWithBToD[*CaptchaBO, *CaptchaDO](captchaBoToDO),
		DOWithDToB[*CaptchaBO, *CaptchaDO](captchaDOToBo),
	)
}

func captchaBoToDO(b *CaptchaBO) *CaptchaDO {
	if b == nil {
		return nil
	}
	return &CaptchaDO{
		Id:       b.Id,
		Value:    b.Value,
		Image:    b.Image,
		ExpireAt: b.ExpireAt,
	}
}

func captchaDOToBo(d *CaptchaDO) *CaptchaBO {
	if d == nil {
		return nil
	}
	return &CaptchaBO{
		Id:       d.Id,
		Value:    d.Value,
		Image:    d.Image,
		ExpireAt: d.ExpireAt,
	}
}
