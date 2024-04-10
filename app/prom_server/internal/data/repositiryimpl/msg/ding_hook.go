package msg

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/pkg/httpx"
)

var _ HookNotify = (*dingNotify)(nil)

type dingNotify struct{}

type DingDingHookResp struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *DingDingHookResp) Error() string {
	return fmt.Sprintf("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

func (l *dingNotify) Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error {
	timestamp := time.Now().UnixMilli()
	params := httpx.ParseQuery(map[string]any{
		"timestamp": timestamp,
		"sign":      l.generateSignature(timestamp, msg.Secret),
	})
	reqUrl := fmt.Sprintf("%s&%s", url, params)
	response, err := httpx.NewHttpX().POSTWithContext(ctx, reqUrl, []byte(msg.Content))
	body := response.Body
	resBytes, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	var dingResp DingDingHookResp
	if err = json.Unmarshal(resBytes, &dingResp); err != nil {
		return err
	}
	if dingResp.ErrCode != 0 {
		return &dingResp
	}
	return err
}

func (l *dingNotify) generateSignature(timestamp int64, secret string) string {
	message := fmt.Sprintf("%d\n%s", timestamp, secret)
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	signatureURLEncoded := url.QueryEscape(signatureBase64)
	return signatureURLEncoded
}

func NewDingNotify() HookNotify {
	return &dingNotify{}
}
