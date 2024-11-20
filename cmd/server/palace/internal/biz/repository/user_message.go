package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
)

// UserMessage 用户消息
type UserMessage interface {
	// Create 创建用户消息
	Create(context.Context, *model.SysUserMessage) error

	// Delete 删除用户消息
	Delete(context.Context, []uint32) error

	// DeleteAll 删除所有用户消息
	DeleteAll(context.Context) error

	// List 分页查询用户消息
	List(context.Context, *bo.QueryUserMessageListParams) ([]*model.SysUserMessage, error)

	// GetByID 根据ID获取用户消息
	GetByID(context.Context, uint32) (*model.SysUserMessage, error)
}
