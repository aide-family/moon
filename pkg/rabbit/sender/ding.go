package sender

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aide-family/moon/pkg/rabbit"
	"github.com/aide-family/moon/pkg/utils/httpx"
	"io"
	"time"
)

type Ding struct {
	url            string
	client         *httpx.HttpX
	secretProvider rabbit.SecretProvider
}

func NewDing(url string) (*Ding, error) {
	return &Ding{
		url:            url,
		client:         httpx.NewHttpX(),
		secretProvider: &DingSecretProvider{},
	}, nil
}

func (d *Ding) Name() string {
	return "ding"
}

func (d *Ding) Send(ctx context.Context, content []byte, secret []byte) error {
	sec := &DingSecret{}
	err := d.secretProvider.Provider(secret, sec)
	if err != nil {
		return err
	}
	params := map[string]any{
		"access_token": sec.Token,
	}
	if sec.Secret != "" {
		params["timestamp"] = sec.Timestamp
		params["sign"] = sec.Sign
	}
	reqUrl := fmt.Sprintf("%s&%s", d.url, httpx.ParseQuery(params))
	response, err := httpx.NewHttpX().POSTWithContext(ctx, reqUrl, content)
	if err != nil {
		return err
	}
	body := response.Body
	defer body.Close()
	resBytes, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	var dingResp dingTalkHookResp
	if err = json.Unmarshal(resBytes, &dingResp); err != nil {
		return err
	}
	if dingResp.ErrCode != 0 {
		return &dingResp
	}
	return err
}

type dingTalkHookResp struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *dingTalkHookResp) Error() string {
	return fmt.Sprintf("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

type DingSecretProvider struct {
}

type DingSecret struct {
	Token     string
	Secret    string
	Sign      string
	Timestamp int64
}

func (d *DingSecretProvider) Provider(in []byte, out any) error {
	_, ok := out.(*DingSecret)
	if !ok {
		return fmt.Errorf("convert failed: target secret structure is not DingSecret")
	}
	err := json.Unmarshal(in, out)
	if err != nil {
		return err
	}
	if out.(*DingSecret).Secret != "" {
		t := time.Now().UnixMilli()
		out.(*DingSecret).Sign = DingSign(t, out.(*DingSecret).Secret)
		out.(*DingSecret).Timestamp = t
	}

	return nil
}

func DingSign(t int64, secret string) string {
	strToHash := fmt.Sprintf("%d\n%s", t, secret)
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(strToHash))
	data := hmac256.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}
