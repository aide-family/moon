package impl

import (
	"context"
	"strings"
	"time"

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

// NewNmspacesMenuRepository 创建租户菜单仓储实例
func NewNmspacesMenuRepository(d *data.Data) repository.NmspacesMenu {
	return NewNmspacesMenuRepositoryWithDB(d.DB())
}

// NewNmspacesMenuRepositoryWithDB 使用指定数据库连接创建租户菜单仓储实例
func NewNmspacesMenuRepositoryWithDB(db *gorm.DB) repository.NmspacesMenu {
	return &nmspacesMenuRepository{db: db}
}

type nmspacesMenuRepository struct {
	db *gorm.DB
}

// ListNmspacesMenu 分页查询租户菜单列表
func (r *nmspacesMenuRepository) ListNmspacesMenu(ctx context.Context, req *bo.ListNmspacesMenuBo) ([]*bo.NmspacesMenuItemBo, int64, error) {
	db := getDBWithTransaction(ctx, r.db).WithContext(ctx).
		Where("namespace_uid = ?", req.NamespaceUID)

	// 关键字模糊查询（菜单编码或菜单名称）
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("menu_code LIKE ? OR menu_name LIKE ?", keyword, keyword)
	}
	// 状态筛选
	if req.MenuStatus != nil {
		db = db.Where("menu_status = ?", *req.MenuStatus)
	}

	// 统计总数
	var total int64
	if err := db.Model(&do.NmspacesMenu{}).Count(&total).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("list nmspaces menu failed: %v", err)
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
		db = db.Order("menu_sort_order ASC, created_at DESC")
	}

	// 分页
	if req.Page > 0 && req.PageSize > 0 {
		db = db.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
	}

	var menus []*do.NmspacesMenu
	if err := db.Find(&menus).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("list nmspaces menu failed: %v", err)
	}

	items := make([]*bo.NmspacesMenuItemBo, 0, len(menus))
	for _, menu := range menus {
		items = append(items, convert.NmspacesMenuToBo(menu))
	}
	return items, total, nil
}

// GetNmspacesMenu 查询租户菜单详情
func (r *nmspacesMenuRepository) GetNmspacesMenu(ctx context.Context, namespaceUID snowflake.ID, menuCode string) (*bo.NmspacesMenuItemBo, error) {
	db := getDBWithTransaction(ctx, r.db).WithContext(ctx)
	var menu do.NmspacesMenu
	if err := db.Where("namespace_uid = ? AND menu_code = ?", namespaceUID, menuCode).First(&menu).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, merr.ErrorInternalServer("get nmspaces menu failed: %v", err)
	}
	return convert.NmspacesMenuToBo(&menu), nil
}

// CreateNmspacesMenu 新增租户菜单
func (r *nmspacesMenuRepository) CreateNmspacesMenu(ctx context.Context, req *bo.CreateNmspacesMenuBo) error {
	db := getDBWithTransaction(ctx, r.db)
	menu := &do.NmspacesMenu{
		NamespaceUID:    req.NamespaceUID,
		MenuCode:        req.MenuCode,
		ParentMenuCode:  req.ParentMenuCode,
		MenuName:        req.MenuName,
		MenuPath:        req.MenuPath,
		MenuIcon:        req.MenuIcon,
		MenuComponent:   req.MenuComponent,
		MenuRedirect:    req.MenuRedirect,
		MenuIsHidden:    req.MenuIsHidden,
		MenuIsKeepAlive: req.MenuIsKeepAlive,
		MenuSortOrder:   req.MenuSortOrder,
		MenuStatus:      req.MenuStatus,
		MenuRemark:      req.MenuRemark,
		Creator:         req.Creator,
	}
	return db.Create(menu).Error
}

// UpdateNmspacesMenu 编辑租户菜单
func (r *nmspacesMenuRepository) UpdateNmspacesMenu(ctx context.Context, req *bo.UpdateNmspacesMenuBo) error {
	db := getDBWithTransaction(ctx, r.db)
	updates := make(map[string]interface{})
	if req.ParentMenuCode != nil {
		updates["parent_menu_code"] = *req.ParentMenuCode
	}
	if req.MenuName != nil {
		updates["menu_name"] = *req.MenuName
	}
	if req.MenuPath != nil {
		updates["menu_path"] = *req.MenuPath
	}
	if req.MenuIcon != nil {
		updates["menu_icon"] = *req.MenuIcon
	}
	if req.MenuComponent != nil {
		updates["menu_component"] = *req.MenuComponent
	}
	if req.MenuRedirect != nil {
		updates["menu_redirect"] = *req.MenuRedirect
	}
	if req.MenuIsHidden != nil {
		updates["menu_is_hidden"] = *req.MenuIsHidden
	}
	if req.MenuIsKeepAlive != nil {
		updates["menu_is_keep_alive"] = *req.MenuIsKeepAlive
	}
	if req.MenuSortOrder != nil {
		updates["menu_sort_order"] = *req.MenuSortOrder
	}
	if req.MenuStatus != nil {
		updates["menu_status"] = *req.MenuStatus
	}
	if req.MenuRemark != nil {
		updates["menu_remark"] = *req.MenuRemark
	}
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return db.Model(&do.NmspacesMenu{}).
		Where("namespace_uid = ? AND menu_code = ? AND deleted_at IS NULL", req.NamespaceUID, req.MenuCode).
		Updates(updates).Error
}

// DeleteNmspacesMenu 删除租户菜单（软删除）
func (r *nmspacesMenuRepository) DeleteNmspacesMenu(ctx context.Context, namespaceUID snowflake.ID, menuCode string) error {
	db := getDBWithTransaction(ctx, r.db)
	return db.Where("namespace_uid = ? AND menu_code = ? AND deleted_at IS NULL", namespaceUID, menuCode).
		Delete(&do.NmspacesMenu{}).Error
}

// ExistsByNamespaceAndMenuCode 检查 menuCode 是否已存在（含软删除行）
func (r *nmspacesMenuRepository) ExistsByNamespaceAndMenuCode(ctx context.Context, namespaceUID snowflake.ID, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesMenu{}).Unscoped().
		Where("namespace_uid = ? AND menu_code = ?", namespaceUID, menuCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamespaceAndMenuCodeNotDeleted 检查 menuCode 是否在未删除行中存在
func (r *nmspacesMenuRepository) ExistsByNamespaceAndMenuCodeNotDeleted(ctx context.Context, namespaceUID snowflake.ID, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesMenu{}).
		Where("namespace_uid = ? AND menu_code = ? AND deleted_at IS NULL", namespaceUID, menuCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamespaceAndParentAndName 检查同一 parentMenuCode 下 menuName 是否已存在（未删除行）
func (r *nmspacesMenuRepository) ExistsByNamespaceAndParentAndName(ctx context.Context, namespaceUID snowflake.ID, parentMenuCode, menuName string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesMenu{}).
		Where("namespace_uid = ? AND parent_menu_code = ? AND menu_name = ? AND deleted_at IS NULL", namespaceUID, parentMenuCode, menuName).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamespaceAndParentMenuCode 检查 parentMenuCode 是否存在且未删除
func (r *nmspacesMenuRepository) ExistsByNamespaceAndParentMenuCode(ctx context.Context, namespaceUID snowflake.ID, parentMenuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesMenu{}).
		Where("namespace_uid = ? AND menu_code = ? AND deleted_at IS NULL", namespaceUID, parentMenuCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountChildren 统计子菜单数量（未删除行）
func (r *nmspacesMenuRepository) CountChildren(ctx context.Context, namespaceUID snowflake.ID, parentMenuCode string) (int64, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesMenu{}).
		Where("namespace_uid = ? AND parent_menu_code = ? AND deleted_at IS NULL", namespaceUID, parentMenuCode).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
