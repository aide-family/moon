package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysRoleFunctionService 创建系统角色功能服务实例
func NewSysRoleFunctionService(sysRoleFunctionBiz *biz.SysRoleFunction) *SysRoleFunctionService {
	return &SysRoleFunctionService{
		sysRoleFunctionBiz: sysRoleFunctionBiz,
	}
}

// SysRoleFunctionService 系统角色功能服务，实现 proto 定义的 SysRoleFunctionServer 接口
type SysRoleFunctionService struct {
	apiv1.UnimplementedSysRoleFunctionServer

	sysRoleFunctionBiz *biz.SysRoleFunction
}

// ListRoleFunction 分页查询系统角色拥有的功能列表
func (s *SysRoleFunctionService) ListRoleFunction(ctx context.Context, req *apiv1.ListRoleFunctionRequest) (*apiv1.ListRoleFunctionReply, error) {
	listBo := bo.NewListRoleFunctionBo(req)
	items, total, err := s.sysRoleFunctionBiz.ListRoleFunction(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.RoleFunctionItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, convert.RoleFunctionItemBoToAPIV1(item))
	}
	return &apiv1.ListRoleFunctionReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.PageRequestBo.Page,
		PageSize: listBo.PageRequestBo.PageSize,
	}, nil
}

// GrantRoleFunctions 授予系统角色功能
func (s *SysRoleFunctionService) GrantRoleFunctions(ctx context.Context, req *apiv1.GrantRoleFunctionsRequest) (*apiv1.GrantRoleFunctionsReply, error) {
	grantBo := bo.NewGrantRoleFunctionsBo(req)
	grantBo.Creator = int64(contextx.GetUserUID(ctx))
	result, err := s.sysRoleFunctionBiz.GrantRoleFunctions(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantRoleFunctionsReply{
		GrantedCount:         result.GrantedCount,
		AutoGrantedMenuCodes: result.AutoGrantedMenus,
	}, nil
}

// RevokeRoleFunctions 撤销系统角色功能
func (s *SysRoleFunctionService) RevokeRoleFunctions(ctx context.Context, req *apiv1.RevokeRoleFunctionsRequest) (*apiv1.RevokeRoleFunctionsReply, error) {
	revokeBo := bo.NewRevokeRoleFunctionsBo(req)
	result, err := s.sysRoleFunctionBiz.RevokeRoleFunctions(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeRoleFunctionsReply{
		RevokedCount:         result.RevokedCount,
		AutoRevokedMenuCodes: result.AutoRevokedMenus,
	}, nil
}
