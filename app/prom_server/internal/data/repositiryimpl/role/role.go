package role

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.RoleRepo = (*roleRepoImpl)(nil)

type roleRepoImpl struct {
	repository.UnimplementedRoleRepo
	log  *log.Helper
	data *data.Data
}

func (l *roleRepoImpl) Create(ctx context.Context, role *bo.RoleBO) (*bo.RoleBO, error) {
	newRole := role.ToModel()
	if err := l.data.DB().WithContext(ctx).Create(newRole).Error; err != nil {
		return nil, err
	}

	return bo.RoleModelToBO(newRole), nil
}

func (l *roleRepoImpl) Update(ctx context.Context, role *bo.RoleBO, scopes ...basescopes.ScopeMethod) (*bo.RoleBO, error) {
	if len(scopes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "更新角色时，必须指定更新条件")
	}

	// 判断要修改的数据是否只有一条
	var total int64
	if err := l.data.DB().WithContext(ctx).Model(new(do.SysRole)).Scopes(scopes...).Count(&total).Error; err != nil {
		return nil, err
	}
	if total != 1 {
		return nil, status.Error(codes.InvalidArgument, "更新角色时，必须指定一条数据")
	}

	newRole := role.ToModel()
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Updates(newRole).Error; err != nil {
		return nil, err
	}
	return bo.RoleModelToBO(newRole), nil
}

func (l *roleRepoImpl) UpdateAll(ctx context.Context, role *bo.RoleBO, scopes ...basescopes.ScopeMethod) error {
	newRole := role.ToModel()
	return l.data.DB().WithContext(ctx).Scopes(scopes...).Updates(newRole).Error
}

func (l *roleRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	if len(scopes) == 0 {
		return status.Error(codes.InvalidArgument, "删除角色时，必须指定删除条件")
	}

	return l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 清除关联关系
		if err := tx.Model(&do.SysRole{}).Scopes(scopes...).Association(basescopes.RoleAssociationReplaceApis).Clear(); err != nil {
			return err
		}

		if err := tx.Model(&do.SysRole{}).Scopes(scopes...).Association(basescopes.RoleAssociationReplaceUsers).Clear(); err != nil {
			return err
		}

		// 删除主数据
		if err := tx.Model(&do.SysRole{}).Scopes(scopes...).Delete(&do.SysRole{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (l *roleRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.RoleBO, error) {
	var roleDetail do.SysRole
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&roleDetail).Error; err != nil {
		return nil, err
	}
	return bo.RoleModelToBO(&roleDetail), nil
}

func (l *roleRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.RoleBO, error) {
	var roleModelList []*do.SysRole
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Find(&roleModelList).Error; err != nil {
		return nil, err
	}

	list := slices.To(roleModelList, func(role *do.SysRole) *bo.RoleBO {
		return bo.RoleModelToBO(role)
	})

	return list, nil
}

func (l *roleRepoImpl) List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.RoleBO, error) {
	var roleList []*do.SysRole
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&roleList).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.SysRole{}).Scopes(scopes...).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	list := slices.To(roleList, func(role *do.SysRole) *bo.RoleBO {
		return bo.RoleModelToBO(role)
	})

	return list, nil
}

func (l *roleRepoImpl) RelateApi(ctx context.Context, roleId uint32, apiList []*bo.ApiBO) error {
	if roleId == 1 {
		return perrors.ErrorPermissionDenied("超级管理员角色不允许操作")
	}
	var roleDetail do.SysRole
	if err := l.data.DB().WithContext(ctx).First(&roleDetail, roleId).Error; err != nil {
		return err
	}

	apiModelList := slices.To(apiList, func(api *bo.ApiBO) *do.SysAPI {
		return api.ToModel()
	})

	if err := l.data.DB().WithContext(ctx).Model(&roleDetail).Association(basescopes.RoleAssociationReplaceApis).Replace(&apiModelList); err != nil {
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
	}
}
