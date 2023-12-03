package user

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/system"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.UserRepo = (*userRepoImpl)(nil)

type userRepoImpl struct {
	repository.UnimplementedUserRepo
	log  *log.Helper
	data *data.Data
	query.IAction[model.SysUser]
}

func (l *userRepoImpl) RelateRoles(ctx context.Context, userBO *bo.UserBO, roleList []*bo.RoleBO) error {
	roleModelList := slices.To(roleList, func(roleInfo *bo.RoleBO) *model.SysRole {
		return roleInfo.ToModel()
	})

	return l.DB().WithContext(ctx).Model(userBO.ToModel()).
		Association(string(system.UserAssociationReplaceRoles)).
		Replace(&roleModelList)

}

func (l *userRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.UserBO, error) {
	userDetail, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}
	return bo.UserModelToBO(userDetail), nil
}

func (l *userRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.UserBO, error) {
	var userDetailList []*model.SysUser
	if err := l.DB().WithContext(ctx).Scopes(scopes...).Find(&userDetailList).Error; err != nil {
		return nil, err
	}
	list := slices.To(userDetailList, func(user *model.SysUser) *bo.UserBO {
		return bo.UserModelToBO(user)
	})
	return list, nil
}

func (l *userRepoImpl) Count(ctx context.Context, scopes ...query.ScopeMethod) (int64, error) {
	var count int64
	if err := l.DB().WithContext(ctx).Scopes(scopes...).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (l *userRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.UserBO, error) {
	userList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := slices.To(userList, func(user *model.SysUser) *bo.UserBO {
		return bo.UserModelToBO(user)
	})
	return list, nil
}

func (l *userRepoImpl) Create(ctx context.Context, user *bo.UserBO) (*bo.UserBO, error) {
	newUser := user.ToModel()
	if err := l.WithContext(ctx).Create(newUser); err != nil {
		return nil, err
	}
	return bo.UserModelToBO(newUser), nil
}

func (l *userRepoImpl) Update(ctx context.Context, user *bo.UserBO, scopes ...query.ScopeMethod) (*bo.UserBO, error) {
	newUser := user.ToModel()
	if err := l.WithContext(ctx).Update(newUser, scopes...); err != nil {
		return nil, err
	}
	return bo.UserModelToBO(newUser), nil
}

func (l *userRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	return l.WithContext(ctx).Delete(scopes...)
}

func NewUserRepo(data *data.Data, logger log.Logger) repository.UserRepo {
	return &userRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "repository/user")),
		data: data,
		IAction: query.NewAction[model.SysUser](
			query.WithDB[model.SysUser](data.DB()),
		),
	}
}
