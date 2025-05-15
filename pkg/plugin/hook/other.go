package hook

import (
	"context"
	"io"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/httpx"
)

var _ Sender = (*otherHook)(nil)

func NewOtherHook(api string, opts ...OtherHookOption) Sender {
	h := &otherHook{
		api:    api,
		header: make(http.Header),
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.helper == nil {
		WithOtherLogger(log.DefaultLogger)(h)
	}
	return h
}

func WithOtherBasicAuth(username, password string) OtherHookOption {
	return func(h *otherHook) {
		h.basicAuth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

func WithOtherHeader(header map[string]string) OtherHookOption {
	return func(h *otherHook) {
		for k, v := range header {
			h.header.Set(k, v)
		}
	}
}

func WithOtherLogger(logger log.Logger) OtherHookOption {
	return func(h *otherHook) {
		h.helper = log.NewHelper(log.With(logger, "module", "plugin.hook.other"))
	}
}

type OtherHookOption func(*otherHook)

type otherHook struct {
	api       string
	basicAuth *BasicAuth
	header    http.Header
	helper    *log.Helper
}

func (o *otherHook) Type() common.HookAPP {
	return common.HookAPP_OTHER
}

// Send implements Hook.
func (o *otherHook) Send(ctx context.Context, message Message) (err error) {
	defer func() {
		if err != nil {
			o.helper.WithContext(ctx).Warnw("msg", "send other hook failed", "error", err, "req", string(message))
		}
	}()
	opts := []httpx.Option{
		httpx.WithHeader(o.header),
	}
	if o.basicAuth != nil {
		opts = append(opts, httpx.WithBasicAuth(o.basicAuth.Username, o.basicAuth.Password))
	}
	response, err := httpx.PostJsonWithOptions(ctx, o.api, message, opts...)
	if err != nil {
		o.helper.WithContext(ctx).Warnf("send other hook failed: %v", err)
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			o.helper.WithContext(ctx).Warnf("close other hook response body failed: %v", err)
		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		o.helper.WithContext(ctx).Warnf("read other hook response body failed: %v", err)
		return merr.ErrorBadRequest("read other hook response body failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		o.helper.WithContext(ctx).Warnf("send other hook failed: status code: %d, response: %s", response.StatusCode, string(body))
		return merr.ErrorBadRequest("status code: %d, response: %s", response.StatusCode, string(body))
	}

	return nil
}
