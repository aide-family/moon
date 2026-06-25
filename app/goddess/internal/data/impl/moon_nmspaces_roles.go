package impl

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// NewNmspacesRoleRepository 创建租户角色仓储实例
func NewNmspacesRoleRepository(d *data.Data) repository.NmspacesRole {
	return NewNmspacesRoleRepositoryWithDB(d.DB())
}

// NewNmspacesRoleRepositoryWithDB 使用指定数据库连接创建租户角色仓储实例
func NewNmspacesRoleRepositoryWithDB(db *gorm.DB) repository.NmspacesRole {
	return &nmspacesRoleRepository{db: db}
}

type nmspacesRoleRepository struct {
	db *gorm.DB
}

// ListNmspacesRole 分页查询租户角色列表
func (r *nmspacesRoleRepository) ListNmspacesRole(ctx context.Context, req *bo.ListNmspacesRoleBo) (*bo.PageResponseBo[*bo.NmspacesRoleItemBo], error) {
	db := getDBWithTransaction(ctx, r.db).WithContext(ctx).
		Where("namespace_uid = ?", req.NamespaceUID)

	// 角色编码精确查询
	if strutil.IsNotEmpty(req.RoleCode) {
		db = db.Where("role_code = ?", req.RoleCode)
	}
	// 角色名称模糊查询
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("role_name LIKE ?", keyword)
	}
	// 状态筛选
	if req.RoleStatus != nil {
		db = db.Where("role_status = ?", *req.RoleStatus)
	}

	// 排序
	if strutil.IsNotEmpty(req.SortBy) {
		orderField := req.SortBy
		if strings.EqualFold(req.SortOrder, "asc") {
			db = db.Order(orderField)
		} else {
			db = db.Order(orderField + " DESC")
		}
	} else {
		db = db.Order("created_at DESC")
	}

	// 统计总数
	if req.Page > 0 && req.PageSize > 0 {
		var total int64
		if err := db.Model(&do.NmspacesRole{}).Count(&total).Error; err != nil {
			return nil, merr.ErrorInternalServer("list nmspaces role failed: %v", err)
		}
		req.WithTotal(total)
		db = db.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
	}

	var roles []*do.NmspacesRole
	if err := db.Find(&roles).Error; err != nil {
		return nil, merr.ErrorInternalServer("list nmspaces role failed: %v", err)
	}

	items := make([]*bo.NmspacesRoleItemBo, 0, len(roles))
	for _, role := range roles {
		items = append(items, convert.NmspacesRoleToBo(role))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

// GetNmspacesRole 查询租户角色详情
func (r *nmspacesRoleRepository) GetNmspacesRole(ctx context.Context, namespaceUID snowflake.ID, roleCode string) (*bo.NmspacesRoleItemBo, error) {
	db := getDBWithTransaction(ctx, r.db).WithContext(ctx)
	var role do.NmspacesRole
	if err := db.Where("namespace_uid = ? AND role_code = ?", namespaceUID, roleCode).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, merr.ErrorInternalServer("get nmspaces role failed: %v", err)
	}
	return convert.NmspacesRoleToBo(&role), nil
}

// CreateNmspacesRole 新增租户角色
func (r *nmspacesRoleRepository) CreateNmspacesRole(ctx context.Context, req *bo.CreateNmspacesRoleBo) error {
	db := getDBWithTransaction(ctx, r.db)
	role := &do.NmspacesRole{
		NamespaceUID:  req.NamespaceUID,
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

// UpdateNmspacesRole 编辑租户角色
func (r *nmspacesRoleRepository) UpdateNmspacesRole(ctx context.Context, req *bo.UpdateNmspacesRoleBo) error {
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
	return db.Model(&do.NmspacesRole{}).
		Where("namespace_uid = ? AND role_code = ? AND deleted_at IS NULL", req.NamespaceUID, req.RoleCode).
		Updates(updates).Error
}

// DeleteNmspacesRole 删除租户角色（软删除）
func (r *nmspacesRoleRepository) DeleteNmspacesRole(ctx context.Context, namespaceUID snowflake.ID, roleCode string) error {
	db := getDBWithTransaction(ctx, r.db)
	return db.Where("namespace_uid = ? AND role_code = ? AND deleted_at IS NULL", namespaceUID, roleCode).
		Delete(&do.NmspacesRole{}).Error
}

// ExistsByNamespaceAndRoleCode 检查 roleCode 是否已存在（含软删除行）
func (r *nmspacesRoleRepository) ExistsByNamespaceAndRoleCode(ctx context.Context, namespaceUID snowflake.ID, roleCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesRole{}).Unscoped().
		Where("namespace_uid = ? AND role_code = ?", namespaceUID, roleCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamespaceAndRoleName 检查 roleName 是否在未删除行中已存在
func (r *nmspacesRoleRepository) ExistsByNamespaceAndRoleName(ctx context.Context, namespaceUID snowflake.ID, roleName string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesRole{}).
		Where("namespace_uid = ? AND role_name = ? AND deleted_at IS NULL", namespaceUID, roleName).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamespaceAndRoleCodeNotDeleted 检查 roleCode 是否在未删除行中存在
func (r *nmspacesRoleRepository) ExistsByNamespaceAndRoleCodeNotDeleted(ctx context.Context, namespaceUID snowflake.ID, roleCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesRole{}).
		Where("namespace_uid = ? AND role_code = ? AND deleted_at IS NULL", namespaceUID, roleCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
