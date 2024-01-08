package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/do/systemscopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/after"
)

type (
	RoleBiz struct {
		log *log.Helper

		roleRepo repository.RoleRepo
		apiRepo  repository.ApiRepo
		dataRepo repository.DataRepo
	}
)

func NewRoleBiz(roleRepo repository.RoleRepo, apiRepo repository.ApiRepo, dataRepo repository.DataRepo, logger log.Logger) *RoleBiz {
	return &RoleBiz{
		log:      log.NewHelper(logger),
		roleRepo: roleRepo,
		apiRepo:  apiRepo,
		dataRepo: dataRepo,
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
	return b.roleRepo.Delete(ctx, basescopes.InIds(ids...))
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
	roleBO, err := b.roleRepo.Get(ctx, basescopes.InIds(id), systemscopes.RolePreloadUsers(), systemscopes.RolePreloadApis())
	if err != nil {
		return nil, err
	}

	return roleBO, nil
}

// UpdateRoleById 更新角色
func (b *RoleBiz) UpdateRoleById(ctx context.Context, roleBO *bo.RoleBO) (*bo.RoleBO, error) {
	roleBO, err := b.roleRepo.Update(ctx, roleBO, basescopes.InIds(roleBO.Id))
	if err != nil {
		return nil, err
	}
	b.cacheRoleByIds(roleBO.Id)
	return roleBO, nil
}

// UpdateRoleStatusById 更新角色状态
func (b *RoleBiz) UpdateRoleStatusById(ctx context.Context, status vo.Status, ids []uint32) error {
	roleBo := &bo.RoleBO{Status: status}
	if err := b.roleRepo.UpdateAll(ctx, roleBo, basescopes.InIds(ids...)); err != nil {
		return err
	}
	b.cacheRoleByIds(ids...)
	return nil
}

// cacheRoleByIds 缓存角色信息
func (b *RoleBiz) cacheRoleByIds(roleIds ...uint32) {
	go func() {
		defer after.Recover(b.log)
		db, err := b.dataRepo.DB()
		if err != nil {
			return
		}
		cacheClient, err := b.dataRepo.Client()
		if err != nil {
			return
		}
		if err = do.CacheDisabledRoles(db, cacheClient, roleIds...); err != nil {
			b.log.Error(err)
		}
	}()
}

// RelateApiById 关联角色和api
func (b *RoleBiz) RelateApiById(ctx context.Context, roleId uint32, apiIds []uint32) error {
	var (
		findBoList []*bo.ApiBO
		err        error
	)

	if len(apiIds) > 0 {
		// 查询API
		findBoList, err = b.apiRepo.Find(ctx, basescopes.InIds(apiIds...))
		if err != nil {
			return err
		}
	}

	roleBoInfo, err := b.roleRepo.Get(ctx, basescopes.InIds(roleId))
	if err != nil {
		return err
	}

	return b.roleRepo.RelateApi(ctx, roleBoInfo.Id, findBoList)
}
