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

// NewSysUserRoleRepository 创建系统用户角色仓储实例
func NewSysUserRoleRepository(d *data.Data) repository.SysUserRole {
	return &sysUserRoleRepository{db: d.DB()}
}

type sysUserRoleRepository struct {
	db *gorm.DB
}

// ListSysUserRole 分页查询用户的系统角色列表（JOIN moon_sys_roles，过滤启用且未删除的角色）
func (r *sysUserRoleRepository) ListSysUserRole(ctx context.Context, req *bo.ListSysUserRoleBo) ([]*bo.SysUserRoleItemBo, int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	// JOIN 查询模板
	joinSQL := `INNER JOIN moon_sys_roles r ON moon_sys_user_roles.role_code = r.role_code`
	whereSQL := `moon_sys_user_roles.user_uid = ? AND r.role_status = 1 AND r.deleted_at IS NULL`

	// 统计总数
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM moon_sys_user_roles %s WHERE %s`, joinSQL, whereSQL)
	if err := db.WithContext(ctx).Raw(countSQL, req.UserUID).Scan(&total).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("count sys user roles failed: %v", err)
	}

	// 查询数据
	querySQL := fmt.Sprintf(`SELECT r.role_code, r.role_name, r.role_data_scope, r.role_status, moon_sys_user_roles.created_at, moon_sys_user_roles.creator FROM moon_sys_user_roles %s WHERE %s ORDER BY moon_sys_user_roles.created_at DESC LIMIT ? OFFSET ?`, joinSQL, whereSQL)

	var rows []convert.SysUserRoleJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, req.UserUID, req.PageRequestBo.Limit(), req.PageRequestBo.Offset()).Scan(&rows).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("list sys user roles failed: %v", err)
	}

	items := make([]*bo.SysUserRoleItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.SysUserRoleJoinToBo(&rows[i]))
	}
	return items, total, nil
}

// BindSysUserRoles 批量绑定用户角色（冲突时忽略，幂等）
func (r *sysUserRoleRepository) BindSysUserRoles(ctx context.Context, req *bo.BindSysUserRolesBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.SysUserRole, 0, len(req.RoleCodes))
	for _, roleCode := range req.RoleCodes {
		records = append(records, &do.SysUserRole{
			UserUID:  req.UserUID,
			RoleCode: roleCode,
			Creator:  req.Creator,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("bind sys user roles failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// UnbindSysUserRoles 批量解绑用户角色（硬删除）
func (r *sysUserRoleRepository) UnbindSysUserRoles(ctx context.Context, req *bo.UnbindSysUserRolesBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	result := db.WithContext(ctx).
		Where("user_uid = ? AND role_code IN ?", req.UserUID, req.RoleCodes).
		Delete(&do.SysUserRole{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("unbind sys user roles failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// CountByUserUID 统计某用户当前拥有的角色数量
func (r *sysUserRoleRepository) CountByUserUID(ctx context.Context, userUID snowflake.ID) (int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.SysUserRole{}).
		Where("user_uid = ?", userUID).
		Count(&count).Error; err != nil {
		return 0, merr.ErrorInternalServer("count sys user roles failed: %v", err)
	}
	return count, nil
}

// ListRoleCodesByUserUID 查询某用户拥有的所有角色编码
func (r *sysUserRoleRepository) ListRoleCodesByUserUID(ctx context.Context, userUID snowflake.ID) ([]string, error) {
	db := getDBWithTransaction(ctx, r.db)

	var roleCodes []string
	if err := db.WithContext(ctx).Model(&do.SysUserRole{}).
		Where("user_uid = ?", userUID).
		Pluck("role_code", &roleCodes).Error; err != nil {
		return nil, merr.ErrorInternalServer("list role codes by user failed: %v", err)
	}
	return roleCodes, nil
}

// CountByUserUIDAndRoleCodes 统计某用户在 roleCodes 中有几个角色
func (r *sysUserRoleRepository) CountByUserUIDAndRoleCodes(ctx context.Context, userUID snowflake.ID, roleCodes []string) (int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.SysUserRole{}).
		Where("user_uid = ? AND role_code IN ?", userUID, roleCodes).
		Count(&count).Error; err != nil {
		return 0, merr.ErrorInternalServer("count sys user roles by codes failed: %v", err)
	}
	return count, nil
}
