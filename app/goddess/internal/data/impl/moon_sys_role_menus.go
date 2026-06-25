package impl

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// NewSysRoleMenuRepository 创建系统角色菜单仓储实例
func NewSysRoleMenuRepository(d *data.Data) repository.SysRoleMenu {
	return &sysRoleMenuRepository{db: d.DB()}
}

type sysRoleMenuRepository struct {
	db *gorm.DB
}

// ListRoleMenu 查询角色的菜单列表（JOIN moon_sys_menus，按 menuSortOrder 升序）
func (r *sysRoleMenuRepository) ListRoleMenu(ctx context.Context, req *bo.ListRoleMenuBo) ([]*bo.RoleMenuItemBo, error) {
	db := getDBWithTransaction(ctx, r.db)

	joinSQL := `INNER JOIN moon_sys_menus m ON moon_sys_role_menus.menu_code = m.menu_code`
	whereSQL := `moon_sys_role_menus.role_code = ? AND m.deleted_at IS NULL`

	// 菜单状态筛选
	if req.MenuStatus != nil {
		whereSQL = fmt.Sprintf("%s AND m.menu_status = %d", whereSQL, *req.MenuStatus)
	}

	querySQL := fmt.Sprintf(`SELECT m.menu_code, m.parent_menu_code, m.menu_name, m.menu_status FROM moon_sys_role_menus %s WHERE %s ORDER BY m.menu_sort_order ASC`, joinSQL, whereSQL)

	var rows []convert.RoleMenuJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, req.RoleCode).Scan(&rows).Error; err != nil {
		return nil, merr.ErrorInternalServer("list role menus failed: %v", err)
	}

	items := make([]*bo.RoleMenuItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.RoleMenuJoinToBo(&rows[i]))
	}
	return items, nil
}

// GrantRoleMenus 批量授予菜单（冲突时忽略，幂等）
func (r *sysRoleMenuRepository) GrantRoleMenus(ctx context.Context, req *bo.GrantRoleMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.SysRoleMenu, 0, len(req.MenuCodes))
	for _, menuCode := range req.MenuCodes {
		records = append(records, &do.SysRoleMenu{
			RoleCode: req.RoleCode,
			MenuCode: menuCode,
			Creator:  req.Creator,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("grant role menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// RevokeRoleMenus 批量撤销菜单（硬删除）
func (r *sysRoleMenuRepository) RevokeRoleMenus(ctx context.Context, req *bo.RevokeRoleMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	result := db.WithContext(ctx).
		Where("role_code = ? AND menu_code IN ?", req.RoleCode, req.MenuCodes).
		Delete(&do.SysRoleMenu{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("revoke role menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// ExistsByRoleCodeAndMenuCode 检查某菜单是否已授权给某角色
func (r *sysRoleMenuRepository) ExistsByRoleCodeAndMenuCode(ctx context.Context, roleCode, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.SysRoleMenu{}).
		Where("role_code = ? AND menu_code = ?", roleCode, menuCode).
		Count(&count).Error; err != nil {
		return false, merr.ErrorInternalServer("check role menu existence failed: %v", err)
	}
	return count > 0, nil
}

// ListMenuCodesByRoleCodes 查询多个角色拥有的所有菜单编码（去重）
func (r *sysRoleMenuRepository) ListMenuCodesByRoleCodes(ctx context.Context, roleCodes []string) ([]string, error) {
	db := getDBWithTransaction(ctx, r.db)

	if len(roleCodes) == 0 {
		return nil, nil
	}

	var menuCodes []string
	if err := db.WithContext(ctx).Model(&do.SysRoleMenu{}).
		Where("role_code IN ?", roleCodes).
		Distinct().
		Pluck("menu_code", &menuCodes).Error; err != nil {
		return nil, merr.ErrorInternalServer("list menu codes by role codes failed: %v", err)
	}
	return menuCodes, nil
}
