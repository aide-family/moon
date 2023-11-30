package biz

import (
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type (
	ChatGroupBiz struct {
		log *log.Helper

		chatGroupRepo repository.ChatGroupRepo
	}
)

// NewChatGroupBiz .
func NewChatGroupBiz(chatGroupRepo repository.ChatGroupRepo, logger log.Logger) *ChatGroupBiz {
	return &ChatGroupBiz{
		log:           log.NewHelper(logger),
		chatGroupRepo: chatGroupRepo,
	}
}
