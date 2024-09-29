package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen"
)

// NewUserMessageRepository creates a new UserMessageRepository instance.
func NewUserMessageRepository(data *data.Data) repository.UserMessage {
	return &userMessageRepositoryImpl{data: data}
}

type userMessageRepositoryImpl struct {
	data *data.Data
}

func (u *userMessageRepositoryImpl) DeleteAll(ctx context.Context) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return merr.ErrorI18nUnauthorized(ctx)
	}
	mainQuery := query.Use(u.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysUserMessage.Where(
		mainQuery.SysUserMessage.UserID.Eq(claims.GetUser()),
	).Delete()
	return err
}

func (u *userMessageRepositoryImpl) Create(ctx context.Context, message *model.SysUserMessage) error {
	mainQuery := query.Use(u.data.GetMainDB(ctx))
	return mainQuery.WithContext(ctx).SysUserMessage.Create(message)
}

func (u *userMessageRepositoryImpl) Delete(ctx context.Context, uint32s []uint32) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return merr.ErrorI18nUnauthorized(ctx)
	}
	mainQuery := query.Use(u.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysUserMessage.Where(
		mainQuery.SysUserMessage.ID.In(uint32s...),
		mainQuery.SysUserMessage.UserID.Eq(claims.GetUser()),
	).Delete()
	return err
}

func (u *userMessageRepositoryImpl) List(ctx context.Context, params *bo.QueryUserMessageListParams) ([]*model.SysUserMessage, error) {
	mainQuery := query.Use(u.data.GetMainDB(ctx)).SysUserMessage
	userCtxQuery := mainQuery.WithContext(ctx)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, mainQuery.Content.Like(params.Keyword))
	}
	userCtxQuery = userCtxQuery.Where(wheres...)
	if err := types.WithPageQuery[query.ISysUserMessageDo](userCtxQuery, params.Page); err != nil {
		return nil, err
	}
	return userCtxQuery.Order(mainQuery.ID.Desc()).Find()
}
