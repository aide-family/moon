package hook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/util/httpx"
)

var _ Notify = (*feiShu)(nil)

// NewFeiShu 创建飞书通知
func NewFeiShu(config Config) Notify {
	return &feiShu{
		c: config,
	}
}

type feiShu struct {
	c Config
}

func (l *feiShu) Hash() string {
	return types.MD5(l.c.GetWebhook())
}

func (l *feiShu) Type() string {
	return l.c.GetType()
}

type feiShuHookResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (l *feiShuHookResp) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", l.Code, l.Msg)
}

func (l *feiShu) Send(ctx context.Context, msg notify.Msg) error {
	notifyMsg := make(notify.Msg)
	msgStr := l.c.GetTemplate()
	content := l.c.GetContent()
	if content != "" {
		msgStr = content
	}
	msgStr = format.Formatter(msgStr, msg)
	if err := types.Unmarshal([]byte(msgStr), &notifyMsg); err != nil {
		return err
	}

	timeNow := time.Now()
	notifyMsg["timestamp"] = strconv.FormatInt(timeNow.Unix(), 10)
	notifyMsg["sign"] = genSign(l.c.GetSecret(), timeNow.Unix())
	notifyMsgBytes, _ := types.Marshal(notifyMsg)
	response, err := httpx.NewHTTPX().POSTWithContext(ctx, l.c.GetWebhook(), notifyMsgBytes)
	if err != nil {
		return err
	}
	body := response.Body
	defer body.Close()
	resBytes, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	var resp feiShuHookResp
	if err = types.Unmarshal(resBytes, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return &resp
	}
	return err
}

func genSign(secret string, timestamp int64) string {
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
