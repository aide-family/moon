package user

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
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

func (l *userRepoImpl) RelateRoles(ctx context.Context, userDo *dobo.UserDO, roleList []*dobo.RoleDO) error {
	roleModelList := slices.To(roleList, func(roleInfo *dobo.RoleDO) *model.SysRole {
		return roleInfo.ToModel()
	})

	return l.DB().WithContext(ctx).Model(userDo.ToModel()).
		Association(string(system.UserAssociationReplaceRoles)).
		Replace(&roleModelList)

}

func (l *userRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.UserDO, error) {
	userDetail, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}
	return dobo.UserModelToDO(userDetail), nil
}

func (l *userRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.UserDO, error) {
	userList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := slices.To(userList, func(user *model.SysUser) *dobo.UserDO {
		return dobo.UserModelToDO(user)
	})
	return list, nil
}

func (l *userRepoImpl) Create(ctx context.Context, user *dobo.UserDO) (*dobo.UserDO, error) {
	newUser := user.ToModel()
	if err := l.WithContext(ctx).Create(newUser); err != nil {
		return nil, err
	}
	return dobo.UserModelToDO(newUser), nil
}

func (l *userRepoImpl) Update(ctx context.Context, user *dobo.UserDO, scopes ...query.ScopeMethod) (*dobo.UserDO, error) {
	newUser := user.ToModel()
	if err := l.WithContext(ctx).Update(newUser, scopes...); err != nil {
		return nil, err
	}
	return dobo.UserModelToDO(newUser), nil
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
