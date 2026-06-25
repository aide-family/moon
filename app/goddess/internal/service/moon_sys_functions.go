package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysFunctionService 创建系统功能服务实例
func NewSysFunctionService(sysFunctionBiz *biz.SysFunction) *SysFunctionService {
	return &SysFunctionService{
		sysFunctionBiz: sysFunctionBiz,
	}
}

// SysFunctionService 系统功能服务，实现 proto 定义的 SysFunctionServer 接口
type SysFunctionService struct {
	apiv1.UnimplementedSysFunctionServer

	sysFunctionBiz *biz.SysFunction
}

// CreateSysFunction 新增系统功能
func (s *SysFunctionService) CreateSysFunction(ctx context.Context, req *apiv1.CreateSysFunctionRequest) (*apiv1.CreateSysFunctionReply, error) {
	createBo := bo.NewCreateSysFunctionBo(req)
	createBo.Creator = int64(contextx.GetUserUID(ctx))
	result, err := s.sysFunctionBiz.CreateSysFunction(ctx, createBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateSysFunctionReply{
		Function: result.ToAPIV1SysFunctionItem(),
	}, nil
}

// ListSysFunction 分页查询系统功能列表
func (s *SysFunctionService) ListSysFunction(ctx context.Context, req *apiv1.ListSysFunctionRequest) (*apiv1.ListSysFunctionReply, error) {
	listBo := bo.NewListSysFunctionBo(req)
	items, total, err := s.sysFunctionBiz.ListSysFunction(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.SysFunctionItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, item.ToAPIV1SysFunctionItem())
	}
	return &apiv1.ListSysFunctionReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.PageRequestBo.Page,
		PageSize: listBo.PageRequestBo.PageSize,
	}, nil
}

// GetSysFunction 查询系统功能详情
func (s *SysFunctionService) GetSysFunction(ctx context.Context, req *apiv1.GetSysFunctionRequest) (*apiv1.GetSysFunctionReply, error) {
	result, err := s.sysFunctionBiz.GetSysFunction(ctx, req.MenuCode, req.FunctionCode)
	if err != nil {
		return nil, err
	}
	return &apiv1.GetSysFunctionReply{
		Function: result.ToAPIV1SysFunctionItem(),
	}, nil
}

// UpdateSysFunction 编辑系统功能
func (s *SysFunctionService) UpdateSysFunction(ctx context.Context, req *apiv1.UpdateSysFunctionRequest) (*apiv1.UpdateSysFunctionReply, error) {
	updateBo := bo.NewUpdateSysFunctionBo(req)
	result, err := s.sysFunctionBiz.UpdateSysFunction(ctx, updateBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.UpdateSysFunctionReply{
		Function: result.ToAPIV1SysFunctionItem(),
	}, nil
}

// DeleteSysFunction 删除系统功能
func (s *SysFunctionService) DeleteSysFunction(ctx context.Context, req *apiv1.DeleteSysFunctionRequest) (*apiv1.DeleteSysFunctionReply, error) {
	if err := s.sysFunctionBiz.DeleteSysFunction(ctx, req.MenuCode, req.FunctionCode); err != nil {
		return nil, err
	}
	return &apiv1.DeleteSysFunctionReply{
		Success: true,
	}, nil
}
