package role

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.RoleRepo = (*roleRepoImpl)(nil)

type roleRepoImpl struct {
	log  *log.Helper
	data *data.Data
	query.IAction[model.SysRole]
}

func (l *roleRepoImpl) Create(ctx context.Context, role *dobo.RoleDO) (*dobo.RoleDO, error) {
	newRole := role.ModelRole()
	if err := l.WithContext(ctx).Create(newRole); err != nil {
		return nil, err
	}

	return dobo.RoleModelToDO(newRole), nil
}

func (l *roleRepoImpl) Update(ctx context.Context, role *dobo.RoleDO, scopes ...query.ScopeMethod) (*dobo.RoleDO, error) {
	newRole := role.ModelRole()
	if err := l.WithContext(ctx).Update(newRole, scopes...); err != nil {
		return nil, err
	}
	return dobo.RoleModelToDO(newRole), nil
}

func (l *roleRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	return l.WithContext(ctx).Delete(scopes...)
}

func (l *roleRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.RoleDO, error) {
	roleDetail, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}
	return dobo.RoleModelToDO(roleDetail), nil
}

func (l *roleRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.RoleDO, error) {
	roleList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := slices.To(roleList, func(role *model.SysRole) *dobo.RoleDO {
		return dobo.RoleModelToDO(role)
	})
	
	return list, nil
}

func NewRoleRepo(data *data.Data, logger log.Logger) repository.RoleRepo {
	return &roleRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "role")),
		data: data,
		IAction: query.NewAction[model.SysRole](
			query.WithDB[model.SysRole](data.DB()),
		),
	}
}
