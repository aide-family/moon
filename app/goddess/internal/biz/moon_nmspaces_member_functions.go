package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewNmspacesMemberFunction 创建租户成员功能直授业务逻辑实例
func NewNmspacesMemberFunction(
	nmspacesMemberFunctionRepo repository.NmspacesMemberFunction,
	nmspacesFunctionRepo repository.NmspacesFunction,
	nmspacesMemberMenuRepo repository.NmspacesMemberMenu,
	memberRepo repository.Member,
	helper *klog.Helper,
) *NmspacesMemberFunctionBiz {
	return &NmspacesMemberFunctionBiz{
		nmspacesMemberFunctionRepo: nmspacesMemberFunctionRepo,
		nmspacesFunctionRepo:       nmspacesFunctionRepo,
		nmspacesMemberMenuRepo:     nmspacesMemberMenuRepo,
		memberRepo:                 memberRepo,
		helper:                     klog.NewHelper(klog.With(helper.Logger(), "biz", "nmspaces_member_function")),
	}
}

// NmspacesMemberFunctionBiz 租户成员功能直授业务逻辑
type NmspacesMemberFunctionBiz struct {
	helper                     *klog.Helper
	nmspacesMemberFunctionRepo repository.NmspacesMemberFunction
	nmspacesFunctionRepo       repository.NmspacesFunction
	nmspacesMemberMenuRepo     repository.NmspacesMemberMenu
	memberRepo                 repository.Member
}

// ListNmspacesMemberFunction 查询成员直授功能列表
func (s *NmspacesMemberFunctionBiz) ListNmspacesMemberFunction(ctx context.Context, req *bo.ListNmspacesMemberFunctionBo) ([]*bo.NmspacesMemberFunctionItemBo, int64, error) {
	items, total, err := s.nmspacesMemberFunctionRepo.ListNmspacesMemberFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list nmspaces member function failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list nmspaces member function failed").WithCause(err)
	}
	return items, total, nil
}

// GrantNmspacesMemberFunctions 直授成员功能
func (s *NmspacesMemberFunctionBiz) GrantNmspacesMemberFunctions(ctx context.Context, req *bo.GrantNmspacesMemberFunctionsBo) (*bo.GrantFunctionResultBo, error) {
	// 1. 校验 memberUID 存在
	member, err := s.memberRepo.GetMember(ctx, req.MemberUID)
	if err != nil {
		s.helper.Errorw("msg", "check member existence failed", "error", err, "memberUID", req.MemberUID)
		return nil, merr.ErrorInternalServer("check member %d failed", req.MemberUID).WithCause(err)
	}
	if member == nil {
		return nil, merr.ErrorNotFound("member %d not found", req.MemberUID)
	}

	// 2. 校验每个 functionKey 在当前租户存在且未删除
	for _, fk := range req.FunctionKeys {
		exists, err := s.nmspacesFunctionRepo.ExistsByNamespaceAndMenuCodeAndFunctionCodeNotDeleted(ctx, req.NamespaceUID, fk.MenuCode, fk.FunctionCode)
		if err != nil {
			s.helper.Errorw("msg", "check function existence failed", "error", err)
			return nil, merr.ErrorInternalServer("check function %s/%s failed", fk.MenuCode, fk.FunctionCode).WithCause(err)
		}
		if !exists {
			return nil, merr.ErrorNotFound("function %s/%s not found", fk.MenuCode, fk.FunctionCode)
		}
	}

	// 3. 批量插入
	grantedCount, err := s.nmspacesMemberFunctionRepo.GrantNmspacesMemberFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "grant nmspaces member functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("grant nmspaces member functions failed").WithCause(err)
	}

	// 4. autoGrantMenu: 自动追加对应菜单直授
	var autoGrantedMenus []string
	if req.AutoGrantMenu {
		menuCodeSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuCodeSet[fk.MenuCode] = struct{}{}
		}

		var toGrant []string
		for menuCode := range menuCodeSet {
			exists, err := s.nmspacesMemberMenuRepo.ExistsByNamespaceAndMemberUIDAndMenuCode(ctx, req.NamespaceUID, req.MemberUID, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "check member menu existence failed", "error", err)
				return nil, merr.ErrorInternalServer("check member menu failed").WithCause(err)
			}
			if !exists {
				toGrant = append(toGrant, menuCode)
			}
		}

		if len(toGrant) > 0 {
			grantReq := &bo.GrantNmspacesMemberMenusBo{
				NamespaceUID: req.NamespaceUID,
				MemberUID:    req.MemberUID,
				MenuCodes:    toGrant,
				Creator:      req.Creator,
			}
			_, err := s.nmspacesMemberMenuRepo.GrantNmspacesMemberMenus(ctx, grantReq)
			if err != nil {
				s.helper.Errorw("msg", "auto grant member menus failed", "error", err)
				return nil, merr.ErrorInternalServer("auto grant member menus failed").WithCause(err)
			}
			autoGrantedMenus = toGrant
		}
	}

	return &bo.GrantFunctionResultBo{
		GrantedCount:     grantedCount,
		AutoGrantedMenus: autoGrantedMenus,
	}, nil
}

// RevokeNmspacesMemberFunctions 撤销成员直授功能
func (s *NmspacesMemberFunctionBiz) RevokeNmspacesMemberFunctions(ctx context.Context, req *bo.RevokeNmspacesMemberFunctionsBo) (*bo.RevokeFunctionResultBo, error) {
	// 1. 校验 memberUID 存在
	member, err := s.memberRepo.GetMember(ctx, req.MemberUID)
	if err != nil {
		s.helper.Errorw("msg", "check member existence failed", "error", err, "memberUID", req.MemberUID)
		return nil, merr.ErrorInternalServer("check member %d failed", req.MemberUID).WithCause(err)
	}
	if member == nil {
		return nil, merr.ErrorNotFound("member %d not found", req.MemberUID)
	}

	// 2. 硬删除
	revokedCount, err := s.nmspacesMemberFunctionRepo.RevokeNmspacesMemberFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "revoke nmspaces member functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("revoke nmspaces member functions failed").WithCause(err)
	}

	// 3. autoRevokeMenu: 检查被撤销功能对应的菜单是否还有其他功能
	var autoRevokedMenus []string
	if req.AutoRevokeMenu {
		menuCodeSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuCodeSet[fk.MenuCode] = struct{}{}
		}

		for menuCode := range menuCodeSet {
			remaining, err := s.nmspacesMemberFunctionRepo.CountByNamespaceAndMemberUIDAndMenuCode(ctx, req.NamespaceUID, req.MemberUID, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "count remaining functions failed", "error", err)
				return nil, merr.ErrorInternalServer("count remaining functions failed").WithCause(err)
			}
			if remaining == 0 {
				revokeMenuReq := &bo.RevokeNmspacesMemberMenusBo{
					NamespaceUID: req.NamespaceUID,
					MemberUID:    req.MemberUID,
					MenuCodes:    []string{menuCode},
				}
				_, err := s.nmspacesMemberMenuRepo.RevokeNmspacesMemberMenus(ctx, revokeMenuReq)
				if err != nil {
					s.helper.Errorw("msg", "auto revoke member menu failed", "error", err)
					return nil, merr.ErrorInternalServer("auto revoke member menu failed").WithCause(err)
				}
				autoRevokedMenus = append(autoRevokedMenus, menuCode)
			}
		}
	}

	return &bo.RevokeFunctionResultBo{
		RevokedCount:     revokedCount,
		AutoRevokedMenus: autoRevokedMenus,
	}, nil
}
