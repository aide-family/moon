//go:build plugin

package main

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	
	"github.com/moon-monitor/moon/pkg/plugin"
	"github.com/moon-monitor/moon/pkg/plugin/sms"
)

type MockSender struct {
	helper *log.Helper
}

func (m *MockSender) Send(_ context.Context, phoneNumber string, message sms.Message) error {
	m.helper.Debugf("MockSender.Send called with number: %s, message: %+v", phoneNumber, message)
	return nil
}

func (m *MockSender) SendBatch(_ context.Context, phoneNumbers []string, message sms.Message) error {
	m.helper.Debugf("MockSender.SendBatch called with numbers: %v, message: %+v", phoneNumbers, message)
	return nil
}

// New is the exported plugin factory function
// Note: This must exactly match the expected signature in the main program
func New(config *plugin.LoadConfig) (sms.Sender, error) {
	return &MockSender{
		helper: log.NewHelper(log.With(config.Logger, "module", "plugin.sms.mock")),
	}, nil
}
