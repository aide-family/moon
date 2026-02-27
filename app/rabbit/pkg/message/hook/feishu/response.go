package feishu

import (
	"encoding/json"
	"fmt"
	"io"
)

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (l *response) Error() string {
	if l.Code == 0 {
		return ""
	}
	return fmt.Sprintf("code: %d, msg: %s, data: %v", l.Code, l.Msg, l.Data)
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
