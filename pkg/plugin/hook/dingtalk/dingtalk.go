package dingtalk

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
	"github.com/aide-family/moon/pkg/plugin/hook"
	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/timex"
)

var _ hook.Sender = (*hookImpl)(nil)

func New(api, secret string, opts ...Option) hook.Sender {
	d := &hookImpl{api: api, secret: secret}
	for _, opt := range opts {
		opt(d)
	}
	if d.helper == nil {
		WithLogger(log.DefaultLogger)(d)
	}
	return d
}

type Option func(*hookImpl)

type hookImpl struct {
	api    string
	secret string
	helper *log.Helper
}

func (d *hookImpl) Type() common.HookAPP {
	return common.HookAPP_DINGTALK
}

type hookResp struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *hookResp) Error() error {
	if l.ErrCode == 0 {
		return nil
	}
	return merr.ErrorBadRequest("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

func WithLogger(logger log.Logger) Option {
	return func(d *hookImpl) {
		d.helper = log.NewHelper(log.With(logger, "module", "plugin.hook.dingtalk"))
	}
}

func (d *hookImpl) Send(ctx context.Context, message hook.Message) (err error) {
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

	var resp hookResp
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		d.helper.WithContext(ctx).Warnf("unmarshal dingtalk hook response failed: %v", err)
		body, _ := io.ReadAll(response.Body)
		return merr.ErrorBadRequest("unmarshal dingtalk hook response failed: %v, response: %s", err, string(body))
	}

	return resp.Error()
}

// parseApi parse api and query
func (d *hookImpl) parseApi(api string, query string) string {
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
func (d *hookImpl) parseQuery(qr map[string]any) string {
	if len(qr) == 0 {
		return ""
	}
	query := url.Values{}
	for k, v := range qr {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	return query.Encode()
}

func (d *hookImpl) sign(timestamp int64) string {
	message := fmt.Sprintf("%d\n%s", timestamp, d.secret)
	key := []byte(d.secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	signatureURLEncoded := url.QueryEscape(signatureBase64)
	return signatureURLEncoded
}
