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

// NewNmspacesMemberMenuRepository 创建租户成员菜单直授仓储实例
func NewNmspacesMemberMenuRepository(d *data.Data) repository.NmspacesMemberMenu {
	return &nmspacesMemberMenuRepository{db: d.DB()}
}

type nmspacesMemberMenuRepository struct {
	db *gorm.DB
}

// ListNmspacesMemberMenu 查询成员直授的菜单列表（JOIN moon_nmspaces_menus，按 menuSortOrder 升序）
func (r *nmspacesMemberMenuRepository) ListNmspacesMemberMenu(ctx context.Context, req *bo.ListNmspacesMemberMenuBo) ([]*bo.NmspacesMemberMenuItemBo, error) {
	db := getDBWithTransaction(ctx, r.db)

	joinSQL := `INNER JOIN moon_nmspaces_menus m ON moon_nmspaces_member_menus.menu_code = m.menu_code AND moon_nmspaces_member_menus.namespace_uid = m.namespace_uid`
	whereSQL := `moon_nmspaces_member_menus.namespace_uid = ? AND moon_nmspaces_member_menus.member_uid = ? AND m.deleted_at IS NULL`

	args := []interface{}{req.NamespaceUID, req.MemberUID}

	// 菜单状态筛选
	if req.MenuStatus != nil {
		whereSQL = fmt.Sprintf("%s AND m.menu_status = %d", whereSQL, *req.MenuStatus)
	}

	querySQL := fmt.Sprintf(`SELECT m.menu_code, m.parent_menu_code, m.menu_name, m.menu_status FROM moon_nmspaces_member_menus %s WHERE %s ORDER BY m.menu_sort_order ASC`, joinSQL, whereSQL)

	var rows []convert.NmspacesMemberMenuJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, args...).Scan(&rows).Error; err != nil {
		return nil, merr.ErrorInternalServer("list nmspaces member menus failed: %v", err)
	}

	items := make([]*bo.NmspacesMemberMenuItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.NmspacesMemberMenuJoinToBo(&rows[i]))
	}
	return items, nil
}

// GrantNmspacesMemberMenus 批量直授菜单（冲突时忽略，幂等）
func (r *nmspacesMemberMenuRepository) GrantNmspacesMemberMenus(ctx context.Context, req *bo.GrantNmspacesMemberMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.NmspacesMemberMenu, 0, len(req.MenuCodes))
	for _, menuCode := range req.MenuCodes {
		records = append(records, &do.NmspacesMemberMenu{
			NamespaceUID: req.NamespaceUID,
			MemberUID:    req.MemberUID,
			MenuCode:     menuCode,
			Creator:      req.Creator,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("grant nmspaces member menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// RevokeNmspacesMemberMenus 批量撤销直授菜单（硬删除）
func (r *nmspacesMemberMenuRepository) RevokeNmspacesMemberMenus(ctx context.Context, req *bo.RevokeNmspacesMemberMenusBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	result := db.WithContext(ctx).
		Where("namespace_uid = ? AND member_uid = ? AND menu_code IN ?", req.NamespaceUID, req.MemberUID, req.MenuCodes).
		Delete(&do.NmspacesMemberMenu{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("revoke nmspaces member menus failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// ExistsByNamespaceAndMemberUIDAndMenuCode 检查某菜单是否已直授给某成员
func (r *nmspacesMemberMenuRepository) ExistsByNamespaceAndMemberUIDAndMenuCode(ctx context.Context, namespaceUID, memberUID snowflake.ID, menuCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.NmspacesMemberMenu{}).
		Where("namespace_uid = ? AND member_uid = ? AND menu_code = ?", namespaceUID, memberUID, menuCode).
		Count(&count).Error; err != nil {
		return false, merr.ErrorInternalServer("check nmspaces member menu existence failed: %v", err)
	}
	return count > 0, nil
}
