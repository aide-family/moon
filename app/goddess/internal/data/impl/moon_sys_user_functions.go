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

// NewSysUserFunctionRepository 创建系统用户功能直授仓储实例
func NewSysUserFunctionRepository(d *data.Data) repository.SysUserFunction {
	return &sysUserFunctionRepository{db: d.DB()}
}

type sysUserFunctionRepository struct {
	db *gorm.DB
}

// ListUserFunction 分页查询用户直授的功能列表（JOIN moon_sys_functions，按 functionSortOrder 升序）
func (r *sysUserFunctionRepository) ListUserFunction(ctx context.Context, req *bo.ListUserFunctionBo) ([]*bo.UserFunctionItemBo, int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	joinSQL := `INNER JOIN moon_sys_functions f ON moon_sys_user_functions.menu_code = f.menu_code AND moon_sys_user_functions.function_code = f.function_code`
	whereSQL := `moon_sys_user_functions.user_uid = ? AND f.deleted_at IS NULL`

	args := []interface{}{req.UserUID}

	// 按菜单编码过滤
	if req.MenuCode != "" {
		whereSQL = fmt.Sprintf("%s AND moon_sys_user_functions.menu_code = ?", whereSQL)
		args = append(args, req.MenuCode)
	}

	// 统计总数
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM moon_sys_user_functions %s WHERE %s`, joinSQL, whereSQL)
	if err := db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("count user functions failed: %v", err)
	}

	// 查询数据
	querySQL := fmt.Sprintf(`SELECT f.menu_code, f.function_code, f.function_name, f.function_sort_order, f.function_status, moon_sys_user_functions.creator, moon_sys_user_functions.created_at FROM moon_sys_user_functions %s WHERE %s ORDER BY f.function_sort_order ASC, f.created_at DESC`, joinSQL, whereSQL)

	if req.PageRequestBo.Page > 0 && req.PageRequestBo.PageSize > 0 {
		querySQL = fmt.Sprintf("%s LIMIT ? OFFSET ?", querySQL)
		args = append(args, req.PageRequestBo.Limit(), req.PageRequestBo.Offset())
	}

	var rows []convert.UserFunctionJoinRow
	if err := db.WithContext(ctx).Raw(querySQL, args...).Scan(&rows).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("list user functions failed: %v", err)
	}

	items := make([]*bo.UserFunctionItemBo, 0, len(rows))
	for i := range rows {
		items = append(items, convert.UserFunctionJoinToBo(&rows[i]))
	}
	return items, total, nil
}

// GrantUserFunctions 批量直授功能（冲突时忽略，幂等）
func (r *sysUserFunctionRepository) GrantUserFunctions(ctx context.Context, req *bo.GrantUserFunctionsBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	records := make([]*do.SysUserFunction, 0, len(req.FunctionKeys))
	for _, fk := range req.FunctionKeys {
		records = append(records, &do.SysUserFunction{
			UserUID:      req.UserUID,
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
			Creator:      req.Creator,
		})
	}

	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("grant user functions failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// RevokeUserFunctions 批量撤销直授功能（硬删除）
func (r *sysUserFunctionRepository) RevokeUserFunctions(ctx context.Context, req *bo.RevokeUserFunctionsBo) (int32, error) {
	db := getDBWithTransaction(ctx, r.db)

	// 构建 OR 条件列表
	conditions := make([]string, 0, len(req.FunctionKeys))
	args := make([]interface{}, 0, len(req.FunctionKeys)*2+1)
	args = append(args, req.UserUID)
	for _, fk := range req.FunctionKeys {
		conditions = append(conditions, "(menu_code = ? AND function_code = ?)")
		args = append(args, fk.MenuCode, fk.FunctionCode)
	}

	whereSQL := fmt.Sprintf("user_uid = ? AND (%s)", joinOr(conditions))
	result := db.WithContext(ctx).
		Where(whereSQL, args...).
		Delete(&do.SysUserFunction{})
	if result.Error != nil {
		return 0, merr.ErrorInternalServer("revoke user functions failed: %v", result.Error)
	}
	return int32(result.RowsAffected), nil
}

// CountByUserUIDAndMenuCode 统计某用户在某菜单下拥有的直授功能数量
func (r *sysUserFunctionRepository) CountByUserUIDAndMenuCode(ctx context.Context, userUID snowflake.ID, menuCode string) (int64, error) {
	db := getDBWithTransaction(ctx, r.db)

	var count int64
	if err := db.WithContext(ctx).Model(&do.SysUserFunction{}).
		Where("user_uid = ? AND menu_code = ?", userUID, menuCode).
		Count(&count).Error; err != nil {
		return 0, merr.ErrorInternalServer("count user functions failed: %v", err)
	}
	return count, nil
}
