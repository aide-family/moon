package impl

import (
	"context"
	"strings"
	"time"

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

// NewSysMenuRepository 创建系统菜单仓储实例
func NewSysMenuRepository(d *data.Data) repository.SysMenu {
	return NewSysMenuRepositoryWithDB(d.DB())
}

// NewSysMenuRepositoryWithDB 使用指定数据库连接创建系统菜单仓储实例
func NewSysMenuRepositoryWithDB(db *gorm.DB) repository.SysMenu {
	query.SetDefault(db)
	return &sysMenuRepository{db: db}
}

type sysMenuRepository struct {
	db *gorm.DB
}

// ListSysMenu 分页查询系统菜单列表
func (r *sysMenuRepository) ListSysMenu(ctx context.Context, req *bo.ListSysMenuBo) ([]*bo.SysMenuItemBo, int64, error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysMenu
	wrappers := mutation.WithContext(ctx)

	// 关键字模糊查询（菜单编码 OR 菜单名称）
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.MenuCode.Like(keyword)).
			Or(mutation.MenuName.Like(keyword))
	}
	// 菜单状态筛选
	if req.MenuStatus != nil {
		wrappers = wrappers.Where(mutation.MenuStatus.Eq(*req.MenuStatus))
	}

	// 统计总数
	var total int64
	if req.Page > 0 && req.PageSize > 0 {
		var err error
		total, err = wrappers.Count()
		if err != nil {
			return nil, 0, merr.ErrorInternalServer("list sys menu count failed: %v", err)
		}
	}

	// 排序：默认按 menuSortOrder 升序，createdAt 降序
	orderField, ok := mutation.GetFieldByName(req.SortBy)
	if ok && req.SortBy != "" {
		if strings.EqualFold(req.SortOrder, "asc") {
			wrappers = wrappers.Order(orderField)
		} else {
			wrappers = wrappers.Order(orderField.Desc())
		}
	} else {
		wrappers = wrappers.Order(mutation.MenuSortOrder.Asc(), mutation.CreatedAt.Desc())
	}

	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
	}

	menus, err := wrappers.Find()
	if err != nil {
		return nil, 0, merr.ErrorInternalServer("list sys menu failed: %v", err)
	}

	items := make([]*bo.SysMenuItemBo, 0, len(menus))
	for _, menu := range menus {
		items = append(items, convert.SysMenuToBo(menu))
	}
	return items, total, nil
}

// TreeSysMenu 查询全量菜单（内存中构建树）
func (r *sysMenuRepository) TreeSysMenu(ctx context.Context, req *bo.TreeSysMenuBo) ([]*bo.SysMenuItemBo, error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysMenu
	wrappers := mutation.WithContext(ctx)

	if req.MenuStatus != nil {
		wrappers = wrappers.Where(mutation.MenuStatus.Eq(*req.MenuStatus))
	}
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.MenuCode.Like(keyword)).
			Or(mutation.MenuName.Like(keyword))
	}

	wrappers = wrappers.Order(mutation.MenuSortOrder.Asc())

	menus, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("tree sys menu failed: %v", err)
	}

	items := make([]*bo.SysMenuItemBo, 0, len(menus))
	for _, menu := range menus {
		items = append(items, convert.SysMenuToBo(menu))
	}
	return items, nil
}

// GetSysMenu 根据 menuCode 查询菜单详情
func (r *sysMenuRepository) GetSysMenu(ctx context.Context, menuCode string) (*bo.SysMenuItemBo, error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysMenu
	menu, err := mutation.WithContext(ctx).Where(mutation.MenuCode.Eq(menuCode)).First()
	if err != nil {
		return nil, merr.ErrorInternalServer("get sys menu failed: %v", err)
	}
	if menu == nil {
		return nil, nil
	}
	return convert.SysMenuToBo(menu), nil
}

// CountChildren 统计子菜单数量
func (r *sysMenuRepository) CountChildren(ctx context.Context, parentMenuCode string) (int64, error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysMenu
	count, err := mutation.WithContext(ctx).Where(mutation.ParentMenuCode.Eq(parentMenuCode)).Count()
	if err != nil {
		return 0, merr.ErrorInternalServer("count children failed: %v", err)
	}
	return count, nil
}

// ExistsByMenuCode 检查 menuCode 是否已存在（含软删除行）
func (r *sysMenuRepository) ExistsByMenuCode(ctx context.Context, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysMenu{}).Unscoped().Where("menu_code = ?", menuCode).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByMenuCodeNotDeleted 检查 menuCode 是否在未删除行中存在
func (r *sysMenuRepository) ExistsByMenuCodeNotDeleted(ctx context.Context, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysMenu{}).Where("menu_code = ? AND deleted_at IS NULL", menuCode).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByParentAndName 检查同一 parentMenuCode 下 menuName 是否已存在
func (r *sysMenuRepository) ExistsByParentAndName(ctx context.Context, parentMenuCode, menuName string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysMenu{}).Where("parent_menu_code = ? AND menu_name = ? AND deleted_at IS NULL", parentMenuCode, menuName).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByParentMenuCode 检查 parentMenuCode 是否存在且未删除
func (r *sysMenuRepository) ExistsByParentMenuCode(ctx context.Context, parentMenuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysMenu{}).Where("menu_code = ? AND deleted_at IS NULL", parentMenuCode).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateSysMenu 新增系统菜单
func (r *sysMenuRepository) CreateSysMenu(ctx context.Context, req *bo.CreateSysMenuBo) error {
	db := getDBWithTransaction(ctx, r.db)
	menu := &do.SysMenu{
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

// UpdateSysMenu 编辑系统菜单（仅更新非 nil 字段）
func (r *sysMenuRepository) UpdateSysMenu(ctx context.Context, req *bo.UpdateSysMenuBo) error {
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
	return db.Model(&do.SysMenu{}).Where("menu_code = ? AND deleted_at IS NULL", req.MenuCode).Updates(updates).Error
}

// DeleteSysMenu 删除系统菜单（软删除）
func (r *sysMenuRepository) DeleteSysMenu(ctx context.Context, menuCode string) error {
	db := getDBWithTransaction(ctx, r.db)
	return db.Model(&do.SysMenu{}).Where("menu_code = ? AND deleted_at IS NULL", menuCode).
		Update("deleted_at", time.Now()).Error
}
