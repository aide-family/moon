package role

import (
	"context"
	"strconv"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/systemscopes"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.RoleRepo = (*roleRepoImpl)(nil)

type roleRepoImpl struct {
	repository.UnimplementedRoleRepo
	log  *log.Helper
	data *data.Data
	query.IAction[model.SysRole]
}

func (l *roleRepoImpl) Create(ctx context.Context, role *bo.RoleBO) (*bo.RoleBO, error) {
	newRole := role.ToModel()
	if err := l.WithContext(ctx).Create(newRole); err != nil {
		return nil, err
	}

	return bo.RoleModelToBO(newRole), nil
}

func (l *roleRepoImpl) Update(ctx context.Context, role *bo.RoleBO, scopes ...query.ScopeMethod) (*bo.RoleBO, error) {
	if len(scopes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "更新角色时，必须指定更新条件")
	}

	// 判断要修改的数据是否只有一条
	total, err := l.WithContext(ctx).Count(scopes...)
	if err != nil {
		return nil, err
	}
	if total != 1 {
		return nil, status.Error(codes.InvalidArgument, "更新角色时，必须指定一条数据")
	}

	newRole := role.ToModel()
	if err = l.WithContext(ctx).Update(newRole, scopes...); err != nil {
		return nil, err
	}
	return bo.RoleModelToBO(newRole), nil
}

func (l *roleRepoImpl) UpdateAll(ctx context.Context, role *bo.RoleBO, scopes ...query.ScopeMethod) error {
	newRole := role.ToModel()
	return l.WithContext(ctx).Update(newRole, scopes...)
}

func (l *roleRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	if len(scopes) == 0 {
		return status.Error(codes.InvalidArgument, "删除角色时，必须指定删除条件")
	}

	return l.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 清除关联关系
		if err := tx.Model(&model.SysRole{}).Scopes(scopes...).Association(systemscopes.RoleAssociationReplaceApis).Clear(); err != nil {
			return err
		}

		if err := tx.Model(&model.SysRole{}).Scopes(scopes...).Association(systemscopes.RoleAssociationReplaceUsers).Clear(); err != nil {
			return err
		}

		// 删除主数据
		if err := tx.Model(&model.SysRole{}).Scopes(scopes...).Delete(&model.SysRole{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (l *roleRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.RoleBO, error) {
	roleDetail, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}
	return bo.RoleModelToBO(roleDetail), nil
}

func (l *roleRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.RoleBO, error) {
	var roleModelList []*model.SysRole

	if err := l.DB().WithContext(ctx).Scopes(scopes...).Find(&roleModelList).Error; err != nil {
		return nil, err
	}

	list := slices.To(roleModelList, func(role *model.SysRole) *bo.RoleBO {
		return bo.RoleModelToBO(role)
	})

	return list, nil
}

func (l *roleRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.RoleBO, error) {
	roleList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := slices.To(roleList, func(role *model.SysRole) *bo.RoleBO {
		return bo.RoleModelToBO(role)
	})

	return list, nil
}

func (l *roleRepoImpl) RelateApi(ctx context.Context, roleId uint32, apiList []*bo.ApiBO) error {
	if roleId == 1 {
		return perrors.ErrorPermissionDenied("超级管理员角色不允许操作")
	}
	roleDetail, err := l.WithContext(ctx).FirstByID(roleId)
	if err != nil {
		return err
	}

	apiModelList := slices.To(apiList, func(api *bo.ApiBO) *model.SysAPI {
		return api.ToModel()
	})

	err = l.DB().WithContext(ctx).Model(roleDetail).Association(systemscopes.RoleAssociationReplaceApis).Replace(&apiModelList)
	if err != nil {
		return err
	}

	enforcer := l.data.Enforcer()
	polices := make([][]string, 0, len(apiList))
	roleIdStr := strconv.Itoa(int(roleId))
	for _, api := range apiList {
		polices = append(polices, []string{roleIdStr, api.Path, api.Method})
	}

	// 删除这个角色之前的权限
	_, removeErr := enforcer.RemoveFilteredPolicy(0, roleIdStr)
	if removeErr != nil {
		return removeErr
	}

	if len(polices) == 0 {
		return nil
	}

	policiesAddOk, err := enforcer.AddPolicies(polices)
	if err != nil {
		return err
	}
	if !policiesAddOk {
		return perrors.ErrorUnknown("add policies failed")
	}

	return nil
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
