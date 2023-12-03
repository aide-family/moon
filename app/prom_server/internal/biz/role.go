package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/helper/model/system"
	"prometheus-manager/pkg/helper/valueobj"
)

type (
	RoleBiz struct {
		log *log.Helper

		roleRepo repository.RoleRepo
		apiRepo  repository.ApiRepo
	}
)

func NewRoleBiz(roleRepo repository.RoleRepo, apiRepo repository.ApiRepo, logger log.Logger) *RoleBiz {
	return &RoleBiz{
		log:      log.NewHelper(logger),
		roleRepo: roleRepo,
		apiRepo:  apiRepo,
	}
}

// CreateRole 创建角色
func (b *RoleBiz) CreateRole(ctx context.Context, roleBO *bo.RoleBO) (*bo.RoleBO, error) {
	roleBO, err := b.roleRepo.Create(ctx, roleBO)
	if err != nil {
		return nil, err
	}

	return roleBO, nil
}

// DeleteRoleByIds 删除角色
func (b *RoleBiz) DeleteRoleByIds(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	return b.roleRepo.Delete(ctx, system.RoleInIds(ids...))
}

// ListRole 角色列表
func (b *RoleBiz) ListRole(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.RoleBO, error) {
	roleBOList, err := b.roleRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return roleBOList, nil
}

// GetRoleById 获取角色
func (b *RoleBiz) GetRoleById(ctx context.Context, id uint32) (*bo.RoleBO, error) {
	roleBO, err := b.roleRepo.Get(ctx, system.RoleInIds(id), system.RolePreloadUsers[uint32]())
	if err != nil {
		return nil, err
	}

	return roleBO, nil
}

// UpdateRoleById 更新角色
func (b *RoleBiz) UpdateRoleById(ctx context.Context, roleBO *bo.RoleBO) (*bo.RoleBO, error) {
	roleBO, err := b.roleRepo.Update(ctx, roleBO, system.RoleInIds(roleBO.Id))
	if err != nil {
		return nil, err
	}

	return roleBO, nil
}

// UpdateRoleStatusById 更新角色状态
func (b *RoleBiz) UpdateRoleStatusById(ctx context.Context, status api.Status, ids []uint32) error {
	do := &bo.RoleBO{Status: valueobj.Status(status)}
	do, err := b.roleRepo.Update(ctx, do, system.RoleInIds(ids...))
	if err != nil {
		return err
	}

	return nil
}

// RelateApiById 关联角色和api
func (b *RoleBiz) RelateApiById(ctx context.Context, roleId uint32, apiIds []uint32) error {
	var (
		findBoList []*bo.ApiBO
		err        error
	)

	if len(apiIds) > 0 {
		// 查询API
		findBoList, err = b.apiRepo.Find(ctx, system.ApiInIds(apiIds...))
		if err != nil {
			return err
		}
	}

	roleBoInfo, err := b.roleRepo.Get(ctx, system.RoleInIds(roleId))
	if err != nil {
		return err
	}

	return b.roleRepo.RelateApi(ctx, roleBoInfo.Id, findBoList)
}
