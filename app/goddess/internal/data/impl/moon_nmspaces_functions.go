package impl

import (
	"context"

	"time"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// NewNmspacesFunctionRepository 创建租户功能仓储实例
func NewNmspacesFunctionRepository(d *data.Data) repository.NmspacesFunction {
	return NewNmspacesFunctionRepositoryWithDB(d.DB())
}

// NewNmspacesFunctionRepositoryWithDB 使用指定数据库连接创建租户功能仓储实例
func NewNmspacesFunctionRepositoryWithDB(db *gorm.DB) repository.NmspacesFunction {
	return &nmspacesFunctionRepository{db: db}
}

type nmspacesFunctionRepository struct {
	db *gorm.DB
}

// CreateNmspacesFunction 新增租户功能
func (r *nmspacesFunctionRepository) CreateNmspacesFunction(ctx context.Context, req *bo.CreateNmspacesFunctionBo) (*bo.NmspacesFunctionBo, error) {
	db := getDBWithTransaction(ctx, r.db)
	fn := &do.NmspacesFunction{
		NamespaceUID:      req.NamespaceUID,
		MenuCode:          req.MenuCode,
		FunctionCode:      req.FunctionCode,
		FunctionName:      req.FunctionName,
		FunctionSortOrder: req.FunctionSortOrder,
		FunctionStatus:    req.FunctionStatus,
		FunctionRemark:    req.FunctionRemark,
		Creator:           req.Creator,
	}
	if err := db.Create(fn).Error; err != nil {
		return nil, err
	}
	return convert.NmspacesFunctionToBo(fn), nil
}

// ListNmspacesFunction 分页查询租户功能列表
func (r *nmspacesFunctionRepository) ListNmspacesFunction(ctx context.Context, req *bo.ListNmspacesFunctionBo) ([]*bo.NmspacesFunctionBo, int64, error) {
	db := getDBWithTransaction(ctx, r.db).WithContext(ctx).
		Where("namespace_uid = ?", req.NamespaceUID)

	// menuCode
	if strutil.IsNotEmpty(req.MenuCode) {
		db = db.Where("menu_code = ?", req.MenuCode)
	}

	// 关键字模糊查询（功能编码 OR 功能名称）
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("function_code LIKE ? OR function_name LIKE ?", keyword, keyword)
	}
	// 功能状态筛选
	if req.FunctionStatus != nil {
		db = db.Where("function_status = ?", *req.FunctionStatus)
	}

	// 统计总数
	var total int64
	if req.PageRequestBo.Page > 0 && req.PageRequestBo.PageSize > 0 {
		if err := db.Model(&do.NmspacesFunction{}).Count(&total).Error; err != nil {
			return nil, 0, merr.ErrorInternalServer("list nmspaces function count failed: %v", err)
		}
	}

	// 排序
	db = db.Order("function_sort_order ASC, created_at DESC")

	if req.PageRequestBo.Page > 0 && req.PageRequestBo.PageSize > 0 {
		db = db.Limit(int(req.PageRequestBo.PageSize)).Offset(int((req.PageRequestBo.Page - 1) * req.PageRequestBo.PageSize))
	}

	var fns []*do.NmspacesFunction
	if err := db.Find(&fns).Error; err != nil {
		return nil, 0, merr.ErrorInternalServer("list nmspaces function failed: %v", err)
	}

	items := make([]*bo.NmspacesFunctionBo, 0, len(fns))
	for _, fn := range fns {
		items = append(items, convert.NmspacesFunctionToBo(fn))
	}
	return items, total, nil
}

// GetNmspacesFunction 根据联合主键查询功能详情
func (r *nmspacesFunctionRepository) GetNmspacesFunction(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) (*bo.NmspacesFunctionBo, error) {
	db := getDBWithTransaction(ctx, r.db).WithContext(ctx)
	var fn do.NmspacesFunction
	if err := db.Where("namespace_uid = ? AND menu_code = ? AND function_code = ?", namespaceUID, menuCode, functionCode).First(&fn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, merr.ErrorInternalServer("get nmspaces function failed: %v", err)
	}
	return convert.NmspacesFunctionToBo(&fn), nil
}

// UpdateNmspacesFunction 编辑租户功能（仅更新功能状态字段，并记录更新时间）
func (r *nmspacesFunctionRepository) UpdateNmspacesFunction(ctx context.Context, req *bo.UpdateNmspacesFunctionBo) (*bo.NmspacesFunctionBo, error) {
	db := getDBWithTransaction(ctx, r.db)
	updates := make(map[string]interface{})
	if req.FunctionStatus != nil {
		updates["function_status"] = *req.FunctionStatus
	}
	if len(updates) == 0 {
		return r.GetNmspacesFunction(ctx, req.NamespaceUID, req.MenuCode, req.FunctionCode)
	}
	updates["updated_at"] = time.Now()
	if err := db.Model(&do.NmspacesFunction{}).
		Where("namespace_uid = ? AND menu_code = ? AND function_code = ? AND deleted_at IS NULL", req.NamespaceUID, req.MenuCode, req.FunctionCode).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetNmspacesFunction(ctx, req.NamespaceUID, req.MenuCode, req.FunctionCode)
}

// DeleteNmspacesFunction 删除租户功能（软删除）
func (r *nmspacesFunctionRepository) DeleteNmspacesFunction(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) error {
	db := getDBWithTransaction(ctx, r.db)
	return db.Model(&do.NmspacesFunction{}).
		Where("namespace_uid = ? AND menu_code = ? AND function_code = ? AND deleted_at IS NULL", namespaceUID, menuCode, functionCode).
		Update("deleted_at", time.Now()).Error
}

// ExistsByNamespaceAndMenuCodeAndFunctionCode 检查联合主键是否已存在（含软删除行）
func (r *nmspacesFunctionRepository) ExistsByNamespaceAndMenuCodeAndFunctionCode(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesFunction{}).Unscoped().
		Where("namespace_uid = ? AND menu_code = ? AND function_code = ?", namespaceUID, menuCode, functionCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamespaceAndMenuCodeAndFunctionCodeNotDeleted 检查联合主键在未删除行中是否存在
func (r *nmspacesFunctionRepository) ExistsByNamespaceAndMenuCodeAndFunctionCodeNotDeleted(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesFunction{}).
		Where("namespace_uid = ? AND menu_code = ? AND function_code = ? AND deleted_at IS NULL", namespaceUID, menuCode, functionCode).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByNamespaceAndMenuCodeAndFunctionName 检查同一 menuCode 下 functionName 是否已存在（未删除行）
func (r *nmspacesFunctionRepository) ExistsByNamespaceAndMenuCodeAndFunctionName(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionName string) (bool, error) {
	db := getDBWithTransaction(ctx, r.db)
	var count int64
	if err := db.Model(&do.NmspacesFunction{}).
		Where("namespace_uid = ? AND menu_code = ? AND function_name = ? AND deleted_at IS NULL", namespaceUID, menuCode, functionName).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
