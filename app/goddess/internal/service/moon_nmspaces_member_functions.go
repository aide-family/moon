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

// NewNmspacesMemberFunctionService 创建租户成员功能直授服务实例
func NewNmspacesMemberFunctionService(nmspacesMemberFunctionBiz *biz.NmspacesMemberFunctionBiz) *NmspacesMemberFunctionService {
	return &NmspacesMemberFunctionService{
		nmspacesMemberFunctionBiz: nmspacesMemberFunctionBiz,
	}
}

// NmspacesMemberFunctionService 租户成员功能直授服务
type NmspacesMemberFunctionService struct {
	apiv1.UnimplementedNmspacesMemberFunctionServer

	nmspacesMemberFunctionBiz *biz.NmspacesMemberFunctionBiz
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesMemberFunctionService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspacesMemberFunction 分页查询成员直授功能列表
func (s *NmspacesMemberFunctionService) ListNmspacesMemberFunction(ctx context.Context, req *apiv1.ListNmspacesMemberFunctionRequest) (*apiv1.ListNmspacesMemberFunctionReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesMemberFunctionBo(namespaceUID, req)
	items, total, err := s.nmspacesMemberFunctionBiz.ListNmspacesMemberFunction(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.NmspacesMemberFunctionItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, item.ToAPIV1NmspacesMemberFunctionItem())
	}
	return &apiv1.ListNmspacesMemberFunctionReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.PageRequestBo.Page,
		PageSize: listBo.PageRequestBo.PageSize,
	}, nil
}

// GrantNmspacesMemberFunctions 直授成员功能
func (s *NmspacesMemberFunctionService) GrantNmspacesMemberFunctions(ctx context.Context, req *apiv1.GrantNmspacesMemberFunctionsRequest) (*apiv1.GrantNmspacesMemberFunctionsReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	creator := contextx.GetUserUID(ctx)
	grantBo := bo.NewGrantNmspacesMemberFunctionsBo(namespaceUID, creator, req)
	result, err := s.nmspacesMemberFunctionBiz.GrantNmspacesMemberFunctions(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantNmspacesMemberFunctionsReply{
		GrantedCount:         result.GrantedCount,
		AutoGrantedMenuCodes: result.AutoGrantedMenus,
	}, nil
}

// RevokeNmspacesMemberFunctions 撤销成员直授功能
func (s *NmspacesMemberFunctionService) RevokeNmspacesMemberFunctions(ctx context.Context, req *apiv1.RevokeNmspacesMemberFunctionsRequest) (*apiv1.RevokeNmspacesMemberFunctionsReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	revokeBo := bo.NewRevokeNmspacesMemberFunctionsBo(namespaceUID, req)
	result, err := s.nmspacesMemberFunctionBiz.RevokeNmspacesMemberFunctions(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeNmspacesMemberFunctionsReply{
		RevokedCount:         result.RevokedCount,
		AutoRevokedMenuCodes: result.AutoRevokedMenus,
	}, nil
}
