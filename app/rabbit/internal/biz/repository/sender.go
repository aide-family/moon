package repository

import (
	"context"

	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

// MessageSender 定义消息发送器接口
type MessageSender interface {
	// Send 发送消息
	Send(ctx context.Context, messageLog *bo.MessageLogItemBo) error
	// Type 返回发送器支持的消息类型
	Type() enum.MessageType
}
