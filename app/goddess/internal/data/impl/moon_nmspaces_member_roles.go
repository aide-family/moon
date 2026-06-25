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

// NewNmspacesMemberRoleRepository 创建租户成员角色仓储实例
func NewNmspacesMemberRoleRepository(d *data.Data) repository.NmspacesMemberRole {
	return &nmspacesMemberRoleRepository{db: d.DB()}
}

type nmspacesMemberRoleRepository struct {
	db *gorm.DB
}

// ListNmspacesMemberRole 分页查询成员的角色列表（JOIN moon_nmspaces_roles，过滤启用且未删除的角色）
func (r *nmspacesMemberRoleRepository) ListNmspacesMemberRole(ctx context.Context, req *bo.ListNmspacesMemberRoleBo) ([]*bo.NmspacesMemberRoleItemBo, int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	joinSQL := `INNER JOIN moon_nmspaces_roles r ON moon_nmspaces_member_roles.role_code = r.role_code AND moon_nmspaces_member_roles.namespace_uid = r.namespace_uid`
	whereSQL := `moon_nmspaces_member_roles.namespace_uid = ? AND moon_nmspaces_member_roles.member_uid = ? AND r.role_status = 1 AND r.deleted_at IS NULL`

	// 统计总数
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM moon_nmspaces_member_roles %s WHERE %s`, joinSQL, whereSQL)
	if err := db.WithContext(ctx).Raw(countSQL, req.NamespaceUID, req.MemberUID).Scan(&total).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("count nmspaces member roles failed: %v", err)
	}

	// 查询数据
	querySQL := fmt.Sprintf(`SELECT r.role_code, r.role_name, r.role_data_scope, r.role_status, moon_nmspaces_member_roles.created_at, moon_nmspaces_member_roles.grantor FROM moon_nmspaces_member_roles %s WHERE %s ORDER BY moon_nmspaces_member_roles.created_at DESC LIMIT ? OFFSET ?`, joinSQL, whereSQL)

	var rows []convert.NmspacesMemberRoleJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, req.NamespaceUID, req.MemberUID, req.PageRequestBo.Limit(), req.PageRequestBo.Offset()).Scan(&rows).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("list nmspaces member roles failed: %v", err)
	}

	items := make([]*bo.NmspacesMemberRoleItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.NmspacesMemberRoleJoinToBo(&rows[i]))
	}
	return items, total, nil
}

// BindNmspacesMemberRoles 批量绑定成员角色（冲突时忽略，幂等）
func (r *nmspacesMemberRoleRepository) BindNmspacesMemberRoles(ctx context.Context, req *bo.BindNmspacesMemberRolesBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.NmspacesMemberRole, 0, len(req.RoleCodes))
	for _, roleCode := range req.RoleCodes {
		records = append(records, &do.NmspacesMemberRole{
			NamespaceUID: req.NamespaceUID,
			MemberUID:    req.MemberUID,
			RoleCode:     roleCode,
			Grantor:      req.Grantor,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("bind nmspaces member roles failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// UnbindNmspacesMemberRoles 批量解绑成员角色（硬删除）
func (r *nmspacesMemberRoleRepository) UnbindNmspacesMemberRoles(ctx context.Context, req *bo.UnbindNmspacesMemberRolesBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	result := db.WithContext(ctx).
		Where("namespace_uid = ? AND member_uid = ? AND role_code IN ?", req.NamespaceUID, req.MemberUID, req.RoleCodes).
		Delete(&do.NmspacesMemberRole{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("unbind nmspaces member roles failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// CountByNamespaceAndMemberUID 统计某成员在当前租户拥有的角色数量
func (r *nmspacesMemberRoleRepository) CountByNamespaceAndMemberUID(ctx context.Context, namespaceUID, memberUID snowflake.ID) (int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.NmspacesMemberRole{}).
		Where("namespace_uid = ? AND member_uid = ?", namespaceUID, memberUID).
		Count(&count).Error; err != nil {
		return 0, merr.ErrorInternalServer("count nmspaces member roles failed: %v", err)
	}
	return count, nil
}

// CountByNamespaceAndMemberUIDAndRoleCodes 统计某成员在 roleCodes 中有几个角色
func (r *nmspacesMemberRoleRepository) CountByNamespaceAndMemberUIDAndRoleCodes(ctx context.Context, namespaceUID, memberUID snowflake.ID, roleCodes []string) (int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.NmspacesMemberRole{}).
		Where("namespace_uid = ? AND member_uid = ? AND role_code IN ?", namespaceUID, memberUID, roleCodes).
		Count(&count).Error; err != nil {
		return 0, merr.ErrorInternalServer("count nmspaces member roles by codes failed: %v", err)
	}
	return count, nil
}

// ListRoleCodesByNamespaceAndMemberUID 查询某成员在指定租户拥有的所有角色编码
func (r *nmspacesMemberRoleRepository) ListRoleCodesByNamespaceAndMemberUID(ctx context.Context, namespaceUID, memberUID snowflake.ID) ([]string, error) {
	db := getDBWithTransaction(ctx, r.db)

	var roleCodes []string
	if err := db.WithContext(ctx).Model(&do.NmspacesMemberRole{}).
		Where("namespace_uid = ? AND member_uid = ?", namespaceUID, memberUID).
		Pluck("role_code", &roleCodes).Error; err != nil {
		return nil, merr.ErrorInternalServer("list role codes by member failed: %v", err)
	}
	return roleCodes, nil
}
