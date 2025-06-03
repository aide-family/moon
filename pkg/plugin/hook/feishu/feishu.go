package feishu

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/hook"
	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/timex"
)

var _ hook.Sender = (*hookImpl)(nil)

func New(api, secret string, opts ...Option) hook.Sender {
	h := &hookImpl{api: api, secret: secret}
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
		h.helper = log.NewHelper(log.With(logger, "module", "plugin.hook.feishu"))
	}
}

type Option func(*hookImpl)

type hookImpl struct {
	api    string
	secret string
	helper *log.Helper
}

func (f *hookImpl) Type() common.HookAPP {
	return common.HookAPP_FEISHU
}

type hookResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (l *hookResp) Error() error {
	if l.Code == 0 {
		return nil
	}
	return merr.ErrorBadRequest("code: %d, msg: %s, data: %v", l.Code, l.Msg, l.Data)
}

func (f *hookImpl) Send(ctx context.Context, message hook.Message) (err error) {
	defer func() {
		if err != nil {
			f.helper.WithContext(ctx).Warnw("msg", "send feishu hook failed", "error", err, "req", string(message))
		}
	}()
	msg := make(map[string]any)
	if err := json.Unmarshal(message, &msg); err != nil {
		f.helper.WithContext(ctx).Warnf("unmarshal feishu hook message failed: %v", err)
		return err
	}
	requestTime := timex.Now().Unix()
	msg["timestamp"] = strconv.FormatInt(requestTime, 10)
	sign, err := f.sign(ctx, requestTime)
	if err != nil {
		f.helper.WithContext(ctx).Warnf("gen feishu hook sign failed: %v", err)
		return err
	}
	msg["sign"] = sign
	requestBody, err := json.Marshal(msg)
	if err != nil {
		f.helper.WithContext(ctx).Warnf("marshal feishu hook message failed: %v", err)
		return err
	}
	response, err := httpx.PostJson(ctx, f.api, requestBody)
	if err != nil {
		f.helper.WithContext(ctx).Warnf("send feishu hook failed: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			f.helper.WithContext(ctx).Warnf("close feishu hook response body failed: %v", err)
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		f.helper.WithContext(ctx).Warnf("send feishu hook failed: status code: %d", response.StatusCode)
		return merr.ErrorBadRequest("status code: %d", response.StatusCode)
	}

	var resp hookResp
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		f.helper.WithContext(ctx).Warnf("unmarshal feishu hook response failed: %v", err)
		body, _ := io.ReadAll(response.Body)
		return merr.ErrorBadRequest("unmarshal feishu hook response failed: %v, response: %s", err, string(body))
	}

	return resp.Error()
}

func (f *hookImpl) sign(ctx context.Context, timestamp int64) (string, error) {
	// timestamp + key sha256, then base64 encode
	signString := strconv.FormatInt(timestamp, 10) + "\n" + f.secret

	var data []byte
	h := hmac.New(sha256.New, []byte(signString))
	_, err := h.Write(data)
	if err != nil {
		f.helper.WithContext(ctx).Warnf("gen feishu hook sign failed: %v", err)
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
