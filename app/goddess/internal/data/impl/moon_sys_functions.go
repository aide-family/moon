package impl

import (
	"context"

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

// NewSysFunctionRepository 创建系统功能仓储实例
func NewSysFunctionRepository(d *data.Data) repository.SysFunction {
	return NewSysFunctionRepositoryWithDB(d.DB())
}

// NewSysFunctionRepositoryWithDB 使用指定数据库连接创建系统功能仓储实例
func NewSysFunctionRepositoryWithDB(db *gorm.DB) repository.SysFunction {
	query.SetDefault(db)
	return &sysFunctionRepository{db: db}
}

type sysFunctionRepository struct {
	db *gorm.DB
}

// CreateSysFunction 新增系统功能
func (r *sysFunctionRepository) CreateSysFunction(ctx context.Context, req *bo.CreateSysFunctionBo) (*bo.SysFunctionBo, error) {
	db := getDBWithTransaction(ctx, r.db)
	now := time.Now().Format(time.DateTime)
	fn := &do.SysFunction{
		MenuCode:          req.MenuCode,
		FunctionCode:      req.FunctionCode,
		FunctionName:      req.FunctionName,
		FunctionSortOrder: req.FunctionSortOrder,
		FunctionStatus:    req.FunctionStatus,
		FunctionRemark:    req.FunctionRemark,
		Creator:           req.Creator,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(fn).Error; err != nil {
		return nil, err
	}
	return convert.SysFunctionToBo(fn), nil
}

// ListSysFunction 分页查询系统功能列表
func (r *sysFunctionRepository) ListSysFunction(ctx context.Context, req *bo.ListSysFunctionBo) ([]*bo.SysFunctionBo, int64, error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysFunction
	wrappers := mutation.WithContext(ctx)

	// menuCode 筛选（非必传）
	if strutil.IsNotEmpty(req.MenuCode) {
		wrappers = wrappers.Where(mutation.MenuCode.Eq(req.MenuCode))
	}

	// 排除软删除数据
	wrappers = wrappers.Where(mutation.DeletedAt.IsNull())

	// 关键字模糊查询（功能编码 OR 功能名称）
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.FunctionCode.Like(keyword)).
			Or(mutation.FunctionName.Like(keyword))
	}
	// 功能状态筛选
	if req.FunctionStatus != nil {
		wrappers = wrappers.Where(mutation.FunctionStatus.Eq(*req.FunctionStatus))
	}

	// 统计总数
	var total int64
	if req.PageRequestBo.Page > 0 && req.PageRequestBo.PageSize > 0 {
		var err error
		total, err = wrappers.Count()
		if err != nil {
			return nil, 0, merr.ErrorInternalServer("list sys function count failed: %v", err)
		}
	}

	// 排序：默认按 functionSortOrder 升序，createdAt 降序
	wrappers = wrappers.Order(mutation.FunctionSortOrder.Asc(), mutation.CreatedAt.Desc())

	if req.PageRequestBo.Page > 0 && req.PageRequestBo.PageSize > 0 {
		wrappers = wrappers.Limit(int(req.PageRequestBo.PageSize)).Offset(int((req.PageRequestBo.Page - 1) * req.PageRequestBo.PageSize))
	}

	fns, err := wrappers.Find()
	if err != nil {
		return nil, 0, merr.ErrorInternalServer("list sys function failed: %v", err)
	}

	items := make([]*bo.SysFunctionBo, 0, len(fns))
	for _, fn := range fns {
		items = append(items, convert.SysFunctionToBo(fn))
	}
	return items, total, nil
}

// GetSysFunction 根据联合主键查询功能详情
func (r *sysFunctionRepository) GetSysFunction(ctx context.Context, menuCode, functionCode string) (*bo.SysFunctionBo, error) {
	mutation := query.Use(getDBWithTransaction(ctx, r.db)).SysFunction
	fn, err := mutation.WithContext(ctx).
		Where(mutation.MenuCode.Eq(menuCode), mutation.FunctionCode.Eq(functionCode)).
		Where(mutation.DeletedAt.IsNull()).
		First()
	if err != nil {
		return nil, merr.ErrorInternalServer("get sys function failed: %v", err)
	}
	if fn == nil {
		return nil, nil
	}
	return convert.SysFunctionToBo(fn), nil
}

// UpdateSysFunction 编辑系统功能（仅更新非 nil 字段）
func (r *sysFunctionRepository) UpdateSysFunction(ctx context.Context, req *bo.UpdateSysFunctionBo) (*bo.SysFunctionBo, error) {
	db := getDBWithTransaction(ctx, r.db)
	updates := make(map[string]interface{})
	if req.FunctionName != nil {
		updates["function_name"] = *req.FunctionName
	}
	if req.FunctionSortOrder != nil {
		updates["function_sort_order"] = *req.FunctionSortOrder
	}
	if req.FunctionStatus != nil {
		updates["function_status"] = *req.FunctionStatus
	}
	if req.FunctionRemark != nil {
		updates["function_remark"] = *req.FunctionRemark
	}
	// 手动更新 updated_at
	updates["updated_at"] = time.Now().Format(time.DateTime)
	if len(updates) == 0 {
		// 无更新字段，直接查询返回
		return r.GetSysFunction(ctx, req.MenuCode, req.FunctionCode)
	}
	if err := db.Model(&do.SysFunction{}).
		Where("menu_code = ? AND function_code = ? AND deleted_at IS NULL", req.MenuCode, req.FunctionCode).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetSysFunction(ctx, req.MenuCode, req.FunctionCode)
}

// DeleteSysFunction 删除系统功能（软删除）
func (r *sysFunctionRepository) DeleteSysFunction(ctx context.Context, menuCode, functionCode string) error {
	db := getDBWithTransaction(ctx, r.db)
	return db.Model(&do.SysFunction{}).
		Where("menu_code = ? AND function_code = ? AND deleted_at IS NULL", menuCode, functionCode).
		Update("deleted_at", time.Now().Format(time.DateTime)).Error
}

// ExistsByMenuCodeAndFunctionCode 检查联合主键是否已存在（含软删除行）
func (r *sysFunctionRepository) ExistsByMenuCodeAndFunctionCode(ctx context.Context, menuCode, functionCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysFunction{}).Unscoped().
		Where("menu_code = ? AND function_code = ?", menuCode, functionCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByMenuCodeAndFunctionCodeNotDeleted 检查联合主键在未删除行中是否存在
func (r *sysFunctionRepository) ExistsByMenuCodeAndFunctionCodeNotDeleted(ctx context.Context, menuCode, functionCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysFunction{}).
		Where("menu_code = ? AND function_code = ? AND deleted_at IS NULL", menuCode, functionCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByMenuCodeAndFunctionName 检查同一 menuCode 下 functionName 是否已存在（未删除行）
func (r *sysFunctionRepository) ExistsByMenuCodeAndFunctionName(ctx context.Context, menuCode, functionName string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.SysFunction{}).
		Where("menu_code = ? AND function_name = ? AND deleted_at IS NULL", menuCode, functionName).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
