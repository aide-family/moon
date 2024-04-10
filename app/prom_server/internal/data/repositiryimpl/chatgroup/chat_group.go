package chatgroup

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/pkg/helper/middler"

	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ repository.ChatGroupRepo = (*chatGroupRepoImpl)(nil)

var (
	// ErrNoCondition 不允许无条件操作DB
	ErrNoCondition = perrors.ErrorUnknown("no condition")
)

type chatGroupRepoImpl struct {
	repository.UnimplementedChatGroupRepo
	log  *log.Helper
	data *data.Data
}

func (l *chatGroupRepoImpl) Create(ctx context.Context, chatGroup *bo.ChatGroupBO) (*bo.ChatGroupBO, error) {
	newChatGroup := chatGroup.ToModel()
	newChatGroup.CreateBy = middler.GetUserId(ctx)
	if err := l.data.DB().WithContext(ctx).Create(newChatGroup).Error; err != nil {
		return nil, err
	}
	return bo.ChatGroupModelToBO(newChatGroup), nil
}

func (l *chatGroupRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.ChatGroupBO, error) {
	var chatGroupDetail do.PromAlarmChatGroup
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(whereList...).First(&chatGroupDetail).Error; err != nil {
		return nil, err
	}

	return bo.ChatGroupModelToBO(&chatGroupDetail), nil
}

func (l *chatGroupRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.ChatGroupBO, error) {
	var chatGroupList []*do.PromAlarmChatGroup
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(whereList...).Find(&chatGroupList).Error; err != nil {
		return nil, err
	}
	list := slices.To(chatGroupList, func(i *do.PromAlarmChatGroup) *bo.ChatGroupBO {
		return bo.ChatGroupModelToBO(i)
	})
	return list, nil
}

func (l *chatGroupRepoImpl) Update(ctx context.Context, chatGroup *bo.ChatGroupBO, scopes ...basescopes.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	return l.data.DB().WithContext(ctx).Scopes(whereList...).Updates(chatGroup.ToModel()).Error
}

func (l *chatGroupRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	return l.data.DB().WithContext(ctx).Scopes(whereList...).Delete(&do.PromAlarmChatGroup{}).Error
}

func (l *chatGroupRepoImpl) List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.ChatGroupBO, error) {
	var chatGroupList []*do.PromAlarmChatGroup
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(append(whereList, bo.Page(pgInfo))...).Find(&chatGroupList).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Scopes(whereList...).Model(&do.PromAlarmChatGroup{}).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	return slices.To(chatGroupList, func(i *do.PromAlarmChatGroup) *bo.ChatGroupBO {
		return bo.ChatGroupModelToBO(i)
	}), nil
}

func NewChatGroupRepo(d *data.Data, logger log.Logger) repository.ChatGroupRepo {
	return &chatGroupRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.repository.chat_group")),
		data: d,
	}
}
