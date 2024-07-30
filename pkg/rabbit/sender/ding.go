package sender

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aide-family/moon/pkg/rabbit"
	"github.com/aide-family/moon/pkg/util/httpx"
)

// Ding 钉钉消息发送
type Ding struct {
	url            string
	conf           *DingConfig
	client         *httpx.HTTPX
	configProvider rabbit.ConfigProvider
}

// NewDing 创建钉钉消息发送
func NewDing(url string) (*Ding, error) {
	return &Ding{
		url:            url,
		client:         httpx.NewHTTPX(),
		configProvider: &DingConfigProvider{},
	}, nil
}

// Name 获取名称
func (d *Ding) Name() string {
	return "ding"
}

func (d *Ding) Inject(data rabbit.Rule) (rabbit.Sender, error) {
	conf := &DingConfig{}
	sendRule := data.(*rabbit.SendRuleBuilder)
	config := sendRule.Config["config"]
	err := d.configProvider.Provider([]byte(config), conf)
	if err != nil {
		return nil, err
	}
	clone := &Ding{
		url:            d.url,
		conf:           conf,
		client:         d.client,
		configProvider: d.configProvider,
	}
	return clone, nil
}

// Send 发送消息
func (d *Ding) Send(ctx context.Context, content []byte) error {
	params := map[string]any{
		"access_token": d.conf.Token,
	}
	if d.conf.Secret != "" {
		params["timestamp"] = d.conf.Timestamp
		params["sign"] = d.conf.Sign
	}
	reqUrl := fmt.Sprintf("%s&%s", d.url, httpx.ParseQuery(params))
	response, err := httpx.NewHTTPX().POSTWithContext(ctx, reqUrl, content)
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

// DingConfigProvider 钉钉配置提供者
type DingConfigProvider struct {
}

type DingConfig struct {
	Token     string `json:"token" yaml:"token"`
	Secret    string `json:"secret" yaml:"secret"`
	Sign      string `json:"sign" yaml:"sign"`
	Timestamp int64  `json:"timestamp" yaml:"timestamp"`
}

// Provider 配置提供者
func (d *DingConfigProvider) Provider(in []byte, out any) error {
	_, ok := out.(*DingConfig)
	if !ok {
		return fmt.Errorf("convert failed: target secret structure is not DingConfig")
	}
	err := (&JsonProvider{}).Provider(in, out)
	if err != nil {
		return err
	}
	if out.(*DingConfig).Secret != "" {
		t := time.Now().UnixMilli()
		out.(*DingConfig).Sign = DingSign(t, out.(*DingConfig).Secret)
		out.(*DingConfig).Timestamp = t
	}
	return nil
}

// DingSign 钉钉签名
func DingSign(t int64, secret string) string {
	strToHash := fmt.Sprintf("%d\n%s", t, secret)
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(strToHash))
	data := hmac256.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}
