package biz

import (
	"context"
	"slices"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewMessage(
	messageLogRepo repository.MessageLog,
	messageRepo repository.Message,
	helper *klog.Helper,
) *Message {
	return &Message{
		messageLogRepo: messageLogRepo,
		messageRepo:    messageRepo,
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "message")),
	}
}

type Message struct {
	helper         *klog.Helper
	messageLogRepo repository.MessageLog
	messageRepo    repository.Message
}

func (m *Message) SendMessage(ctx context.Context, uid snowflake.ID) error {
	messageLog, err := m.messageLogRepo.GetMessageLog(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return err
		}
		m.helper.Errorw("msg", "get message log failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("get message log failed").WithCause(err)
	}
	if slices.Contains([]enum.MessageStatus{enum.MessageStatus_SENT, enum.MessageStatus_SENDING, enum.MessageStatus_CANCELLED}, messageLog.Status) {
		m.helper.Warnw("msg", "message already sent or sending or cancelled", "uid", uid, "status", messageLog.Status)
		return nil
	}
	return m.messageRepo.AppendMessage(ctx, uid)
}
