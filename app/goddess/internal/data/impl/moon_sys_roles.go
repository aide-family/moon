package impl

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
	"github.com/aide-family/goddess/internal/data/impl/query"
)

// NewSysRoleRepository 创建系统角色仓储实例
func NewSysRoleRepository(d *data.Data) repository.SysRole {
	return NewSysRoleRepositoryWithDB(d.DB())
}

// NewSysRoleRepositoryWithDB 使用指定数据库连接创建系统角色仓储实例
func NewSysRoleRepositoryWithDB(db *gorm.DB) repository.SysRole {
	query.SetDefault(db)
	return &sysRoleRepository{db: db}
}

type sysRoleRepository struct {
	db *gorm.DB
}

// ListSysRole 分页查询系统角色列表
// 支持关键字模糊查询（角色名称）、角色编码精确查询、状态筛选，默认按创建时间降序排序
func (r *sysRoleRepository) ListSysRole(ctx context.Context, req *bo.ListSysRoleBo) (*bo.PageResponseBo[*bo.SysRoleItemBo], error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysRole
	wrappers := mutation.WithContext(ctx)

	// 角色编码精确查询
	if strutil.IsNotEmpty(req.RoleCode) {
		wrappers = wrappers.Where(mutation.RoleCode.Eq(req.RoleCode))
	}
	// 角色名称模糊查询（关键字）
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.RoleName.Like(keyword))
	}
	// 角色状态筛选（RoleStatus 为 *int32，nil 表示不过滤，支持筛选状态值为 0 的情况）
	if req.RoleStatus != nil {
		wrappers = wrappers.Where(mutation.RoleStatus.Eq(*req.RoleStatus))
	}

	// 排序：支持 sortBy 字段，默认按创建时间降序
	orderField, ok := mutation.GetFieldByName(req.SortBy)
	if ok && req.SortBy != "" {
		if strings.EqualFold(req.SortOrder, "asc") {
			wrappers = wrappers.Order(orderField)
		} else {
			wrappers = wrappers.Order(orderField.Desc())
		}
	} else {
		// 默认按创建时间降序
		wrappers = wrappers.Order(mutation.CreatedAt.Desc())
	}

	// 统计总数
	if req.Page > 0 && req.PageSize > 0 {
		total, err := wrappers.Count()
		if err != nil {
			return nil, merr.ErrorInternalServer("list sys role failed: %v", err)
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
	}

	roles, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list sys role failed: %v", err)
	}

	items := make([]*bo.SysRoleItemBo, 0, len(roles))
	for _, role := range roles {
		items = append(items, convert.SysRoleToBo(role))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

// GetSysRole 根据 roleCode 查询角色详情
// 基础版：仅查询角色主表，菜单和功能权限待关联表建立后补全
func (r *sysRoleRepository) GetSysRole(ctx context.Context, roleCode string) (*bo.GetSysRoleBo, error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysRole

	role, err := mutation.WithContext(ctx).Where(mutation.RoleCode.Eq(roleCode)).First()
	if err != nil {
		return nil, merr.ErrorInternalServer("get sys role failed: %v", err)
	}
	if role == nil {
		return nil, nil
	}

	// 基础版：菜单和功能权限返回空数组，待关联表建立后补全
	return &bo.GetSysRoleBo{
		SysRoleItemBo: convert.SysRoleToBo(role),
		MenuCodes:     make([]*bo.SysRoleMenuItemBo, 0),
		FunctionKeys:  make([]*bo.SysRoleFunctionItemBo, 0),
	}, nil
}

// ExistsByRoleCode 检查 roleCode 是否已存在（含软删除行，避免重复注册）
func (r *sysRoleRepository) ExistsByRoleCode(ctx context.Context, roleCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	// 含软删除行，使用 Unscoped 查询
	if err := db.Model(&do.SysRole{}).Unscoped().Where("role_code = ?", roleCode).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByRoleName 检查 roleName 是否在未删除行中已存在
func (r *sysRoleRepository) ExistsByRoleName(ctx context.Context, roleName string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysRole{}).Where("role_name = ? AND deleted_at IS NULL", roleName).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateSysRole 新增系统角色
func (r *sysRoleRepository) CreateSysRole(ctx context.Context, req *bo.CreateSysRoleBo) error {
	db := getDBWithTransaction(ctx, r.db)
	role := &do.SysRole{
		RoleCode:      req.RoleCode,
		RoleName:      req.RoleName,
		RoleDataScope: req.RoleDataScope,
		RoleStatus:    req.RoleStatus,
		RoleSortOrder: req.RoleSortOrder,
		RoleRemark:    req.RoleRemark,
		Creator:       req.Creator,
	}
	return db.Create(role).Error
}

// ExistsByRoleCodeNotDeleted 检查 roleCode 是否在未删除行中存在
func (r *sysRoleRepository) ExistsByRoleCodeNotDeleted(ctx context.Context, roleCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysRole{}).Where("role_code = ? AND deleted_at IS NULL", roleCode).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateSysRole 编辑系统角色（仅更新非 nil 字段）
func (r *sysRoleRepository) UpdateSysRole(ctx context.Context, req *bo.UpdateSysRoleBo) error {
	db := getDBWithTransaction(ctx, r.db)
	updates := make(map[string]interface{})
	if req.RoleName != nil {
		updates["role_name"] = *req.RoleName
	}
	if req.RoleDataScope != nil {
		updates["role_data_scope"] = *req.RoleDataScope
	}
	if req.RoleStatus != nil {
		updates["role_status"] = *req.RoleStatus
	}
	if req.RoleSortOrder != nil {
		updates["role_sort_order"] = *req.RoleSortOrder
	}
	if req.RoleRemark != nil {
		updates["role_remark"] = *req.RoleRemark
	}
	if len(updates) == 0 {
		return nil
	}
	return db.Model(&do.SysRole{}).Where("role_code = ? AND deleted_at IS NULL", req.RoleCode).Updates(updates).Error
}

// DeleteSysRole 删除系统角色（软删除，role_status 置为 0）
func (r *sysRoleRepository) DeleteSysRole(ctx context.Context, roleCode string) error {
	db := getDBWithTransaction(ctx, r.db)
	return db.Model(&do.SysRole{}).Where("role_code = ? AND deleted_at IS NULL", roleCode).
		Update("role_status", 0).Error
}
