package ali

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/plugin/sms"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapiV3 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/pointer"
)

var _ sms.Sender = (*aliyun)(nil)

func NewAliyun(c Config, opts ...AliyunOption) (sms.Sender, error) {
	a := &aliyun{
		accessKeyID:     c.GetAccessKeyId(),
		accessKeySecret: c.GetAccessKeySecret(),
		signName:        c.GetSignName(),
		endpoint:        c.GetEndpoint(),
		clientV3:        nil,
		helper:          nil,
	}
	for _, opt := range opts {
		opt(a)
	}
	if a.helper == nil {
		WithAliyunLogger(log.DefaultLogger)(a)
	}
	var err error
	a.clientV3, err = a.initV3()
	if err != nil {
		return nil, err
	}
	return a, nil
}

type Config interface {
	GetAccessKeySecret() string
	GetAccessKeyId() string
	GetSignName() string
	GetEndpoint() string
}

type AliyunOption func(*aliyun)

type aliyun struct {
	accessKeyID     string
	accessKeySecret string
	signName        string
	endpoint        string

	clientV3 *dysmsapiV3.Client
	helper   *log.Helper
}

// initV3 initializes the SMS clientV3
func (a *aliyun) initV3() (*dysmsapiV3.Client, error) {
	if a.accessKeySecret == "" || a.accessKeyID == "" {
		return nil, merr.ErrorBadRequest("SMS sending credential information is not configured")
	}
	config := &openapi.Config{
		AccessKeyId:     &a.accessKeyID,
		AccessKeySecret: &a.accessKeySecret,
	}
	if a.endpoint != "" {
		config.Endpoint = tea.String(a.endpoint)
	}
	client, err := dysmsapiV3.NewClient(config)
	if err != nil {
		return nil, merr.ErrorBadRequest("Failed to initialize SMS clientV3").WithCause(err)
	}
	return client, nil
}

func (a *aliyun) Send(ctx context.Context, phoneNumber string, message sms.Message) error {
	sendSmsRequest := &dysmsapiV3.SendSmsRequest{
		PhoneNumbers:  pointer.Of(phoneNumber),
		SignName:      pointer.Of(a.signName),
		TemplateCode:  pointer.Of(message.TemplateCode),
		TemplateParam: pointer.Of(message.TemplateParam),
	}

	response, err := a.clientV3.SendSmsWithOptions(sendSmsRequest, runtimeOptions)
	if err != nil {
		a.helper.WithContext(ctx).Debugf("send sms failed: %v", err)
		return err
	}
	a.helper.WithContext(ctx).Debugf("send sms response: %v", response)
	if pointer.Get(response.Body.Code) != "OK" {
		a.helper.WithContext(ctx).Errorf("send sms failed: %v", response)
		body := pointer.Get(response.Body)
		return merr.ErrorBadRequest("send sms failed: %v", body)
	}
	return nil
}

func (a *aliyun) SendBatch(ctx context.Context, phoneNumbers []string, message sms.Message) error {
	signNames := make([]string, 0, len(phoneNumbers))
	templateParams := make([]string, 0, len(phoneNumbers))
	for range phoneNumbers {
		signNames = append(signNames, a.signName)
		templateParams = append(templateParams, message.TemplateParam)
	}

	phoneNumberJson, err := json.Marshal(phoneNumbers)
	if err != nil {
		a.helper.WithContext(ctx).Errorf("Failed to marshal phone numbers: %v", err)
		return merr.ErrorBadRequest("Failed to marshal phone numbers").WithCause(err)
	}
	signNameJson, err := json.Marshal(signNames)
	if err != nil {
		a.helper.WithContext(ctx).Errorf("Failed to marshal sign names: %v", err)
		return merr.ErrorBadRequest("Failed to marshal sign names").WithCause(err)
	}
	templateParamJson, err := json.Marshal(templateParams)
	if err != nil {
		a.helper.WithContext(ctx).Errorf("Failed to marshal template params: %v", err)
		return merr.ErrorBadRequest("Failed to marshal template params").WithCause(err)
	}
	sendBatchSmsRequest := &dysmsapiV3.SendBatchSmsRequest{
		PhoneNumberJson:   pointer.Of(string(phoneNumberJson)),
		SignNameJson:      pointer.Of(string(signNameJson)),
		TemplateParamJson: pointer.Of(string(templateParamJson)),
		TemplateCode:      pointer.Of(message.TemplateCode),
	}

	response, err := a.clientV3.SendBatchSmsWithOptions(sendBatchSmsRequest, runtimeOptions)
	if err != nil {
		a.helper.WithContext(ctx).Errorf("send batch sms failed: %v", err)
		return err
	}
	a.helper.WithContext(ctx).Debugf("send batch sms response: %v", response)
	if pointer.Get(response.Body.Code) != "OK" {
		a.helper.WithContext(ctx).Errorf("send batch sms failed: %v", response)
		body := pointer.Get(response.Body)
		return merr.ErrorBadRequest("send batch sms failed: %v", body)
	}
	return nil
}
