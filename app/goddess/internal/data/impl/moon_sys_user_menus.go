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

// NewSysUserMenuRepository 创建系统用户菜单直授仓储实例
func NewSysUserMenuRepository(d *data.Data) repository.SysUserMenu {
	return &sysUserMenuRepository{db: d.DB()}
}

type sysUserMenuRepository struct {
	db *gorm.DB
}

// ListUserMenu 查询用户直授的菜单列表（JOIN moon_sys_menus，按 menuSortOrder 升序）
func (r *sysUserMenuRepository) ListUserMenu(ctx context.Context, req *bo.ListUserMenuBo) ([]*bo.UserMenuItemBo, error) {
	db := getDBWithTransaction(ctx, r.db)

	joinSQL := `INNER JOIN moon_sys_menus m ON moon_sys_user_menus.menu_code = m.menu_code`
	whereSQL := `moon_sys_user_menus.user_uid = ? AND m.deleted_at IS NULL`

	args := []interface{}{req.UserUID}

	// 菜单状态筛选
	if req.MenuStatus != nil {
		whereSQL = fmt.Sprintf("%s AND m.menu_status = %d", whereSQL, *req.MenuStatus)
	}

	querySQL := fmt.Sprintf(`SELECT m.menu_code, m.parent_menu_code, m.menu_name, m.menu_status FROM moon_sys_user_menus %s WHERE %s ORDER BY m.menu_sort_order ASC`, joinSQL, whereSQL)

	var rows []convert.UserMenuJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, args...).Scan(&rows).Error; err != nil {
		return nil, merr.ErrorInternalServer("list user menus failed: %v", err)
	}

	items := make([]*bo.UserMenuItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.UserMenuJoinToBo(&rows[i]))
	}
	return items, nil
}

// GrantUserMenus 批量直授菜单（冲突时忽略，幂等）
func (r *sysUserMenuRepository) GrantUserMenus(ctx context.Context, req *bo.GrantUserMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.SysUserMenu, 0, len(req.MenuCodes))
	for _, menuCode := range req.MenuCodes {
		records = append(records, &do.SysUserMenu{
			UserUID:  req.UserUID,
			MenuCode: menuCode,
			Creator:  req.Creator,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("grant user menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// RevokeUserMenus 批量撤销直授菜单（硬删除）
func (r *sysUserMenuRepository) RevokeUserMenus(ctx context.Context, req *bo.RevokeUserMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	result := db.WithContext(ctx).
		Where("user_uid = ? AND menu_code IN ?", req.UserUID, req.MenuCodes).
		Delete(&do.SysUserMenu{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("revoke user menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// ExistsByUserUIDAndMenuCode 检查某菜单是否已直授给某用户
func (r *sysUserMenuRepository) ExistsByUserUIDAndMenuCode(ctx context.Context, userUID snowflake.ID, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.SysUserMenu{}).
		Where("user_uid = ? AND menu_code = ?", userUID, menuCode).
		Count(&count).Error; err != nil {
		return false, merr.ErrorInternalServer("check user menu existence failed: %v", err)
	}
	return count > 0, nil
}
