package user

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/do/systemscopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.UserRepo = (*userRepoImpl)(nil)

type userRepoImpl struct {
	repository.UnimplementedUserRepo
	log  *log.Helper
	data *data.Data
}

func (l *userRepoImpl) RelateRoles(ctx context.Context, userBO *bo.UserBO, roleList []*bo.RoleBO) error {
	roleModelList := slices.To(roleList, func(roleInfo *bo.RoleBO) *do.SysRole {
		return roleInfo.ToModel()
	})

	return l.data.DB().WithContext(ctx).Model(userBO.ToModel()).
		Association(systemscopes.UserAssociationReplaceRoles).
		Replace(&roleModelList)

}

func (l *userRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.UserBO, error) {
	var userDetail do.SysUser
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&userDetail).Error; err != nil {
		return nil, err
	}
	return bo.UserModelToBO(&userDetail), nil
}

func (l *userRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.UserBO, error) {
	var userDetailList []*do.SysUser
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Find(&userDetailList).Error; err != nil {
		return nil, err
	}
	list := slices.To(userDetailList, func(user *do.SysUser) *bo.UserBO {
		return bo.UserModelToBO(user)
	})
	return list, nil
}

func (l *userRepoImpl) Count(ctx context.Context, scopes ...query.ScopeMethod) (int64, error) {
	var count int64
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (l *userRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.UserBO, error) {
	var userList []*do.SysUser
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&userList).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.SysUser{}).Scopes(scopes...).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	list := slices.To(userList, func(user *do.SysUser) *bo.UserBO {
		return bo.UserModelToBO(user)
	})
	return list, nil
}

func (l *userRepoImpl) Create(ctx context.Context, user *bo.UserBO) (*bo.UserBO, error) {
	newUser := user.ToModel()
	if err := l.data.DB().WithContext(ctx).Create(newUser).Error; err != nil {
		return nil, err
	}
	return bo.UserModelToBO(newUser), nil
}

func (l *userRepoImpl) Update(ctx context.Context, user *bo.UserBO, scopes ...query.ScopeMethod) (*bo.UserBO, error) {
	newUser := user.ToModel()
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Updates(newUser).Error; err != nil {
		return nil, err
	}
	return bo.UserModelToBO(newUser), nil
}

func (l *userRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	return l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := basescopes.WithTx(ctx, tx)
		// 删除关联关系
		if err := tx.WithContext(txCtx).Model(&do.SysUser{}).Association(systemscopes.UserAssociationReplaceRoles).Clear(); err != nil {
			return err
		}
		// 删除主数据
		if err := tx.WithContext(txCtx).Scopes(scopes...).Delete(&do.SysUser{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func NewUserRepo(data *data.Data, logger log.Logger) repository.UserRepo {
	return &userRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "repository/user")),
		data: data,
	}
}
