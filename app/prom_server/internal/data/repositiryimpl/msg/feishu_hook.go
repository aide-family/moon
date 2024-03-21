package msg

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/httpx"
)

var _ HookNotify = (*feishuNotify)(nil)

type feishuNotify struct{}

type FeiShuHookResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (l *FeiShuHookResp) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", l.Code, l.Msg)
}

func (l *feishuNotify) Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error {
	notifyMsg := make(map[string]any)
	_ = json.Unmarshal([]byte(msg.Content), &notifyMsg)
	timeNow := time.Now()
	notifyMsg["timestamp"] = strconv.FormatInt(timeNow.Unix(), 10)
	notifyMsg["sign"] = GenSign(msg.Secret, timeNow.Unix())
	notifyMsgBytes, _ := json.Marshal(notifyMsg)
	response, err := httpx.NewHttpX().POSTWithContext(ctx, url, notifyMsgBytes)
	body := response.Body
	resBytes, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	var resp FeiShuHookResp
	if err = json.Unmarshal(resBytes, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return &resp
	}
	return err
}

func NewFeishuNotify() HookNotify {
	return &feishuNotify{}
}

func GenSign(secret string, timestamp int64) string {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret

	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return ""
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}
