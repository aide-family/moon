package chatgroup

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.ChatGroupRepo = (*chatGroupRepoImpl)(nil)

var (
	// ErrNoCondition 不允许无条件操作DB
	ErrNoCondition = perrors.ErrorUnknown("no condition")
)

type chatGroupRepoImpl struct {
	repository.UnimplementedChatGroupRepo
	log *log.Helper
	d   *data.Data

	query.IAction[model.PromAlarmChatGroup]
}

func (l *chatGroupRepoImpl) Create(ctx context.Context, chatGroup *dobo.ChatGroupDO) (*dobo.ChatGroupDO, error) {
	newChatGroup := chatGroup.ToModel()
	if err := l.WithContext(ctx).Create(newChatGroup); err != nil {
		return nil, err
	}
	return dobo.ChatGroupModelToDO(newChatGroup), nil
}

func (l *chatGroupRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.ChatGroupDO, error) {
	chatGroupDetail, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}

	return dobo.ChatGroupModelToDO(chatGroupDetail), nil
}

func (l *chatGroupRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*dobo.ChatGroupDO, error) {
	var chatGroupList []*model.PromAlarmChatGroup
	if err := l.DB().WithContext(ctx).Scopes(scopes...).Find(&chatGroupList).Error; err != nil {
		return nil, err
	}
	list := slices.To(chatGroupList, func(i *model.PromAlarmChatGroup) *dobo.ChatGroupDO {
		return dobo.ChatGroupModelToDO(i)
	})
	return list, nil
}

func (l *chatGroupRepoImpl) Update(ctx context.Context, chatGroup *dobo.ChatGroupDO, scopes ...query.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	return l.WithContext(ctx).Update(chatGroup.ToModel(), scopes...)
}

func (l *chatGroupRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	return l.WithContext(ctx).Delete(scopes...)
}

func (l *chatGroupRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.ChatGroupDO, error) {
	chatGroupList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return slices.To(chatGroupList, func(i *model.PromAlarmChatGroup) *dobo.ChatGroupDO {
		return dobo.ChatGroupModelToDO(i)
	}), nil
}

func NewChatGroupRepo(d *data.Data, logger log.Logger) repository.ChatGroupRepo {
	return &chatGroupRepoImpl{
		log:     log.NewHelper(log.With(logger, "module", "data.repository.chat_group")),
		d:       d,
		IAction: query.NewAction[model.PromAlarmChatGroup](query.WithDB[model.PromAlarmChatGroup](d.DB())),
	}
}
