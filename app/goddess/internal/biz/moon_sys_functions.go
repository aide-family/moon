package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysFunction 创建系统功能业务逻辑实例
func NewSysFunction(
	sysFunctionRepo repository.SysFunction,
	sysMenuRepo repository.SysMenu,
	helper *klog.Helper,
) *SysFunction {
	return &SysFunction{
		sysFunctionRepo: sysFunctionRepo,
		sysMenuRepo:     sysMenuRepo,
		helper:          klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_function")),
	}
}

// SysFunction 系统功能业务逻辑
type SysFunction struct {
	helper          *klog.Helper
	sysFunctionRepo repository.SysFunction
	sysMenuRepo     repository.SysMenu
}

// CreateSysFunction 新增系统功能
func (s *SysFunction) CreateSysFunction(ctx context.Context, req *bo.CreateSysFunctionBo) (*bo.SysFunctionBo, error) {
	// 校验 menuCode 在 moon_sys_menus 中存在且未删除
	menuExists, err := s.sysMenuRepo.ExistsByMenuCodeNotDeleted(ctx, req.MenuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menuCode existence failed", "error", err, "menuCode", req.MenuCode)
		return nil, merr.ErrorInternalServer("check menuCode failed").WithCause(err)
	}
	if !menuExists {
		return nil, merr.ErrorNotFound("menu %s not found", req.MenuCode)
	}

	// 校验联合主键 (menuCode, functionCode) 唯一（含软删除行）
	exists, err := s.sysFunctionRepo.ExistsByMenuCodeAndFunctionCode(ctx, req.MenuCode, req.FunctionCode)
	if err != nil {
		s.helper.Errorw("msg", "check functionCode uniqueness failed", "error", err, "menuCode", req.MenuCode, "functionCode", req.FunctionCode)
		return nil, merr.ErrorInternalServer("check functionCode failed").WithCause(err)
	}
	if exists {
		return nil, merr.ErrorBadRequest("function %s already exists under menu %s", req.FunctionCode, req.MenuCode)
	}

	// 校验同一 menuCode 下 functionName 在未删除行中唯一
	nameExists, err := s.sysFunctionRepo.ExistsByMenuCodeAndFunctionName(ctx, req.MenuCode, req.FunctionName)
	if err != nil {
		s.helper.Errorw("msg", "check functionName uniqueness failed", "error", err, "menuCode", req.MenuCode, "functionName", req.FunctionName)
		return nil, merr.ErrorInternalServer("check functionName failed").WithCause(err)
	}
	if nameExists {
		return nil, merr.ErrorBadRequest("functionName %s already exists under menu %s", req.FunctionName, req.MenuCode)
	}

	// 插入功能
	result, err := s.sysFunctionRepo.CreateSysFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "create sys function failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("create sys function failed").WithCause(err)
	}
	return result, nil
}

// ListSysFunction 分页查询系统功能列表
func (s *SysFunction) ListSysFunction(ctx context.Context, req *bo.ListSysFunctionBo) ([]*bo.SysFunctionBo, int64, error) {
	items, total, err := s.sysFunctionRepo.ListSysFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list sys function failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list sys function failed").WithCause(err)
	}
	return items, total, nil
}

// GetSysFunction 根据联合主键查询功能详情
func (s *SysFunction) GetSysFunction(ctx context.Context, menuCode, functionCode string) (*bo.SysFunctionBo, error) {
	result, err := s.sysFunctionRepo.GetSysFunction(ctx, menuCode, functionCode)
	if err != nil {
		s.helper.Errorw("msg", "get sys function failed", "error", err, "menuCode", menuCode, "functionCode", functionCode)
		return nil, merr.ErrorInternalServer("get sys function failed").WithCause(err)
	}
	if result == nil {
		return nil, merr.ErrorNotFound("function %s under menu %s not found", functionCode, menuCode)
	}
	return result, nil
}

// UpdateSysFunction 编辑系统功能
func (s *SysFunction) UpdateSysFunction(ctx context.Context, req *bo.UpdateSysFunctionBo) (*bo.SysFunctionBo, error) {
	// 校验功能存在且未删除
	existing, err := s.sysFunctionRepo.GetSysFunction(ctx, req.MenuCode, req.FunctionCode)
	if err != nil {
		s.helper.Errorw("msg", "check function existence failed", "error", err, "menuCode", req.MenuCode, "functionCode", req.FunctionCode)
		return nil, merr.ErrorInternalServer("check function existence failed").WithCause(err)
	}
	if existing == nil {
		return nil, merr.ErrorNotFound("function %s under menu %s not found", req.FunctionCode, req.MenuCode)
	}

	// 若修改 functionName，校验同一 menuCode 下唯一
	if req.FunctionName != nil {
		nameExists, err := s.sysFunctionRepo.ExistsByMenuCodeAndFunctionName(ctx, req.MenuCode, *req.FunctionName)
		if err != nil {
			s.helper.Errorw("msg", "check functionName uniqueness failed", "error", err, "menuCode", req.MenuCode, "functionName", *req.FunctionName)
			return nil, merr.ErrorInternalServer("check functionName failed").WithCause(err)
		}
		if nameExists && *req.FunctionName != existing.FunctionName {
			return nil, merr.ErrorBadRequest("functionName %s already exists under menu %s", *req.FunctionName, req.MenuCode)
		}
	}

	// 执行更新
	result, err := s.sysFunctionRepo.UpdateSysFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "update sys function failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("update sys function failed").WithCause(err)
	}
	return result, nil
}

// DeleteSysFunction 删除系统功能（软删除）
func (s *SysFunction) DeleteSysFunction(ctx context.Context, menuCode, functionCode string) error {
	// 校验功能存在且未删除
	existing, err := s.sysFunctionRepo.GetSysFunction(ctx, menuCode, functionCode)
	if err != nil {
		s.helper.Errorw("msg", "check function existence failed", "error", err, "menuCode", menuCode, "functionCode", functionCode)
		return merr.ErrorInternalServer("check function existence failed").WithCause(err)
	}
	if existing == nil {
		return merr.ErrorNotFound("function %s under menu %s not found", functionCode, menuCode)
	}

	// 执行软删除
	if err := s.sysFunctionRepo.DeleteSysFunction(ctx, menuCode, functionCode); err != nil {
		s.helper.Errorw("msg", "delete sys function failed", "error", err, "menuCode", menuCode, "functionCode", functionCode)
		return merr.ErrorInternalServer("delete sys function failed").WithCause(err)
	}
	return nil
}
