package other

import (
	"context"
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
	h := &hookImpl{
		api:    api,
		header: make(http.Header),
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.helper == nil {
		WithLogger(log.DefaultLogger)(h)
	}
	return h
}

func WithBasicAuth(username, password string) Option {
	return func(h *hookImpl) {
		h.basicAuth = &hook.BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

func WithHeader(header map[string]string) Option {
	return func(h *hookImpl) {
		for k, v := range header {
			h.header.Set(k, v)
		}
	}
}

func WithLogger(logger log.Logger) Option {
	return func(h *hookImpl) {
		h.helper = log.NewHelper(log.With(logger, "module", "plugin.hook.other"))
	}
}

type Option func(*hookImpl)

type hookImpl struct {
	api       string
	basicAuth *hook.BasicAuth
	header    http.Header
	helper    *log.Helper
}

func (o *hookImpl) Type() common.HookAPP {
	return common.HookAPP_OTHER
}

// Send implements Hook.
func (o *hookImpl) Send(ctx context.Context, message hook.Message) (err error) {
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
