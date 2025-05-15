package sms

import (
	"context"
)

type Message struct {
	TemplateParam string `json:"templateParam"`
	TemplateCode  string `json:"templateCode"`
}

type Sender interface {
	Send(ctx context.Context, phoneNumber string, message Message) error
	SendBatch(ctx context.Context, phoneNumbers []string, messages Message) error
}
