package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewNmspacesFunction 创建租户功能业务逻辑实例
func NewNmspacesFunction(
	nmspacesFunctionRepo repository.NmspacesFunction,
	nmspacesMenuRepo repository.NmspacesMenu,
	helper *klog.Helper,
) *NmspacesFunctionBiz {
	return &NmspacesFunctionBiz{
		nmspacesFunctionRepo: nmspacesFunctionRepo,
		nmspacesMenuRepo:     nmspacesMenuRepo,
		helper:               klog.NewHelper(klog.With(helper.Logger(), "biz", "nmspaces_function")),
	}
}

// NmspacesFunctionBiz 租户功能业务逻辑
type NmspacesFunctionBiz struct {
	helper               *klog.Helper
	nmspacesFunctionRepo repository.NmspacesFunction
	nmspacesMenuRepo     repository.NmspacesMenu
}

// CreateNmspacesFunction 新增租户功能
func (s *NmspacesFunctionBiz) CreateNmspacesFunction(ctx context.Context, req *bo.CreateNmspacesFunctionBo) (*bo.NmspacesFunctionBo, error) {
	// 校验 menuCode 在 moon_nmspaces_menus 中存在且未删除
	menuExists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndMenuCodeNotDeleted(ctx, req.NamespaceUID, req.MenuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menuCode existence failed", "error", err, "namespaceUID", req.NamespaceUID, "menuCode", req.MenuCode)
		return nil, merr.ErrorInternalServer("check menuCode failed").WithCause(err)
	}
	if !menuExists {
		return nil, merr.ErrorNotFound("menu %s not found in namespace", req.MenuCode)
	}

	// 校验联合主键唯一（含软删除行）
	exists, err := s.nmspacesFunctionRepo.ExistsByNamespaceAndMenuCodeAndFunctionCode(ctx, req.NamespaceUID, req.MenuCode, req.FunctionCode)
	if err != nil {
		s.helper.Errorw("msg", "check functionCode uniqueness failed", "error", err)
		return nil, merr.ErrorInternalServer("check functionCode failed").WithCause(err)
	}
	if exists {
		return nil, merr.ErrorBadRequest("function %s already exists under menu %s", req.FunctionCode, req.MenuCode)
	}

	// 校验同一 menuCode 下 functionName 唯一
	nameExists, err := s.nmspacesFunctionRepo.ExistsByNamespaceAndMenuCodeAndFunctionName(ctx, req.NamespaceUID, req.MenuCode, req.FunctionName)
	if err != nil {
		s.helper.Errorw("msg", "check functionName uniqueness failed", "error", err)
		return nil, merr.ErrorInternalServer("check functionName failed").WithCause(err)
	}
	if nameExists {
		return nil, merr.ErrorBadRequest("functionName %s already exists under menu %s", req.FunctionName, req.MenuCode)
	}

	result, err := s.nmspacesFunctionRepo.CreateNmspacesFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "create nmspaces function failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("create nmspaces function failed").WithCause(err)
	}
	return result, nil
}

// ListNmspacesFunction 分页查询租户功能列表
func (s *NmspacesFunctionBiz) ListNmspacesFunction(ctx context.Context, req *bo.ListNmspacesFunctionBo) ([]*bo.NmspacesFunctionBo, int64, error) {
	items, total, err := s.nmspacesFunctionRepo.ListNmspacesFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list nmspaces function failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list nmspaces function failed").WithCause(err)
	}
	return items, total, nil
}

// GetNmspacesFunction 根据联合主键查询功能详情
func (s *NmspacesFunctionBiz) GetNmspacesFunction(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) (*bo.NmspacesFunctionBo, error) {
	result, err := s.nmspacesFunctionRepo.GetNmspacesFunction(ctx, namespaceUID, menuCode, functionCode)
	if err != nil {
		s.helper.Errorw("msg", "get nmspaces function failed", "error", err)
		return nil, merr.ErrorInternalServer("get nmspaces function failed").WithCause(err)
	}
	if result == nil {
		return nil, merr.ErrorNotFound("function %s under menu %s not found", functionCode, menuCode)
	}
	return result, nil
}

// UpdateNmspacesFunction 编辑租户功能
func (s *NmspacesFunctionBiz) UpdateNmspacesFunction(ctx context.Context, req *bo.UpdateNmspacesFunctionBo) (*bo.NmspacesFunctionBo, error) {
	existing, err := s.nmspacesFunctionRepo.GetNmspacesFunction(ctx, req.NamespaceUID, req.MenuCode, req.FunctionCode)
	if err != nil {
		s.helper.Errorw("msg", "check function existence failed", "error", err)
		return nil, merr.ErrorInternalServer("check function existence failed").WithCause(err)
	}
	if existing == nil {
		return nil, merr.ErrorNotFound("function %s under menu %s not found", req.FunctionCode, req.MenuCode)
	}

	if req.FunctionName != nil {
		nameExists, err := s.nmspacesFunctionRepo.ExistsByNamespaceAndMenuCodeAndFunctionName(ctx, req.NamespaceUID, req.MenuCode, *req.FunctionName)
		if err != nil {
			s.helper.Errorw("msg", "check functionName uniqueness failed", "error", err)
			return nil, merr.ErrorInternalServer("check functionName failed").WithCause(err)
		}
		if nameExists && *req.FunctionName != existing.FunctionName {
			return nil, merr.ErrorBadRequest("functionName %s already exists under menu %s", *req.FunctionName, req.MenuCode)
		}
	}

	result, err := s.nmspacesFunctionRepo.UpdateNmspacesFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "update nmspaces function failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("update nmspaces function failed").WithCause(err)
	}
	return result, nil
}

// DeleteNmspacesFunction 删除租户功能（软删除）
func (s *NmspacesFunctionBiz) DeleteNmspacesFunction(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) error {
	existing, err := s.nmspacesFunctionRepo.GetNmspacesFunction(ctx, namespaceUID, menuCode, functionCode)
	if err != nil {
		s.helper.Errorw("msg", "check function existence failed", "error", err)
		return merr.ErrorInternalServer("check function existence failed").WithCause(err)
	}
	if existing == nil {
		return merr.ErrorNotFound("function %s under menu %s not found", functionCode, menuCode)
	}

	if err := s.nmspacesFunctionRepo.DeleteNmspacesFunction(ctx, namespaceUID, menuCode, functionCode); err != nil {
		s.helper.Errorw("msg", "delete nmspaces function failed", "error", err)
		return merr.ErrorInternalServer("delete nmspaces function failed").WithCause(err)
	}
	return nil
}
