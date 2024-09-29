package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
)

// UserMessage 用户消息
type UserMessage interface {
	Create(context.Context, *model.SysUserMessage) error

	Delete(context.Context, []uint32) error
	DeleteAll(context.Context) error

	List(context.Context, *bo.QueryUserMessageListParams) ([]*model.SysUserMessage, error)
}
