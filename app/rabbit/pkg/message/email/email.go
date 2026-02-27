// Package email is a simple package that provides a email sender.
package email

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/gomail.v2"

	"github.com/aide-family/rabbit/pkg/message"
)

var _ message.Sender = (*emailSender)(nil)

func init() {
	message.RegisterDriver(enum.MessageType_EMAIL, New)
}

func New(c *config.MessageConfig) (message.Sender, error) {
	options := &config.MessageEmailConfig{}
	if err := anypb.UnmarshalTo(c.GetOptions(), options, proto.UnmarshalOptions{Merge: true}); err != nil {
		return nil, merr.ErrorInternalServer("unmarshal email config failed: %v", err)
	}
	host, port := options.GetHost(), options.GetPort()
	username, password := options.GetUsername(), options.GetPassword()
	dialer := gomail.NewDialer(host, int(port), username, password)
	return &emailSender{dialer: dialer, config: options}, nil
}

type emailSender struct {
	dialer *gomail.Dialer
	config *config.MessageEmailConfig
}

func (e *emailSender) Send(ctx context.Context, m message.Message) error {
	if m.Type() != enum.MessageType_EMAIL {
		return fmt.Errorf("message type %s not supported, only %s is supported", m.Type(), enum.MessageType_EMAIL)
	}
	emailMessage := &Message{}
	var ok bool
	if emailMessage, ok = m.(*Message); !ok {
		jsonBytes, err := m.Marshal()
		if err != nil {
			return err
		}
		emailMessage = NewMessage()
		if err := json.Unmarshal(jsonBytes, emailMessage); err != nil {
			return err
		}
	}
	msg := gomail.NewMessage(gomail.SetCharset("UTF-8"), gomail.SetEncoding(gomail.Base64))
	msg.SetHeader("From", e.config.GetUsername())
	msg.SetHeader("To", emailMessage.To...)
	msg.SetHeader("Cc", emailMessage.Cc...)
	msg.SetHeader("Subject", emailMessage.Subject)
	msg.SetBody(emailMessage.ContentType, emailMessage.Body)
	for _, attachment := range emailMessage.Attachments {
		msg.Attach(attachment.Filename, gomail.SetHeader(map[string][]string{
			"Content-Disposition": {"attachment"},
		}), gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(attachment.Data)
			return err
		}))
	}
	for key, values := range emailMessage.Headers {
		msg.SetHeader(key, values...)
	}
	return e.dialer.DialAndSend(msg)
}
