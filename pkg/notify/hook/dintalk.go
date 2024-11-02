package hook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/util/httpx"
)

var _ Notify = (*dingTalk)(nil)

// NewDingTalk 创建钉钉通知
func NewDingTalk(receiverHookDingTalk Config) Notify {
	return &dingTalk{
		c: receiverHookDingTalk,
	}
}

type dingTalk struct {
	c Config
}

func (l *dingTalk) Hash() string {
	return types.MD5(l.c.GetWebhook())
}

func (l *dingTalk) Type() string {
	return l.c.GetType()
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
		"sign":      l.generateSignature(timestamp, l.c.GetSecret()),
	})
	reqURL := fmt.Sprintf("%s&%s", l.c.GetWebhook(), params)
	msgStr := l.c.GetTemplate()
	content := l.c.GetContent()
	if content != "" {
		msgStr = content
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
	if err = types.Unmarshal(resBytes, &dingResp); err != nil {
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
