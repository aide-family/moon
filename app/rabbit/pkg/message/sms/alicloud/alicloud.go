// Package alicloud implements the alicloud sms driver.
package alicloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapiv3 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/rabbit/pkg/message"
)

var _ message.Sender = (*alicloudSmsSender)(nil)

func init() {
	message.RegisterDriver(enum.MessageType_SMS_ALICLOUD, New)
}

func New(c *config.MessageConfig) (message.Sender, error) {
	options := &config.MessageSMSAliCloudConfig{}
	if err := anypb.UnmarshalTo(c.GetOptions(), options, proto.UnmarshalOptions{Merge: true}); err != nil {
		return nil, merr.ErrorInternalServer("unmarshal sms ali cloud config failed: %v", err)
	}
	sender := &alicloudSmsSender{c: options}
	clientV3, err := sender.initV3()
	if err != nil {
		return nil, merr.ErrorInternalServer("init sms ali cloud client failed: %v", err)
	}
	return &alicloudSmsSender{c: options, clientV3: clientV3}, nil
}

type alicloudSmsSender struct {
	c        *config.MessageSMSAliCloudConfig
	clientV3 *dysmsapiv3.Client
}

func (a *alicloudSmsSender) initV3() (*dysmsapiv3.Client, error) {
	accessKeyID := a.c.GetAccessKeyID()
	accessKeySecret := a.c.GetAccessKeySecret()
	if accessKeyID == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("SMS sending credential information is not configured")
	}
	config := &openapi.Config{
		AccessKeyId:     &accessKeyID,
		AccessKeySecret: &accessKeySecret,
	}

	if endpoint := a.c.GetEndpoint(); endpoint != "" {
		config.Endpoint = tea.String(endpoint)
	}
	client, err := dysmsapiv3.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SMS clientV3: %w", err)
	}
	return client, nil
}

// Send implements message.Sender.
func (a *alicloudSmsSender) Send(ctx context.Context, msg message.Message) error {
	if msg.Type() != enum.MessageType_SMS_ALICLOUD {
		return fmt.Errorf("message type %s not supported, only %s is supported", msg.Type(), enum.MessageType_SMS_ALICLOUD)
	}
	newMessage := &Message{}
	var ok bool
	if newMessage, ok = msg.(*Message); !ok {
		jsonBytes, err := msg.Marshal()
		if err != nil {
			return err
		}
		newMessage = &Message{}
		if err := json.Unmarshal(jsonBytes, newMessage); err != nil {
			return err
		}
	}
	if len(newMessage.PhoneNumbers) == 0 {
		return fmt.Errorf("phone numbers is empty")
	}

	if len(newMessage.PhoneNumbers) == 1 {
		return a.send(ctx, newMessage)
	}
	return a.batchSend(ctx, newMessage)
}

func (a *alicloudSmsSender) send(_ context.Context, message *Message) error {
	phoneNumber := message.PhoneNumbers[0]
	sendSmsRequest := &dysmsapiv3.SendSmsRequest{
		PhoneNumbers:  pointer.Of(phoneNumber),
		SignName:      pointer.Of(a.c.GetSignName()),
		TemplateCode:  pointer.Of(message.TemplateCode),
		TemplateParam: pointer.Of(message.TemplateParam),
	}

	response, err := a.clientV3.SendSmsWithOptions(sendSmsRequest, runtimeOptions)
	if err != nil {
		return err
	}
	body := pointer.Get(response.Body)
	if pointer.Get(body.Code) == "OK" {
		return nil
	}

	return fmt.Errorf("send sms failed: %v", body)
}

func (a *alicloudSmsSender) batchSend(_ context.Context, message *Message) error {
	phoneNumbers := message.PhoneNumbers
	signNames := make([]string, 0, len(phoneNumbers))
	templateParams := make([]string, 0, len(phoneNumbers))
	for range phoneNumbers {
		signNames = append(signNames, a.c.GetSignName())
		templateParams = append(templateParams, message.TemplateParam)
	}

	phoneNumberJSON, err := json.Marshal(phoneNumbers)
	if err != nil {
		return fmt.Errorf("failed to marshal phone numbers: %v", err)
	}
	signNameJSON, err := json.Marshal(signNames)
	if err != nil {
		return fmt.Errorf("failed to marshal sign names: %v", err)
	}
	templateParamJSON, err := json.Marshal(templateParams)
	if err != nil {
		return fmt.Errorf("failed to marshal template params: %v", err)
	}
	sendBatchSmsRequest := &dysmsapiv3.SendBatchSmsRequest{
		PhoneNumberJson:   pointer.Of(string(phoneNumberJSON)),
		SignNameJson:      pointer.Of(string(signNameJSON)),
		TemplateParamJson: pointer.Of(string(templateParamJSON)),
		TemplateCode:      pointer.Of(message.TemplateCode),
	}

	response, err := a.clientV3.SendBatchSmsWithOptions(sendBatchSmsRequest, runtimeOptions)
	if err != nil {
		return err
	}
	body := pointer.Get(response.Body)
	if pointer.Get(body.Code) == "OK" {
		return nil
	}
	return fmt.Errorf("send batch sms failed: %v", body)
}
