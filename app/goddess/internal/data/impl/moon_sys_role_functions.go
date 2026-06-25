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

// NewSysRoleFunctionRepository 创建系统角色功能仓储实例
func NewSysRoleFunctionRepository(d *data.Data) repository.SysRoleFunction {
	return &sysRoleFunctionRepository{db: d.DB()}
}

type sysRoleFunctionRepository struct {
	db *gorm.DB
}

// ListRoleFunction 分页查询角色的功能列表（JOIN moon_sys_functions，按 functionSortOrder 升序）
func (r *sysRoleFunctionRepository) ListRoleFunction(ctx context.Context, req *bo.ListRoleFunctionBo) ([]*bo.RoleFunctionItemBo, int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	joinSQL := `INNER JOIN moon_sys_functions f ON moon_sys_role_functions.menu_code = f.menu_code AND moon_sys_role_functions.function_code = f.function_code`
	whereSQL := `moon_sys_role_functions.role_code = ? AND f.deleted_at IS NULL`

	args := []interface{}{req.RoleCode}

	// 按菜单编码过滤
	if req.MenuCode != "" {
		whereSQL = fmt.Sprintf("%s AND moon_sys_role_functions.menu_code = ?", whereSQL)
		args = append(args, req.MenuCode)
	}

	// 统计总数
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM moon_sys_role_functions %s WHERE %s`, joinSQL, whereSQL)
	if err := db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("count role functions failed: %v", err)
	}

	// 查询数据
	querySQL := fmt.Sprintf(`SELECT f.menu_code, f.function_code, f.function_name, f.function_sort_order, f.function_status, moon_sys_role_functions.creator, moon_sys_role_functions.created_at FROM moon_sys_role_functions %s WHERE %s ORDER BY f.function_sort_order ASC, f.created_at DESC`, joinSQL, whereSQL)

	if req.PageRequestBo.Page > 0 && req.PageRequestBo.PageSize > 0 {
		querySQL = fmt.Sprintf("%s LIMIT ? OFFSET ?", querySQL)
		args = append(args, req.PageRequestBo.Limit(), req.PageRequestBo.Offset())
	}

	var rows []convert.RoleFunctionJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, args...).Scan(&rows).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("list role functions failed: %v", err)
	}

	items := make([]*bo.RoleFunctionItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.RoleFunctionJoinToBo(&rows[i]))
	}
	return items, total, nil
}

// GrantRoleFunctions 批量授予功能（冲突时忽略，幂等）
func (r *sysRoleFunctionRepository) GrantRoleFunctions(ctx context.Context, req *bo.GrantRoleFunctionsBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.SysRoleFunction, 0, len(req.FunctionKeys))
	for _, fk := range req.FunctionKeys {
		records = append(records, &do.SysRoleFunction{
			RoleCode:     req.RoleCode,
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
			Creator:      req.Creator,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("grant role functions failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// RevokeRoleFunctions 批量撤销功能（硬删除）
func (r *sysRoleFunctionRepository) RevokeRoleFunctions(ctx context.Context, req *bo.RevokeRoleFunctionsBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	// 构建 OR 条件列表
	conditions := make([]string, 0, len(req.FunctionKeys))
	args := make([]interface{}, 0, len(req.FunctionKeys)*2+1)
	args = append(args, req.RoleCode)
	for _, fk := range req.FunctionKeys {
		conditions = append(conditions, "(menu_code = ? AND function_code = ?)")
		args = append(args, fk.MenuCode, fk.FunctionCode)
	}

	whereSQL := fmt.Sprintf("role_code = ? AND (%s)", joinOr(conditions))
	result := db.WithContext(ctx).
		Where(whereSQL, args...).
		Delete(&do.SysRoleFunction{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("revoke role functions failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// CountByRoleCodeAndMenuCode 统计某角色在某菜单下拥有的功能数量
func (r *sysRoleFunctionRepository) CountByRoleCodeAndMenuCode(ctx context.Context, roleCode, menuCode string) (int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.SysRoleFunction{}).
		Where("role_code = ? AND menu_code = ?", roleCode, menuCode).
		Count(&count).Error; err != nil {
		return 0, merr.ErrorInternalServer("count role functions failed: %v", err)
	}
	return count, nil
}

// ListFunctionsByRoleCodes 查询多个角色拥有的所有功能（JOIN moon_sys_functions，去重）
func (r *sysRoleFunctionRepository) ListFunctionsByRoleCodes(ctx context.Context, roleCodes []string) ([]*bo.RoleFunctionItemBo, error) {
	db := getDBWithTransaction(ctx, r.db)

	if len(roleCodes) == 0 {
		return nil, nil
	}

	joinSQL := `INNER JOIN moon_sys_functions f ON moon_sys_role_functions.menu_code = f.menu_code AND moon_sys_role_functions.function_code = f.function_code`
	whereSQL := `moon_sys_role_functions.role_code IN ? AND f.deleted_at IS NULL`

	querySQL := fmt.Sprintf(`SELECT DISTINCT f.menu_code, f.function_code, f.function_name, f.function_sort_order, f.function_status FROM moon_sys_role_functions %s WHERE %s ORDER BY f.function_sort_order ASC`, joinSQL, whereSQL)

	var rows []convert.RoleFunctionJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, roleCodes).Scan(&rows).Error; err != nil {
		return nil, merr.ErrorInternalServer("list functions by role codes failed: %v", err)
	}

	items := make([]*bo.RoleFunctionItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.RoleFunctionJoinToBo(&rows[i]))
	}
	return items, nil
}

// joinOr 将多个条件用 OR 连接
func joinOr(conditions []string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += " OR "
		}
		result += c
	}
	return result
}
