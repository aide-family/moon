package bo

import (
	"encoding"
	"encoding/json"
)

var _ encoding.BinaryMarshaler = (*CaptchaBO)(nil)
var _ encoding.BinaryUnmarshaler = (*CaptchaBO)(nil)

type (
	CaptchaBO struct {
		Id       string `json:"id"`
		Value    string `json:"value"`
		Image    string `json:"image"`
		ExpireAt int64  `json:"expireAt"`
	}
)

// Bytes 用于redis映射
func (c *CaptchaBO) Bytes() []byte {
	if c == nil {
		return nil
	}
	bs, _ := json.Marshal(c)
	return bs
}

// UnmarshalBinary 用于redis映射
func (c *CaptchaBO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary 用于redis映射
func (c *CaptchaBO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}
