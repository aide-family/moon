package wechat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/hook"
	"github.com/aide-family/moon/pkg/util/httpx"
)

var _ hook.Sender = (*hookImpl)(nil)

func New(api string, opts ...Option) hook.Sender {
	h := &hookImpl{api: api}
	for _, opt := range opts {
		opt(h)
	}
	if h.helper == nil {
		WithLogger(log.DefaultLogger)(h)
	}
	return h
}

func WithLogger(logger log.Logger) Option {
	return func(h *hookImpl) {
		h.helper = log.NewHelper(log.With(logger, "module", "plugin.hook.wechat"))
	}
}

type Option func(*hookImpl)

type hookImpl struct {
	api    string
	helper *log.Helper
}

func (h *hookImpl) Type() common.HookAPP {
	return common.HookAPP_WECHAT
}

type hookResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *hookResp) Error() error {
	if l.ErrCode == 0 {
		return nil
	}
	return merr.ErrorBadRequest("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

func (h *hookImpl) Send(ctx context.Context, message hook.Message) (err error) {
	defer func() {
		if err != nil {
			h.helper.WithContext(ctx).Warnw("msg", "send wechat hook failed", "error", err, "req", string(message))
		}
	}()
	response, err := httpx.PostJson(ctx, h.api, message)
	if err != nil {
		h.helper.WithContext(ctx).Warnf("send wechat hook failed: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			h.helper.WithContext(ctx).Warnf("close wechat hook response body failed: %v", err)
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		h.helper.WithContext(ctx).Warnf("send wechat hook failed: status code: %d", response.StatusCode)
		return merr.ErrorBadRequest("status code: %d", response.StatusCode)
	}

	var resp hookResp
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		h.helper.WithContext(ctx).Warnf("unmarshal wechat hook response failed: %v", err)
		body, _ := io.ReadAll(response.Body)
		return merr.ErrorBadRequest("unmarshal wechat hook response failed: %v, response: %s", err, string(body))
	}

	return resp.Error()
}
