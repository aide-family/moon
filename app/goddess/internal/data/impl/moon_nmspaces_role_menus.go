package impl

import (
	"context"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// NewNmspacesRoleMenuRepository 创建租户角色菜单仓储实例
func NewNmspacesRoleMenuRepository(d *data.Data) repository.NmspacesRoleMenu {
	return &nmspacesRoleMenuRepository{db: d.DB()}
}

type nmspacesRoleMenuRepository struct {
	db *gorm.DB
}

// ListNmspacesRoleMenu 查询角色的菜单列表（JOIN moon_nmspaces_menus，按 menuSortOrder 升序）
func (r *nmspacesRoleMenuRepository) ListNmspacesRoleMenu(ctx context.Context, req *bo.ListNmspacesRoleMenuBo) ([]*bo.NmspacesRoleMenuItemBo, error) {
	db := getDBWithTransaction(ctx, r.db)

	joinSQL := `INNER JOIN moon_nmspaces_menus m ON moon_nmspaces_role_menus.menu_code = m.menu_code AND moon_nmspaces_role_menus.namespace_uid = m.namespace_uid`
	whereSQL := `moon_nmspaces_role_menus.namespace_uid = ? AND moon_nmspaces_role_menus.role_code = ? AND m.deleted_at IS NULL`

	args := []interface{}{req.NamespaceUID, req.RoleCode}

	// 菜单状态筛选
	if req.MenuStatus != nil {
		whereSQL = fmt.Sprintf("%s AND m.menu_status = %d", whereSQL, *req.MenuStatus)
	}

	querySQL := fmt.Sprintf(`SELECT m.menu_code, m.parent_menu_code, m.menu_name, m.menu_status FROM moon_nmspaces_role_menus %s WHERE %s ORDER BY m.menu_sort_order ASC`, joinSQL, whereSQL)

	var rows []convert.NmspacesRoleMenuJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, args...).Scan(&rows).Error; err != nil {
		return nil, merr.ErrorInternalServer("list nmspaces role menus failed: %v", err)
	}

	items := make([]*bo.NmspacesRoleMenuItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.NmspacesRoleMenuJoinToBo(&rows[i]))
	}
	return items, nil
}

// GrantNmspacesRoleMenus 批量授予菜单（冲突时忽略，幂等）
func (r *nmspacesRoleMenuRepository) GrantNmspacesRoleMenus(ctx context.Context, req *bo.GrantNmspacesRoleMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.NmspacesRoleMenu, 0, len(req.MenuCodes))
	for _, menuCode := range req.MenuCodes {
		records = append(records, &do.NmspacesRoleMenu{
			NamespaceUID: req.NamespaceUID,
			RoleCode:     req.RoleCode,
			MenuCode:     menuCode,
			Creator:      req.Creator,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("grant nmspaces role menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// RevokeNmspacesRoleMenus 批量撤销菜单（硬删除）
func (r *nmspacesRoleMenuRepository) RevokeNmspacesRoleMenus(ctx context.Context, req *bo.RevokeNmspacesRoleMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	result := db.WithContext(ctx).
		Where("namespace_uid = ? AND role_code = ? AND menu_code IN ?", req.NamespaceUID, req.RoleCode, req.MenuCodes).
		Delete(&do.NmspacesRoleMenu{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("revoke nmspaces role menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// ExistsByNamespaceAndRoleCodeAndMenuCode 检查某菜单是否已授权给某角色
func (r *nmspacesRoleMenuRepository) ExistsByNamespaceAndRoleCodeAndMenuCode(ctx context.Context, namespaceUID snowflake.ID, roleCode, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.NmspacesRoleMenu{}).
		Where("namespace_uid = ? AND role_code = ? AND menu_code = ?", namespaceUID, roleCode, menuCode).
		Count(&count).Error; err != nil {
		return false, merr.ErrorInternalServer("check nmspaces role menu existence failed: %v", err)
	}
	return count > 0, nil
}

// ListMenuCodesByNamespaceAndRoleCodes 查询多个角色拥有的所有菜单编码（去重）
func (r *nmspacesRoleMenuRepository) ListMenuCodesByNamespaceAndRoleCodes(ctx context.Context, namespaceUID snowflake.ID, roleCodes []string) ([]string, error) {
	db := getDBWithTransaction(ctx, r.db)

	if len(roleCodes) == 0 {
		return nil, nil
	}

	var menuCodes []string
	if err := db.WithContext(ctx).Model(&do.NmspacesRoleMenu{}).
		Where("namespace_uid = ? AND role_code IN ?", namespaceUID, roleCodes).
		Distinct().
		Pluck("menu_code", &menuCodes).Error; err != nil {
		return nil, merr.ErrorInternalServer("list menu codes by role codes failed: %v", err)
	}
	return menuCodes, nil
}
