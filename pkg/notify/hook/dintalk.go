package hook

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

	"github.com/aide-family/moon/pkg/util/format"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/util/httpx"
)

var _ Notify = (*dingTalk)(nil)

// NewDingTalk 创建钉钉通知
func NewDingTalk(receiverHookDingTalk *api.ReceiverHookDingTalk) Notify {
	return &dingTalk{
		ReceiverHookDingTalk: receiverHookDingTalk,
	}
}

type dingTalk struct {
	*api.ReceiverHookDingTalk
}

type dingTalkHookResp struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *dingTalkHookResp) Error() string {
	return fmt.Sprintf("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

func (l *dingTalk) Send(ctx context.Context, msg notify.Msg) error {
	timestamp := time.Now().UnixMilli()
	params := httpx.ParseQuery(map[string]any{
		"timestamp": timestamp,
		"sign":      l.generateSignature(timestamp, l.GetSecret()),
	})
	reqURL := fmt.Sprintf("%s&%s", l.GetWebhook(), params)
	temp := l.GetTemplate()
	msgStr := l.GetContent()
	if temp != "" {
		msgStr = temp
	}
	msgStr = format.Formatter(msgStr, msg)
	response, err := httpx.NewHTTPX().POSTWithContext(ctx, reqURL, []byte(msgStr))
	if err != nil {
		return err
	}
	body := response.Body
	defer body.Close()
	resBytes, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes), "req", msgStr)
	var dingResp dingTalkHookResp
	if err = json.Unmarshal(resBytes, &dingResp); err != nil {
		return err
	}
	if dingResp.ErrCode != 0 {
		return &dingResp
	}
	return err
}

func (l *dingTalk) generateSignature(timestamp int64, secret string) string {
	message := fmt.Sprintf("%d\n%s", timestamp, secret)
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	signatureURLEncoded := url.QueryEscape(signatureBase64)
	return signatureURLEncoded
}
