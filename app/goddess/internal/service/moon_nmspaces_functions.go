package service

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewNmspacesFunctionService 创建租户功能服务实例
func NewNmspacesFunctionService(nmspacesFunctionBiz *biz.NmspacesFunctionBiz, sysFunctionBiz *biz.SysFunction) *NmspacesFunctionService {
	return &NmspacesFunctionService{
		nmspacesFunctionBiz: nmspacesFunctionBiz,
		sysFunctionBiz:      sysFunctionBiz,
	}
}

// NmspacesFunctionService 租户功能服务
type NmspacesFunctionService struct {
	apiv1.UnimplementedNmspacesFunctionServer

	nmspacesFunctionBiz *biz.NmspacesFunctionBiz
	sysFunctionBiz      *biz.SysFunction
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesFunctionService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// CreateNmspacesFunction 批量新增租户功能
func (s *NmspacesFunctionService) CreateNmspacesFunction(ctx context.Context, req *apiv1.CreateNmspacesFunctionRequest) (*apiv1.CreateNmspacesFunctionReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	creator := contextx.GetUserUID(ctx)

	var functions []*apiv1.NmspacesFunctionItem
	for _, key := range req.GetFunctionKeys() {
		// 从系统功能管理获取功能明细
		sysFunc, err := s.sysFunctionBiz.GetSysFunction(ctx, key.MenuCode, key.FunctionCode)
		if err != nil {
			// 系统功能不存在，跳过该记录
			continue
		}

		// 组装参数并存入租户功能表
		createBo := bo.NewCreateNmspacesFunctionBoFromSysFunction(namespaceUID, creator, sysFunc)
		result, err := s.nmspacesFunctionBiz.CreateNmspacesFunction(ctx, createBo)
		if err != nil {
			// 如果联合主键已存在则跳过，其他错误返回
			if kerr := kerrors.FromError(err); kerr != nil && strings.Contains(kerr.Message, "already exists") {
				continue
			}
			return nil, err
		}
		functions = append(functions, result.ToAPIV1NmspacesFunctionItem())
	}

	return &apiv1.CreateNmspacesFunctionReply{
		Functions: functions,
	}, nil
}

// ListNmspacesFunction 分页查询租户功能列表
func (s *NmspacesFunctionService) ListNmspacesFunction(ctx context.Context, req *apiv1.ListNmspacesFunctionRequest) (*apiv1.ListNmspacesFunctionReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesFunctionBo(namespaceUID, req)
	items, total, err := s.nmspacesFunctionBiz.ListNmspacesFunction(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.NmspacesFunctionItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, item.ToAPIV1NmspacesFunctionItem())
	}
	return &apiv1.ListNmspacesFunctionReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.PageRequestBo.Page,
		PageSize: listBo.PageRequestBo.PageSize,
	}, nil
}

// GetNmspacesFunction 查询租户功能详情
func (s *NmspacesFunctionService) GetNmspacesFunction(ctx context.Context, req *apiv1.GetNmspacesFunctionRequest) (*apiv1.GetNmspacesFunctionReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := s.nmspacesFunctionBiz.GetNmspacesFunction(ctx, namespaceUID, req.MenuCode, req.FunctionCode)
	if err != nil {
		return nil, err
	}
	return &apiv1.GetNmspacesFunctionReply{
		Function: result.ToAPIV1NmspacesFunctionItem(),
	}, nil
}

// UpdateNmspacesFunction 编辑租户功能
func (s *NmspacesFunctionService) UpdateNmspacesFunction(ctx context.Context, req *apiv1.UpdateNmspacesFunctionRequest) (*apiv1.UpdateNmspacesFunctionReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	updateBo := bo.NewUpdateNmspacesFunctionBo(namespaceUID, req)
	result, err := s.nmspacesFunctionBiz.UpdateNmspacesFunction(ctx, updateBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.UpdateNmspacesFunctionReply{
		Function: result.ToAPIV1NmspacesFunctionItem(),
	}, nil
}

// DeleteNmspacesFunction 删除租户功能
func (s *NmspacesFunctionService) DeleteNmspacesFunction(ctx context.Context, req *apiv1.DeleteNmspacesFunctionRequest) (*apiv1.DeleteNmspacesFunctionReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.nmspacesFunctionBiz.DeleteNmspacesFunction(ctx, namespaceUID, req.MenuCode, req.FunctionCode); err != nil {
		return nil, err
	}
	return &apiv1.DeleteNmspacesFunctionReply{
		Success: true,
	}, nil
}
