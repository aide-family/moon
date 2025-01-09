package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"gorm.io/gen"
)

// NewUserMessageRepository 创建用户消息实现
func NewUserMessageRepository(data *data.Data) repository.UserMessage {
	return &userMessageRepositoryImpl{data: data}
}

type userMessageRepositoryImpl struct {
	data *data.Data
}

// GetByID 根据ID获取用户消息
func (u *userMessageRepositoryImpl) GetByID(ctx context.Context, id uint32) (*model.SysUserMessage, error) {
	mainQuery := query.Use(u.data.GetMainDB(ctx))
	messageDo, err := mainQuery.WithContext(ctx).SysUserMessage.Where(
		mainQuery.SysUserMessage.ID.Eq(id),
		mainQuery.SysUserMessage.UserID.Eq(middleware.GetUserID(ctx)),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastUserMessageNotFound(ctx)
		}
		return nil, err
	}
	return messageDo, nil
}

// DeleteAll 删除所有用户消息
func (u *userMessageRepositoryImpl) DeleteAll(ctx context.Context) error {
	mainQuery := query.Use(u.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysUserMessage.Where(
		mainQuery.SysUserMessage.UserID.Eq(middleware.GetUserID(ctx)),
	).Delete()
	return err
}

// Create 创建用户消息
func (u *userMessageRepositoryImpl) Create(ctx context.Context, message *model.SysUserMessage) error {
	mainQuery := query.Use(u.data.GetMainDB(ctx))
	if err := mainQuery.WithContext(ctx).SysUserMessage.Create(message); err != nil {
		return err
	}
	client, ok := u.data.GetSSEClientManager().GetClient(message.UserID)
	if ok {
		go func() {
			if err := client.SendMessage(message.String()); err != nil {
				log.Errorw("method", "createUserMessage", "err", err)
			}
		}()
	}
	return nil
}

// Delete 删除用户消息
func (u *userMessageRepositoryImpl) Delete(ctx context.Context, ids []uint32) error {
	mainQuery := query.Use(u.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysUserMessage.Where(
		mainQuery.SysUserMessage.ID.In(ids...),
		mainQuery.SysUserMessage.UserID.Eq(middleware.GetUserID(ctx)),
	).Delete()
	return err
}

// List 分页查询用户消息
func (u *userMessageRepositoryImpl) List(ctx context.Context, params *bo.QueryUserMessageListParams) ([]*model.SysUserMessage, error) {
	mainQuery := query.Use(u.data.GetMainDB(ctx)).SysUserMessage
	userCtxQuery := mainQuery.WithContext(ctx)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, mainQuery.Content.Like(params.Keyword))
	}
	wheres = append(wheres, mainQuery.UserID.Eq(middleware.GetUserID(ctx)))
	userCtxQuery = userCtxQuery.Where(wheres...)
	var err error
	if userCtxQuery, err = types.WithPageQuery(userCtxQuery, params.Page); err != nil {
		return nil, err
	}
	return userCtxQuery.Order(mainQuery.ID.Desc()).Find()
}
