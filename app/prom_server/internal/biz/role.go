package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
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
		logX     repository.SysLogRepo
	}
)

func NewRoleBiz(roleRepo repository.RoleRepo, apiRepo repository.ApiRepo, dataRepo repository.DataRepo, logX repository.SysLogRepo, logger log.Logger) *RoleBiz {
	return &RoleBiz{
		log:      log.NewHelper(logger),
		roleRepo: roleRepo,
		apiRepo:  apiRepo,
		dataRepo: dataRepo,
		logX:     logX,
	}
}

// CreateRole 创建角色
func (b *RoleBiz) CreateRole(ctx context.Context, roleBO *bo.RoleBO) (*bo.RoleBO, error) {
	roleBO, err := b.roleRepo.Create(ctx, roleBO)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vo.ActionCreate, &bo.SysLogBo{
		ModuleName: vo.ModuleRole,
		ModuleId:   roleBO.Id,
		Content:    roleBO.String(),
		Title:      "创建角色",
	})
	return roleBO, nil
}

// DeleteRoleByIds 删除角色
func (b *RoleBiz) DeleteRoleByIds(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	// 查询
	oldRoles, err := b.roleRepo.Find(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err = b.roleRepo.Delete(ctx, basescopes.InIds(ids...)); err != nil {
		return err
	}
	list := slices.To(oldRoles, func(role *bo.RoleBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleRole,
			ModuleId:   role.Id,
			Content:    role.String(),
			Title:      "删除角色",
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionDelete, list...)
	return nil
}

// ListRole 角色列表
func (b *RoleBiz) ListRole(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.RoleBO, error) {
	roleBOList, err := b.roleRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return roleBOList, nil
}

// GetRoleById 获取角色
func (b *RoleBiz) GetRoleById(ctx context.Context, id uint32) (*bo.RoleBO, error) {
	roleBO, err := b.roleRepo.Get(ctx, basescopes.InIds(id), basescopes.RolePreloadUsers(), basescopes.RolePreloadApis())
	if err != nil {
		return nil, err
	}

	return roleBO, nil
}

// UpdateRoleById 更新角色
func (b *RoleBiz) UpdateRoleById(ctx context.Context, roleBO *bo.RoleBO) (*bo.RoleBO, error) {
	// 查询
	oldRole, err := b.roleRepo.Get(ctx, basescopes.InIds(roleBO.Id), basescopes.RolePreloadUsers(), basescopes.RolePreloadApis())
	if err != nil {
		return nil, err
	}
	newRoleBO, err := b.roleRepo.Update(ctx, roleBO, basescopes.InIds(roleBO.Id))
	if err != nil {
		return nil, err
	}
	b.cacheRoleByIds(roleBO.Id)
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, &bo.SysLogBo{
		ModuleName: vo.ModuleRole,
		ModuleId:   roleBO.Id,
		Content:    bo.NewChangeLogBo(oldRole, newRoleBO).String(),
		Title:      "更新角色",
	})
	return roleBO, nil
}

// UpdateRoleStatusById 更新角色状态
func (b *RoleBiz) UpdateRoleStatusById(ctx context.Context, status vo.Status, ids []uint32) error {
	oldList, err := b.roleRepo.Find(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	roleBo := &bo.RoleBO{Status: status}
	if err := b.roleRepo.UpdateAll(ctx, roleBo, basescopes.InIds(ids...)); err != nil {
		return err
	}

	b.cacheRoleByIds(ids...)
	list := slices.To(oldList, func(role *bo.RoleBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleRole,
			ModuleId:   role.Id,
			Content:    bo.NewChangeLogBo(role.Status.String(), status.String()).String(),
			Title:      "更新角色状态",
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, list...)
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
		cacheClient, err := b.dataRepo.Cache()
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

	if err = b.roleRepo.RelateApi(ctx, roleBoInfo.Id, findBoList); err != nil {
		return err
	}

	apiStr := slices.To(findBoList, func(api *bo.ApiBO) string {
		return api.String()
	})
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, &bo.SysLogBo{
		ModuleName: vo.ModuleRole,
		ModuleId:   roleBoInfo.Id,
		Content:    fmt.Sprintf(`{"apis":[%s]}`, strings.Join(apiStr, ",")),
		Title:      "关联角色和API",
	})

	return nil
}
