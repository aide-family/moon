package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysUserFunctionService 创建系统用户功能直授服务实例
func NewSysUserFunctionService(sysUserFunctionBiz *biz.SysUserFunction) *SysUserFunctionService {
	return &SysUserFunctionService{
		sysUserFunctionBiz: sysUserFunctionBiz,
	}
}

// SysUserFunctionService 系统用户功能直授服务
type SysUserFunctionService struct {
	apiv1.UnimplementedSysUserFunctionServer

	sysUserFunctionBiz *biz.SysUserFunction
}

// ListUserFunction 分页查询用户直授功能列表
func (s *SysUserFunctionService) ListUserFunction(ctx context.Context, req *apiv1.ListUserFunctionRequest) (*apiv1.ListUserFunctionReply, error) {
	listBo := bo.NewListUserFunctionBo(req)
	items, total, err := s.sysUserFunctionBiz.ListUserFunction(ctx, listBo)
	if err != nil {
		return nil, err
	}

	apiItems := make([]*apiv1.UserFunctionItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, convert.UserFunctionItemBoToAPIV1(item))
	}
	return &apiv1.ListUserFunctionReply{
		Items:    apiItems,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GrantUserFunctions 直授系统功能给用户
func (s *SysUserFunctionService) GrantUserFunctions(ctx context.Context, req *apiv1.GrantUserFunctionsRequest) (*apiv1.GrantUserFunctionsReply, error) {
	grantBo := bo.NewGrantUserFunctionsBo(req)
	grantBo.Creator = contextx.GetUserUID(ctx)
	result, err := s.sysUserFunctionBiz.GrantUserFunctions(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantUserFunctionsReply{
		GrantedCount:         result.GrantedCount,
		AutoGrantedMenuCodes: result.AutoGrantedMenus,
	}, nil
}

// RevokeUserFunctions 撤销直授系统功能
func (s *SysUserFunctionService) RevokeUserFunctions(ctx context.Context, req *apiv1.RevokeUserFunctionsRequest) (*apiv1.RevokeUserFunctionsReply, error) {
	revokeBo := bo.NewRevokeUserFunctionsBo(req)
	result, err := s.sysUserFunctionBiz.RevokeUserFunctions(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeUserFunctionsReply{
		RevokedCount:         result.RevokedCount,
		AutoRevokedMenuCodes: result.AutoRevokedMenus,
	}, nil
}
