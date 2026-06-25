package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewNmspacesRoleFunctionService 创建租户角色功能服务实例
func NewNmspacesRoleFunctionService(nmspacesRoleFunctionBiz *biz.NmspacesRoleFunctionBiz) *NmspacesRoleFunctionService {
	return &NmspacesRoleFunctionService{
		nmspacesRoleFunctionBiz: nmspacesRoleFunctionBiz,
	}
}

// NmspacesRoleFunctionService 租户角色功能服务
type NmspacesRoleFunctionService struct {
	apiv1.UnimplementedNmspacesRoleFunctionServer

	nmspacesRoleFunctionBiz *biz.NmspacesRoleFunctionBiz
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesRoleFunctionService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspacesRoleFunction 分页查询角色的功能列表
func (s *NmspacesRoleFunctionService) ListNmspacesRoleFunction(ctx context.Context, req *apiv1.ListNmspacesRoleFunctionRequest) (*apiv1.ListNmspacesRoleFunctionReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesRoleFunctionBo(namespaceUID, req)
	items, total, err := s.nmspacesRoleFunctionBiz.ListNmspacesRoleFunction(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.NmspacesRoleFunctionItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, item.ToAPIV1NmspacesRoleFunctionItem())
	}
	return &apiv1.ListNmspacesRoleFunctionReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.PageRequestBo.Page,
		PageSize: listBo.PageRequestBo.PageSize,
	}, nil
}

// GrantNmspacesRoleFunctions 授予角色功能
func (s *NmspacesRoleFunctionService) GrantNmspacesRoleFunctions(ctx context.Context, req *apiv1.GrantNmspacesRoleFunctionsRequest) (*apiv1.GrantNmspacesRoleFunctionsReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	creator := contextx.GetUserUID(ctx)
	grantBo := bo.NewGrantNmspacesRoleFunctionsBo(namespaceUID, creator, req)
	result, err := s.nmspacesRoleFunctionBiz.GrantNmspacesRoleFunctions(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantNmspacesRoleFunctionsReply{
		GrantedCount:         result.GrantedCount,
		AutoGrantedMenuCodes: result.AutoGrantedMenus,
	}, nil
}

// RevokeNmspacesRoleFunctions 撤销角色功能
func (s *NmspacesRoleFunctionService) RevokeNmspacesRoleFunctions(ctx context.Context, req *apiv1.RevokeNmspacesRoleFunctionsRequest) (*apiv1.RevokeNmspacesRoleFunctionsReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	revokeBo := bo.NewRevokeNmspacesRoleFunctionsBo(namespaceUID, req)
	result, err := s.nmspacesRoleFunctionBiz.RevokeNmspacesRoleFunctions(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeNmspacesRoleFunctionsReply{
		RevokedCount:         result.RevokedCount,
		AutoRevokedMenuCodes: result.AutoRevokedMenus,
	}, nil
}
