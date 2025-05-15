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
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/timex"
)

var _ Sender = (*dingTalkHook)(nil)

func NewDingTalkHook(api, secret string, opts ...DingTalkHookOption) Sender {
	d := &dingTalkHook{api: api, secret: secret}
	for _, opt := range opts {
		opt(d)
	}
	if d.helper == nil {
		WithDingTalkLogger(log.DefaultLogger)(d)
	}
	return d
}

type DingTalkHookOption func(*dingTalkHook)

type dingTalkHook struct {
	api    string
	secret string
	helper *log.Helper
}

func (d *dingTalkHook) Type() common.HookAPP {
	return common.HookAPP_DINGTALK
}

type dingTalkHookResp struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *dingTalkHookResp) Error() error {
	if l.ErrCode == 0 {
		return nil
	}
	return merr.ErrorBadRequest("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

func WithDingTalkLogger(logger log.Logger) DingTalkHookOption {
	return func(d *dingTalkHook) {
		d.helper = log.NewHelper(log.With(logger, "module", "plugin.hook.dingtalk"))
	}
}

func (d *dingTalkHook) Send(ctx context.Context, message Message) (err error) {
	defer func() {
		if err != nil {
			d.helper.WithContext(ctx).Warnw("msg", "send dingtalk hook failed", "error", err, "req", string(message))
		}
	}()
	timestamp := timex.Now().UnixMilli()
	sign := d.sign(timestamp)
	query := d.parseQuery(map[string]any{
		"timestamp": timestamp,
		"sign":      sign,
	})

	api := d.parseApi(d.api, query)
	response, err := httpx.PostJson(ctx, api, message)
	if err != nil {
		d.helper.WithContext(ctx).Warnf("send dingtalk hook failed: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			d.helper.WithContext(ctx).Warnf("close dingtalk hook response body failed: %v", err)
		}
	}(response.Body)

	var resp dingTalkHookResp
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		d.helper.WithContext(ctx).Warnf("unmarshal dingtalk hook response failed: %v", err)
		body, _ := io.ReadAll(response.Body)
		return merr.ErrorBadRequest("unmarshal dingtalk hook response failed: %v, response: %s", err, string(body))
	}

	return resp.Error()
}

// parseApi parse api and query
func (d *dingTalkHook) parseApi(api string, query string) string {
	if strings.HasSuffix(api, "?") {
		return api + query
	}

	if strings.HasSuffix(query, "&") {
		return api + query
	}

	if strings.Contains(query, "&") || strings.Contains(query, "?") {
		return api + "&" + query
	}
	return api + "?" + query
}

// parseQuery parse struct to query params
func (d *dingTalkHook) parseQuery(qr map[string]any) string {
	if len(qr) == 0 {
		return ""
	}
	query := url.Values{}
	for k, v := range qr {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	return query.Encode()
}

func (d *dingTalkHook) sign(timestamp int64) string {
	message := fmt.Sprintf("%d\n%s", timestamp, d.secret)
	key := []byte(d.secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	signatureURLEncoded := url.QueryEscape(signatureBase64)
	return signatureURLEncoded
}
