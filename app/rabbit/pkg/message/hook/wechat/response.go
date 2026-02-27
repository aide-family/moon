package wechat

import (
	"encoding/json"
	"fmt"
	"io"
)

type response struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *response) Error() string {
	if l.ErrCode == 0 {
		return ""
	}
	return fmt.Sprintf("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

func unmarshalResponse(body io.ReadCloser) error {
	var resp response
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return err
	}
	if resp.Error() != "" {
		return &resp
	}
	return nil
}
